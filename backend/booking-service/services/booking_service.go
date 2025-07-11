package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/booking-service/models"
	"backend/booking-service/repositories"
	"backend/booking-service/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingService struct {
	BookingRepo          *repositories.BookingRepository
	ConcertRepo          *repositories.ConcertRepository
	SeatRepo             *repositories.SeatRepository
	TicketClassRepo      *repositories.TicketClassRepository
	BuyerRepo            *repositories.BuyerRepository
	TicketHolderRepo     *repositories.TicketHolderRepository
	PaymentServiceAPIURL string
}

func NewBookingService(bRepo *repositories.BookingRepository, cRepo *repositories.ConcertRepository, sRepo *repositories.SeatRepository, tcRepo *repositories.TicketClassRepository, buyerRepo *repositories.BuyerRepository, ticketHolderRepo *repositories.TicketHolderRepository, paymentServiceAPIURL string) *BookingService {
	return &BookingService{
		BookingRepo:          bRepo,
		ConcertRepo:          cRepo,
		SeatRepo:             sRepo,
		TicketClassRepo:      tcRepo,
		BuyerRepo:            buyerRepo,
		TicketHolderRepo:     ticketHolderRepo,
		PaymentServiceAPIURL: paymentServiceAPIURL,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID uint, req *models.CreateBookingRequest) (*models.BookingResponse, error) {

	totalRequestedTickets := 0
	for _, tc := range req.TicketsByClass {
		totalRequestedTickets += tc.Quantity
	}
	if totalRequestedTickets == 0 || totalRequestedTickets > 5 {
		return nil, fmt.Errorf("invalid total number of tickets requested: %d (must be between 1 and 5)", totalRequestedTickets)
	}

	activeBookings, err := s.BookingRepo.GetUserActiveBookingsForConcert(userID, req.ConcertID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.LogError("DB error checking active bookings for user %d, concert %d: %v", userID, req.ConcertID, err)
		return nil, errors.New("failed to check existing bookings")
	}
	if len(activeBookings) > 0 {
		return nil, errors.New("you already have an active (pending or confirmed) booking for this concert. Please cancel your existing booking to proceed")
	}

	concert, err := s.ConcertRepo.GetConcertByID(req.ConcertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("concert not found")
		}
		utils.LogError("DB error getting concert %d for booking: %v", req.ConcertID, err)
		return nil, errors.New("failed to get concert details")
	}

	if concert.Status != models.ConcertStatusActive {
		return nil, fmt.Errorf("concert '%s' is not active for booking (status: %s)", concert.Name, concert.Status)
	}

	concertTicketClassesMap := make(map[uint]models.TicketClass)
	for _, tc := range concert.TicketClasses {
		concertTicketClassesMap[tc.ID] = tc
	}

	var seatsToBook []models.Seat
	var totalPrice float64 = 0.0

	for _, tcRequest := range req.TicketsByClass {
		ticketClass, exists := concertTicketClassesMap[tcRequest.TicketClassID]
		if !exists {
			return nil, fmt.Errorf("ticket class ID %d not found for concert %d", tcRequest.TicketClassID, req.ConcertID)
		}

		if tcRequest.Quantity <= 0 {
			continue
		}

		_, err = utils.DecreaseAvailableSeatsAtomically(ctx, ticketClass.ID, tcRequest.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to reserve tickets for class '%s': %v", ticketClass.Name, err)
		}

		for i := 0; i < tcRequest.Quantity; i++ {
			seatsToBook = append(seatsToBook, models.Seat{
				ConcertID:     concert.ID,
				TicketClassID: ticketClass.ID,
				SeatNumber:    fmt.Sprintf("%s-%d-%s", ticketClass.Name, time.Now().UnixNano()/1000000, uuid.New().String()[:6]),
				Status:        models.SeatStatusReserved,
				UserID:        &userID,
			})
		}
		totalPrice += ticketClass.Price * float64(tcRequest.Quantity)
	}

	newUUID := uuid.New().String()
	booking := &models.Booking{
		ID:         newUUID,
		UserID:     userID,
		ConcertID:  concert.ID,
		TotalPrice: totalPrice,
		Status:     models.BookingStatusPending,
		ExpiresAt:  func() *time.Time { t := time.Now().Add(15 * time.Minute); return &t }(),
	}

	tx := s.BookingRepo.DB.Begin()
	if tx.Error != nil {
		utils.LogError("Failed to begin DB transaction for booking: %v", tx.Error)
		for _, tcRequest := range req.TicketsByClass {
			utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
		}
		return nil, errors.New("failed to initiate booking transaction")
	}

	tempSeatRepo := &repositories.SeatRepository{DB: tx}
	tempBookingRepo := &repositories.BookingRepository{DB: tx}
	tempTicketClassRepo := &repositories.TicketClassRepository{DB: tx}
	tempBuyerRepo := &repositories.BuyerRepository{DB: tx}
	tempTicketHolderRepo := &repositories.TicketHolderRepository{DB: tx}

	if err := tempSeatRepo.CreateSeats(tx, seatsToBook); err != nil {
		tx.Rollback()
		for _, tcRequest := range req.TicketsByClass {
			utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
		}
		utils.LogError("Failed to create new seats for booking: %v", err)
		return nil, errors.New("failed to create seats for booking")
	}

	seatIDs := make([]string, len(seatsToBook))
	for i, seat := range seatsToBook {
		seatIDs[i] = fmt.Sprintf("%d", seat.ID)
	}
	booking.SeatIDs = strings.Join(seatIDs, ",")

	if err := tempBookingRepo.CreateBooking(booking); err != nil {
		tx.Rollback()
		for _, tcRequest := range req.TicketsByClass {
			utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
		}
		utils.LogError("Failed to create booking record in DB: %v", err)
		return nil, errors.New("failed to create booking record")
	}

	buyer := models.Buyer{
		BookingID:   booking.ID,
		FullName:    req.BuyerInfo.FullName,
		PhoneNumber: req.BuyerInfo.PhoneNumber,
		Email:       req.BuyerInfo.Email,
		KTPNumber:   req.BuyerInfo.KTPNumber,
	}
	if err := tempBuyerRepo.CreateBuyer(tx, &buyer); err != nil {
		tx.Rollback()
		for _, tcRequest := range req.TicketsByClass {
			utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
		}
		utils.LogError("Failed to create buyer info for booking %s: %v", booking.ID, err)
		return nil, errors.New("failed to save buyer information")
	}

	var ticketHolder models.TicketHolder
	if req.TicketHolderInfo != nil {
		ticketHolder = models.TicketHolder{
			BookingID: booking.ID,
			FullName:  req.TicketHolderInfo.FullName,
			KTPNumber: req.TicketHolderInfo.KTPNumber,
		}
		if err := tempTicketHolderRepo.CreateTicketHolder(tx, &ticketHolder); err != nil {
			tx.Rollback()
			for _, tcRequest := range req.TicketsByClass {
				utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
			}
			utils.LogError("Failed to create ticket holder info for booking %s: %v", booking.ID, err)
			return nil, errors.New("failed to save ticket holder information")
		}
	}

	tcQuantities := make(map[uint]int)
	for _, item := range req.TicketsByClass {
		tcQuantities[item.TicketClassID] = item.Quantity
	}

	for tcID, qty := range tcQuantities {
		ticketClass, exists := concertTicketClassesMap[tcID]
		if !exists {
			utils.LogError("TicketClass %d not found in concert for update", tcID)
			continue
		}
		ticketClass.AvailableSeatsInClass -= qty

		if err := tempTicketClassRepo.UpdateTicketClass(tx, &ticketClass); err != nil {
			tx.Rollback()
			for _, tcRequest := range req.TicketsByClass {
				utils.IncreaseAvailableSeatsAtomically(ctx, tcRequest.TicketClassID, tcRequest.Quantity)
			}
			utils.LogError("Failed to update available seats for ticket class %d: %v", tcID, err)
			return nil, errors.New("failed to update ticket class availability")
		}
	}

	tx.Commit()

	go func() {
		paymentReq := struct {
			BookingID     string  `json:"booking_id"`
			Amount        float64 `json:"amount"`
			PaymentMethod string  `json:"payment_method"`
		}{
			BookingID:     booking.ID,
			Amount:        booking.TotalPrice,
			PaymentMethod: "credit_card",
		}
		jsonBody, err := json.Marshal(paymentReq)
		if err != nil {
			utils.LogError("Failed to marshal payment request for booking %s: %v", booking.ID, err)
			return
		}

		callCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(callCtx, "POST", s.PaymentServiceAPIURL+"/payments", bytes.NewBuffer(jsonBody))
		if err != nil {
			utils.LogError("Failed to create HTTP request to Payment Service for booking %s: %v", booking.ID, err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			utils.LogError("Failed to send payment request for booking %s to Payment Service: %v", booking.ID, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorBody map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
				utils.LogError("Payment Service returned non-200 status for booking %s: %d, no readable error body", booking.ID, resp.StatusCode)
				return
			}
			utils.LogError("Payment Service returned non-200 status for booking %s: %d, error: %s", booking.ID, resp.StatusCode, errorBody["error"])
			return
		}

		utils.LogInfo("Payment request sent to Payment Service for booking %s", booking.ID)
	}()

	var bookedSeatResponses []models.SeatResponse
	for _, seat := range seatsToBook {
		bookedSeatResponses = append(bookedSeatResponses, seat.ToSeatResponse())
	}
	resp := models.BookingResponse{
		ID:          booking.ID,
		UserID:      booking.UserID,
		ConcertID:   booking.ConcertID,
		TotalPrice:  totalPrice,
		Status:      booking.Status,
		ExpiresAt:   booking.ExpiresAt,
		BookedSeats: bookedSeatResponses,
		ConcertName: concert.Name,
		ConcertDate: concert.Date,
		CreatedAt:   booking.CreatedAt,
		UpdatedAt:   booking.UpdatedAt,
	}

	buyerResp := buyer.ToBuyerResponse()
	resp.BuyerInfo = &buyerResp
	if req.TicketHolderInfo != nil {
		ticketHolderResp := ticketHolder.ToTicketHolderResponse()
		resp.TicketHolderInfo = &ticketHolderResp
	}

	return &resp, nil
}

func (s *BookingService) GetBookingDetails(ctx context.Context, bookingID string, userID uint) (*models.BookingResponse, error) {
	booking, err := s.BookingRepo.GetBookingByID(bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking not found")
		}
		utils.LogError("DB error getting booking %s: %v", bookingID, err)
		return nil, errors.New("failed to retrieve booking details")
	}

	if booking.UserID != userID {
		utils.LogWarning("Unauthorized attempt to view booking %s by user %d. Owned by user %d.", bookingID, userID, booking.UserID)
		return nil, errors.New("unauthorized: you can only view your own bookings")
	}

	var bookedSeatResponses []models.SeatResponse
	for _, seat := range booking.Seats {

		if seat != nil {
			bookedSeatResponses = append(bookedSeatResponses, seat.ToSeatResponse())
		}
	}

	resp := models.BookingResponse{
		ID:          booking.ID,
		UserID:      booking.UserID,
		ConcertID:   booking.ConcertID,
		TotalPrice:  booking.TotalPrice,
		Status:      booking.Status,
		PaymentID:   booking.PaymentID,
		ExpiresAt:   booking.ExpiresAt,
		BookedSeats: bookedSeatResponses,
		ConcertName: booking.Concert.Name,
		ConcertDate: booking.Concert.Date,
		CreatedAt:   booking.CreatedAt,
		UpdatedAt:   booking.UpdatedAt,
	}

	if booking.Buyer != nil {
		buyerResp := booking.Buyer.ToBuyerResponse()
		resp.BuyerInfo = &buyerResp
	}
	if booking.TicketHolder != nil {
		ticketHolderResp := booking.TicketHolder.ToTicketHolderResponse()
		resp.TicketHolderInfo = &ticketHolderResp
	}

	return &resp, nil
}

