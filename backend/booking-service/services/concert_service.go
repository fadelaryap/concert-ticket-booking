package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"backend/booking-service/models"
	"backend/booking-service/repositories"
	"backend/booking-service/utils"

	"gorm.io/gorm"
)

const (
	ConcertStatusPendingSeatCreation = "pending_seat_creation"
	ConcertStatusActive              = "active"
	ConcertStatusFailed              = "failed"
)

type ConcertService struct {
	ConcertRepo     *repositories.ConcertRepository
	SeatRepo        *repositories.SeatRepository
	TicketClassRepo *repositories.TicketClassRepository
}

func NewConcertService(cRepo *repositories.ConcertRepository, sRepo *repositories.SeatRepository, tcRepo *repositories.TicketClassRepository) *ConcertService {
	return &ConcertService{ConcertRepo: cRepo, SeatRepo: sRepo, TicketClassRepo: tcRepo}
}

func (s *ConcertService) CreateConcert(ctx context.Context, req *models.CreateConcertRequest) (*models.ConcertResponse, error) {
	totalSeats := 0
	for _, tcReq := range req.TicketClasses {
		totalSeats += tcReq.TotalSeatsInClass
	}
	if totalSeats == 0 {
		return nil, errors.New("total seats from ticket classes must be greater than 0")
	}

	concert := &models.Concert{
		Name:           req.Name,
		Artist:         req.Artist,
		Date:           req.Date,
		Venue:          req.Venue,
		TotalSeats:     totalSeats,
		AvailableSeats: totalSeats,
		Description:    req.Description,
		ImageUrl:       req.ImageUrl,
		Status:         models.ConcertStatusPendingSeatCreation,
	}

	var ticketClasses []models.TicketClass
	for _, tcReq := range req.TicketClasses {
		ticketClasses = append(ticketClasses, models.TicketClass{
			Name:                  tcReq.Name,
			Price:                 tcReq.Price,
			TotalSeatsInClass:     tcReq.TotalSeatsInClass,
			AvailableSeatsInClass: tcReq.TotalSeatsInClass,
		})
	}
	concert.TicketClasses = ticketClasses

	tx := s.ConcertRepo.DB.Begin()
	if tx.Error != nil {
		utils.LogError("Failed to begin DB transaction for concert creation: %v", tx.Error)
		return nil, errors.New("failed to initiate concert creation transaction")
	}

	if err := s.ConcertRepo.CreateConcert(tx, concert); err != nil {
		tx.Rollback()
		utils.LogError("Failed to create concert in DB (initial entry): %v", err)
		return nil, errors.New("failed to create concert initial entry")
	}

	tx.Commit()

	for _, tc := range concert.TicketClasses {

		err := utils.SetAvailableSeatsCacheByClass(context.Background(), tc.ID, tc.AvailableSeatsInClass)
		if err != nil {
			utils.LogWarning("Failed to cache available seats for ticket class %d (concert %d) in Redis: %v", tc.ID, concert.ID, err)
		}
	}

	if err := utils.SetAvailableSeatsCache(context.Background(), concert.ID, concert.AvailableSeats); err != nil {
		utils.LogWarning("Failed to cache total available seats for concert %d in Redis: %v", concert.ID, err)
	}

	var tcMessages []models.TicketClassMessage
	for _, tc := range concert.TicketClasses {
		tcMessages = append(tcMessages, models.TicketClassMessage{
			TicketClassID:     tc.ID,
			Name:              tc.Name,
			TotalSeatsInClass: tc.TotalSeatsInClass,
		})
	}
	msg := models.SeatCreationMessage{
		ConcertID:     concert.ID,
		TotalSeats:    concert.TotalSeats,
		TicketClasses: tcMessages,
	}
	msgBody, err := json.Marshal(msg)
	if err != nil {
		utils.LogError("Failed to marshal seat creation message for concert %d: %v", concert.ID, err)
		s.ConcertRepo.DB.Model(&models.Concert{}).Where("id = ?", concert.ID).Update("status", models.ConcertStatusFailed)
		return nil, errors.New("failed to marshal seat creation message")
	}

	if err := utils.PublishMessage("", utils.SeatCreationQueue(), msgBody); err != nil {
		utils.LogError("Failed to publish seat creation message for concert %d to RabbitMQ: %v", concert.ID, err)
		s.ConcertRepo.DB.Model(&models.Concert{}).Where("id = ?", concert.ID).Update("status", models.ConcertStatusFailed)
		return nil, errors.New("concert created, but failed to initiate seat creation process")
	}

	utils.LogInfo("Concert '%s' created (ID: %d), seat creation offloaded to background worker.", concert.Name, concert.ID)

	resp := concert.ToConcertResponse()
	return &resp, nil
}

