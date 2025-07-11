package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"backend/payment-service/models"
	"backend/payment-service/repositories"
	"backend/payment-service/utils"

	"gorm.io/gorm"
)

const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
)

type PaymentService struct {
	PaymentRepo          *repositories.PaymentRepository
	BookingServiceAPIURL string
}

func NewPaymentService(pRepo *repositories.PaymentRepository, bookingServiceAPIURL string) *PaymentService {
	return &PaymentService{
		PaymentRepo:          pRepo,
		BookingServiceAPIURL: bookingServiceAPIURL,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *models.ProcessPaymentRequest) (*models.PaymentResponse, error) {

	payment := &models.Payment{
		BookingID:     req.BookingID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        PaymentStatusPending,
	}

	if err := s.PaymentRepo.CreatePayment(payment); err != nil {
		utils.LogError("Failed to create pending payment record for booking %s: %v", req.BookingID, err)
		return nil, errors.New("failed to initiate payment: database error")
	}

	gatewayReq := utils.SimulatePaymentGatewayRequest{
		Amount:        req.Amount,
		CardNumber:    req.CardNumber,
		ExpiryDate:    req.ExpiryDate,
		CVV:           req.CVV,
		PaymentMethod: req.PaymentMethod,
	}
	gatewayResp := utils.SimulatePaymentGateway(gatewayReq)

	payment.TransactionID = gatewayResp.TransactionID
	payment.PaymentGatewayResponse = fmt.Sprintf("Status: %s, Message: %s, Fee: %.2f", gatewayResp.Status, gatewayResp.Message, gatewayResp.GatewayFee)

	var bookingNewStatus string
	if gatewayResp.Status == "success" {
		payment.Status = PaymentStatusCompleted
		bookingNewStatus = "confirmed"
		utils.LogInfo("Payment for booking %s completed successfully. TxID: %s", req.BookingID, payment.TransactionID)
	} else {
		payment.Status = PaymentStatusFailed
		bookingNewStatus = "failed"
		utils.LogWarning("Payment for booking %s failed. Reason: %s. TxID: %s", req.BookingID, gatewayResp.Message, payment.TransactionID)
	}

	if err := s.PaymentRepo.UpdatePayment(payment); err != nil {
		utils.LogError("Failed to update payment record %d after gateway response: %v", payment.ID, err)

		return nil, errors.New("payment processed but failed to update record")
	}

	err := s.SendBookingStatusUpdateToBookingService(ctx, payment.BookingID, bookingNewStatus, payment.ID)
	if err != nil {
		utils.LogError("Failed to notify booking service for booking %s status update to %s: %v", payment.BookingID, bookingNewStatus, err)

	}

	resp := payment.ToPaymentResponse()
	return &resp, nil
}

func (s *PaymentService) SendBookingStatusUpdateToBookingService(ctx context.Context, bookingID string, status string, paymentID uint) error {
	url := fmt.Sprintf("%s/internal/bookings/%s/status", s.BookingServiceAPIURL, bookingID)

	requestBody := models.UpdateBookingStatusInternalRequest{
		Status:    status,
		PaymentID: paymentID,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal booking status update request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request to booking service: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to booking service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]string

		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
			return fmt.Errorf("booking service returned non-200 status: %d, no readable error body", resp.StatusCode)
		}
		return fmt.Errorf("booking service returned non-200 status: %d, error: %s", resp.StatusCode, errorBody["error"])
	}

	utils.LogInfo("Successfully notified booking service for booking %s status update to %s", bookingID, status)
	return nil
}

func (s *PaymentService) GetPaymentDetails(ctx context.Context, paymentID uint, userID uint) (*models.PaymentResponse, error) {
	payment, err := s.PaymentRepo.GetPaymentByID(paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		utils.LogError("DB error getting payment %d: %v", paymentID, err)
		return nil, errors.New("failed to retrieve payment details")
	}

	utils.LogWarning("WARNING: User authorization for payment details (Payment ID: %d, User ID: %d) is NOT implemented via booking service. This is a security risk in production.", paymentID, userID)

	resp := payment.ToPaymentResponse()
	return &resp, nil
}