func (s *BookingService) GetBookingsByUserID(ctx context.Context, userID uint) ([]models.BookingResponse, error) {
	bookings, err := s.BookingRepo.GetBookingsByUserID(userID)
	if err != nil {
		utils.LogError("DB error getting bookings for user %d: %v", userID, err)
		return nil, errors.New("failed to retrieve user bookings")
	}

	var responses []models.BookingResponse
	for _, booking := range bookings {
		var bookedSeatResponses []models.SeatResponse
		for _, seat := range booking.Seats {

			if seat != nil {
				bookedSeatResponses = append(bookedSeatResponses, seat.ToSeatResponse())
			}
		}

		var buyerResp *models.BuyerResponse
		if booking.Buyer != nil {
			b := booking.Buyer.ToBuyerResponse()
			buyerResp = &b
		}
		var ticketHolderResp *models.TicketHolderResponse
		if booking.TicketHolder != nil {
			th := booking.TicketHolder.ToTicketHolderResponse()
			ticketHolderResp = &th
		}

		responses = append(responses, models.BookingResponse{
			ID:               booking.ID,
			UserID:           booking.UserID,
			ConcertID:        booking.ConcertID,
			TotalPrice:       booking.TotalPrice,
			Status:           booking.Status,
			PaymentID:        booking.PaymentID,
			ExpiresAt:        booking.ExpiresAt,
			BookedSeats:      bookedSeatResponses,
			ConcertName:      booking.Concert.Name,
			ConcertDate:      booking.Concert.Date,
			BuyerInfo:        buyerResp,
			TicketHolderInfo: ticketHolderResp,
			CreatedAt:        booking.CreatedAt,
			UpdatedAt:        booking.UpdatedAt,
		})
	}
	return responses, nil
}