func (s *ConcertService) GetConcerts(ctx context.Context) ([]models.ConcertResponse, error) {
	concerts, err := s.ConcertRepo.GetConcerts()
	if err != nil {
		utils.LogError("Failed to get concerts from DB: %v", err)
		return nil, errors.New("failed to retrieve concerts")
	}

	var responses []models.ConcertResponse
	for _, c := range concerts {

		availableSeats, err := utils.GetAvailableSeatsFromCache(ctx, c.ID)
		if err == nil {
			c.AvailableSeats = availableSeats
		} else {
			utils.LogWarning("Cache miss for available seats for concert %d. Error: %v. Falling back to DB count.", c.ID, err)
			dbSeats, dbErr := s.SeatRepo.GetSeatsByConcertID(c.ID)
			if dbErr != nil {
				utils.LogError("Failed to get seats from DB for concert %d during fallback: %v", c.ID, dbErr)
				c.AvailableSeats = 0
			} else {
				availableCount := 0
				for _, seat := range dbSeats {
					if seat.Status == models.SeatStatusAvailable {
						availableCount++
					}
				}
				c.AvailableSeats = availableCount

				if err := utils.SetAvailableSeatsCache(ctx, c.ID, availableCount); err != nil {
					utils.LogWarning("Failed to re-cache available seats for concert %d: %v", c.ID, err)
				}
			}
		}

		for i := range c.TicketClasses {
			tc := &c.TicketClasses[i]
			availableSeatsClass, errClass := utils.GetAvailableSeatsCacheByClass(ctx, tc.ID)
			if errClass == nil {
				tc.AvailableSeatsInClass = availableSeatsClass
			} else {
				utils.LogWarning("Cache miss for available seats for ticket class %d. Error: %v. Falling back to DB count.", tc.ID, errClass)

				seatsInClass, err := s.SeatRepo.GetSeatsByTicketClassID(tc.ID)
				if err != nil {
					utils.LogError("Failed to get seats for ticket class %d from DB during fallback: %v", tc.ID, err)
					tc.AvailableSeatsInClass = 0
				} else {
					availableCount := 0
					for _, seat := range seatsInClass {
						if seat.Status == models.SeatStatusAvailable {
							availableCount++
						}
					}
					tc.AvailableSeatsInClass = availableCount
					if err := utils.SetAvailableSeatsCacheByClass(ctx, tc.ID, availableCount); err != nil {
						utils.LogWarning("Failed to re-cache available seats for ticket class %d: %v", tc.ID, err)
					}
				}
			}
		}
		responses = append(responses, c.ToConcertResponse())
	}
	return responses, nil
}

func (s *ConcertService) GetConcertByID(ctx context.Context, id uint) (*models.ConcertResponse, error) {
	concert, err := s.ConcertRepo.GetConcertByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("concert not found")
		}
		utils.LogError("Failed to get concert ID %d from DB: %v", id, err)
		return nil, errors.New("failed to retrieve concert")
	}

	availableSeats, errCache := utils.GetAvailableSeatsFromCache(ctx, concert.ID)
	if errCache == nil {
		concert.AvailableSeats = availableSeats
	} else {
		utils.LogWarning("Cache miss for available seats for concert %d. Error: %v. Falling back to DB count.", concert.ID, errCache)
		dbSeats, dbErr := s.SeatRepo.GetSeatsByConcertID(concert.ID)
		if dbErr != nil {
			utils.LogError("Failed to get seats from DB for concert %d during fallback: %v", concert.ID, dbErr)
			concert.AvailableSeats = 0
		} else {
			availableCount := 0
			for _, seat := range dbSeats {
				if seat.Status == models.SeatStatusAvailable {
					availableCount++
				}
			}
			concert.AvailableSeats = availableCount

			if err := utils.SetAvailableSeatsCache(ctx, concert.ID, availableCount); err != nil {
				utils.LogWarning("Failed to re-cache available seats for concert %d: %v", concert.ID, err)
			}
		}
	}

	for i := range concert.TicketClasses {
		tc := &concert.TicketClasses[i]
		availableSeatsClass, errClass := utils.GetAvailableSeatsCacheByClass(ctx, tc.ID)
		if errClass == nil {
			tc.AvailableSeatsInClass = availableSeatsClass
		} else {
			utils.LogWarning("Cache miss for available seats for ticket class %d. Error: %v. Falling back to DB count.", tc.ID, errClass)
			seatsInClass, err := s.SeatRepo.GetSeatsByTicketClassID(tc.ID)
			if err != nil {
				utils.LogError("Failed to get seats for ticket class %d from DB during fallback: %v", tc.ID, err)
				tc.AvailableSeatsInClass = 0
			} else {
				availableCount := 0
				for _, seat := range seatsInClass {
					if seat.Status == models.SeatStatusAvailable {
						availableCount++
					}
				}
				tc.AvailableSeatsInClass = availableCount
				if err := utils.SetAvailableSeatsCacheByClass(ctx, tc.ID, availableCount); err != nil {
					utils.LogWarning("Failed to re-cache available seats for ticket class %d: %v", tc.ID, err)
				}
			}
		}
	}

	resp := concert.ToConcertResponse()
	return &resp, nil
}