func (s *BookingService) UpdateBookingStatusFromPayment(ctx context.Context, bookingID string, newStatus string, paymentID uint) error {
	booking, err := s.BookingRepo.GetBookingByID(bookingID)
	if err != nil {
		utils.LogError("Booking %s not found for status update from payment service: %v", bookingID, err)
		return errors.New("booking not found")
	}

	switch newStatus {
	case models.BookingStatusConfirmed:
		if booking.Status != models.BookingStatusPending {
			return fmt.Errorf("invalid status transition: booking %s is %s, cannot confirm", bookingID, booking.Status)
		}
		booking.Status = models.BookingStatusConfirmed
		booking.PaymentID = &paymentID
		booking.ExpiresAt = nil
		for _, seat := range booking.Seats {
			seat.Status = models.SeatStatusBooked
		}
		utils.LogInfo("Booking %s status updated to CONFIRMED. PaymentID: %d", bookingID, paymentID)

	case models.BookingStatusFailed:
		if booking.Status != models.BookingStatusPending {
			return fmt.Errorf("invalid status transition: booking %s is %s, cannot fail", bookingID, booking.Status)
		}
		booking.Status = models.BookingStatusFailed

		for _, seat := range booking.Seats {
			seat.Status = models.SeatStatusAvailable
			seat.UserID = nil
			seat.BookingID = nil
		}

		classQuantitiesToRevert := make(map[uint]int)
		for _, seat := range booking.Seats {
			classQuantitiesToRevert[seat.TicketClassID]++
		}
		for tcID, qty := range classQuantitiesToRevert {
			_, errRedis := utils.IncreaseAvailableSeatsAtomically(ctx, tcID, qty)
			if errRedis != nil {
				utils.LogError("Failed to increase available seats in Redis for class %d after payment failure of booking %s: %v", tcID, bookingID, errRedis)
			}
		}
		utils.LogWarning("Booking %s status updated to FAILED. PaymentID: %d. Seats released.", bookingID, paymentID)

	case models.BookingStatusCancelled:
		if booking.Status == models.BookingStatusConfirmed || booking.Status == models.BookingStatusFailed {
			return fmt.Errorf("invalid status transition: booking %s is %s, cannot be cancelled", bookingID, booking.Status)
		}
		booking.Status = models.BookingStatusCancelled
		for _, seat := range booking.Seats {
			seat.Status = models.SeatStatusAvailable
			seat.UserID = nil
			seat.BookingID = nil
		}

		classQuantitiesToRevert := make(map[uint]int)
		for _, seat := range booking.Seats {
			classQuantitiesToRevert[seat.TicketClassID]++
		}
		for tcID, qty := range classQuantitiesToRevert {
			_, errRedis := utils.IncreaseAvailableSeatsAtomically(ctx, tcID, qty)
			if errRedis != nil {
				utils.LogError("Failed to increase available seats in Redis for class %d after cancellation of booking %s: %v", tcID, bookingID, errRedis)
			}
		}
		utils.LogInfo("Booking %s status updated to CANCELLED.", bookingID)

	default:
		return fmt.Errorf("unsupported new booking status: %s", newStatus)
	}

	tx := s.BookingRepo.DB.Begin()
	if tx.Error != nil {
		utils.LogError("Failed to begin DB transaction for booking status update: %v", tx.Error)
		return errors.New("failed to initiate booking status update transaction")
	}
	tempBookingRepo := &repositories.BookingRepository{DB: tx}
	tempSeatRepo := &repositories.SeatRepository{DB: tx}
	tempTicketClassRepo := &repositories.TicketClassRepository{DB: tx}

	if err := tempSeatRepo.UpdateSeats(booking.Seats); err != nil {
		tx.Rollback()
		utils.LogError("Failed to update seat statuses in DB for booking %s: %v", bookingID, err)
		return errors.New("failed to update seat statuses")
	}

	classQuantitiesToRevert := make(map[uint]int)
	for _, seat := range booking.Seats {
		classQuantitiesToRevert[seat.TicketClassID]++
	}
	for tcID, qty := range classQuantitiesToRevert {
		ticketClass, err := tempTicketClassRepo.GetTicketClassByID(tcID)
		if err != nil {
			utils.LogError("Failed to get TicketClass %d for cancellation revert: %v", tcID, err)
			continue
		}
		ticketClass.AvailableSeatsInClass += qty
		if err := tempTicketClassRepo.UpdateTicketClass(tx, ticketClass); err != nil {
			utils.LogError("Failed to update TicketClass %d for cancellation revert: %v", tcID, err)
		}
	}

	if err := tempBookingRepo.UpdateBooking(booking); err != nil {
		tx.Rollback()
		utils.LogError("Failed to update booking %s status in DB: %v", bookingID, err)
		return errors.New("failed to update booking status")
	}

	tx.Commit()
	return nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID string, userID uint) error {
	booking, err := s.BookingRepo.GetBookingByID(bookingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("booking not found")
		}
		utils.LogError("DB error getting booking %s for cancellation: %v", bookingID, err)
		return errors.New("failed to retrieve booking details for cancellation")
	}

	if booking.UserID != userID {
		utils.LogWarning("Unauthorized attempt to cancel booking %s by user %d. Owned by user %d.", bookingID, userID, booking.UserID)
		return errors.New("unauthorized: you can only cancel your own bookings")
	}

	if booking.Status != models.BookingStatusPending {
		return fmt.Errorf("booking %s cannot be cancelled as its status is %s (only pending bookings can be cancelled)", bookingID, booking.Status)
	}

	booking.Status = models.BookingStatusCancelled
	for _, seat := range booking.Seats {
		seat.Status = models.SeatStatusAvailable
		seat.UserID = nil
		seat.BookingID = nil
	}

	tx := s.BookingRepo.DB.Begin()
	if tx.Error != nil {
		utils.LogError("Failed to begin DB transaction for booking cancellation: %v", tx.Error)
		return errors.New("failed to initiate cancellation transaction")
	}
	tempBookingRepo := &repositories.BookingRepository{DB: tx}
	tempSeatRepo := &repositories.SeatRepository{DB: tx}
	tempTicketClassRepo := &repositories.TicketClassRepository{DB: tx}

	if err := tempSeatRepo.UpdateSeats(booking.Seats); err != nil {
		tx.Rollback()
		utils.LogError("Failed to update seat statuses for booking %s cancellation: %v", bookingID, err)
		return errors.New("failed to release seats during cancellation")
	}

	classQuantitiesToRevert := make(map[uint]int)
	for _, seat := range booking.Seats {
		classQuantitiesToRevert[seat.TicketClassID]++
	}
	for tcID, qty := range classQuantitiesToRevert {
		ticketClass, err := tempTicketClassRepo.GetTicketClassByID(tcID)
		if err != nil {
			utils.LogError("Failed to get TicketClass %d for cancellation revert: %v", tcID, err)
			continue
		}
		ticketClass.AvailableSeatsInClass += qty
		if err := tempTicketClassRepo.UpdateTicketClass(tx, ticketClass); err != nil {
			utils.LogError("Failed to update TicketClass %d for cancellation revert: %v", tcID, err)
		}
	}

	if err := tempBookingRepo.UpdateBooking(booking); err != nil {
		tx.Rollback()
		utils.LogError("Failed to update booking %s status to cancelled: %v", bookingID, err)
		return errors.New("failed to update booking status to cancelled")
	}

	tx.Commit()

	numSeats := len(booking.Seats)
	if numSeats > 0 {
		classQuantitiesToRevert := make(map[uint]int)
		for _, seat := range booking.Seats {
			classQuantitiesToRevert[seat.TicketClassID]++
		}
		for tcID, qty := range classQuantitiesToRevert {
			_, errRedis := utils.IncreaseAvailableSeatsAtomically(ctx, tcID, qty)
			if errRedis != nil {
				utils.LogError("Failed to increase available seats in Redis for concert %d after cancellation of booking %s: %v", booking.ConcertID, bookingID, errRedis)
			}
		}
	}

	utils.LogInfo("Booking %s successfully cancelled by user %d. Seats released.", bookingID, userID)
	return nil
}