func (s *ConcertService) GetSeatsForConcert(ctx context.Context, concertID uint) ([]models.SeatResponse, error) {

	concert, err := s.ConcertRepo.GetConcertByID(concertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("concert not found")
		}
		utils.LogError("Failed to verify concert existence for seat fetching: %v", err)
		return nil, errors.New("database error checking concert existence")
	}

	ticketClassNames := make(map[uint]string)
	for _, tc := range concert.TicketClasses {
		ticketClassNames[tc.ID] = tc.Name
	}

	seats, err := s.SeatRepo.GetSeatsByConcertID(concertID)
	if err != nil {
		utils.LogError("Failed to get seats for concert %d from DB: %v", concertID, err)
		return nil, errors.New("failed to retrieve seats")
	}

	var responses []models.SeatResponse
	for _, seat := range seats {
		resp := seat.ToSeatResponse()
		resp.TicketClassName = ticketClassNames[seat.TicketClassID]
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *ConcertService) ProcessSeatCreationMessage(body []byte) error {
	var msg models.SeatCreationMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		utils.LogError("Failed to unmarshal seat creation message: %v", err)
		return err
	}

	utils.LogInfo("Processing background seat creation for Concert ID: %d, Total Seats: %d, Classes: %v", msg.ConcertID, msg.TotalSeats, msg.TicketClasses)

	concert, err := s.ConcertRepo.GetConcertByID(msg.ConcertID)
	if err != nil {
		utils.LogError("Concert ID %d not found for seat creation background process: %v", msg.ConcertID, err)
		s.ConcertRepo.DB.Model(&models.Concert{}).Where("id = ?", msg.ConcertID).Update("status", models.ConcertStatusFailed)
		return err
	}
	if concert.Status != models.ConcertStatusPendingSeatCreation {
		utils.LogWarning("Concert %d is not in 'pending_seat_creation' status. Current status: %s. Skipping seat creation.", msg.ConcertID, concert.Status)
		return nil
	}

	tx := s.ConcertRepo.DB.Begin()
	if tx.Error != nil {
		utils.LogError("Failed to begin DB transaction for background seat creation for concert %d: %v", msg.ConcertID, tx.Error)
		s.ConcertRepo.DB.Model(&models.Concert{}).Where("id = ?", msg.ConcertID).Update("status", models.ConcertStatusFailed)
		return errors.New("failed to initiate background seat creation transaction")
	}

	ticketClassesMap := make(map[uint]models.TicketClassMessage)
	for _, tcMsg := range msg.TicketClasses {
		ticketClassesMap[tcMsg.TicketClassID] = tcMsg
	}

	var allSeatsToCreate []models.Seat
	for _, tc := range concert.TicketClasses {
		tcMsg, exists := ticketClassesMap[tc.ID]
		if !exists {
			utils.LogError("TicketClass ID %d for concert %d not found in message. Skipping seats for this class.", tc.ID, msg.ConcertID)
			continue
		}

		for i := 0; i < tcMsg.TotalSeatsInClass; i++ {
			seatNumber := fmt.Sprintf("%s-S%d", tc.Name, i+1)
			allSeatsToCreate = append(allSeatsToCreate, models.Seat{
				ConcertID:     msg.ConcertID,
				TicketClassID: tc.ID,
				SeatNumber:    seatNumber,
				Status:        models.SeatStatusAvailable,
			})
		}
	}

	const seatBatchSize = 200
	if err := s.SeatRepo.CreateSeatsInBatches(tx, allSeatsToCreate, seatBatchSize); err != nil {
		tx.Rollback()
		utils.LogError("Failed to create seats in background for concert %d: %v", msg.ConcertID, err)
		s.ConcertRepo.DB.Model(&models.Concert{}).Where("id = ?", msg.ConcertID).Update("status", models.ConcertStatusFailed)
		return errors.New("failed to create seats in background")
	}

	for i := range concert.TicketClasses {
		tc := &concert.TicketClasses[i]
		tcMsg, exists := ticketClassesMap[tc.ID]
		if exists {
			tc.AvailableSeatsInClass = tcMsg.TotalSeatsInClass
			if err := s.TicketClassRepo.UpdateTicketClass(tx, tc); err != nil {
				utils.LogError("Failed to update TicketClass %d available seats for concert %d: %v", tc.ID, msg.ConcertID, err)
			}
		}
	}

	concert.Status = models.ConcertStatusActive
	concert.AvailableSeats = concert.TotalSeats
	if err := s.ConcertRepo.UpdateConcert(tx, concert); err != nil {
		tx.Rollback()
		utils.LogError("Failed to update concert status to 'active' after seat creation for concert %d: %v", msg.ConcertID, err)
		return errors.New("failed to update concert status after seat creation")
	}

	if err := utils.SetAvailableSeatsCache(context.Background(), concert.ID, concert.AvailableSeats); err != nil {
		utils.LogWarning("Failed to cache initial available seats for concert %d in Redis: %v", concert.ID, err)
	}

	for _, tc := range concert.TicketClasses {
		if err := utils.SetAvailableSeatsCacheByClass(context.Background(), tc.ID, tc.AvailableSeatsInClass); err != nil {
			utils.LogWarning("Failed to cache available seats for ticket class %d in Redis after seat creation: %v", tc.ID, err)
		}
	}

	tx.Commit()
	utils.LogInfo("Successfully created %d seats for Concert ID: %d and set status to ACTIVE.", msg.TotalSeats, msg.ConcertID)
	return nil
}