func (s *BookingService) CancelExpiredPendingBookings(ctx context.Context) error {
	expiredBookings, err := s.BookingRepo.GetExpiredPendingBookings()
	if err != nil {
		return fmt.Errorf("error fetching expired pending bookings: %w", err)
	}

	if len(expiredBookings) == 0 {
		utils.LogInfo("No expired pending bookings found.")
		return nil
	}

	for _, booking := range expiredBookings {
		utils.LogInfo("Automatically canceling expired booking %s (Concert ID: %d, User ID: %d)", booking.ID, booking.ConcertID, booking.UserID)

		booking.Status = models.BookingStatusCancelled

		for _, seat := range booking.Seats {
			seat.Status = models.SeatStatusAvailable
			seat.UserID = nil
			seat.BookingID = nil
		}

		tx := s.BookingRepo.DB.Begin()
		if tx.Error != nil {
			utils.LogError("Failed to begin DB transaction for auto-cancellation of booking %s: %v", booking.ID, tx.Error)
			continue
		}
		tempBookingRepo := &repositories.BookingRepository{DB: tx}
		tempSeatRepo := &repositories.SeatRepository{DB: tx}
		tempTicketClassRepo := &repositories.TicketClassRepository{DB: tx}

		if err := tempSeatRepo.UpdateSeats(booking.Seats); err != nil {
			tx.Rollback()
			utils.LogError("Failed to update seat statuses in DB for auto-cancellation of booking %s: %v", booking.ID, err)
			continue
		}

		classQuantitiesToRevert := make(map[uint]int)
		for _, seat := range booking.Seats {
			classQuantitiesToRevert[seat.TicketClassID]++
		}
		for tcID, qty := range classQuantitiesToRevert {
			ticketClass, err := tempTicketClassRepo.GetTicketClassByID(tcID)
			if err != nil {
				utils.LogError("Failed to get TicketClass %d for auto-cancellation revert: %v", tcID, err)
				continue
			}
			ticketClass.AvailableSeatsInClass += qty
			if err := tempTicketClassRepo.UpdateTicketClass(tx, ticketClass); err != nil {
				utils.LogError("Failed to update TicketClass %d for auto-cancellation revert: %v", tcID, err)
			}
			_, errRedis := utils.IncreaseAvailableSeatsAtomically(ctx, tcID, qty)
			if errRedis != nil {
				utils.LogError("Failed to increase available seats in Redis for class %d after auto-cancellation of booking %s: %v", tcID, booking.ID, errRedis)
			}
		}

		if err := tempBookingRepo.UpdateBooking(&booking); err != nil {
			tx.Rollback()
			utils.LogError("Failed to update booking %s status to cancelled during auto-cancellation: %v", booking.ID, err)
			continue
		}

		tx.Commit()
		utils.LogInfo("Booking %s successfully auto-cancelled. Seats released.", booking.ID)
	}
	return nil
}
