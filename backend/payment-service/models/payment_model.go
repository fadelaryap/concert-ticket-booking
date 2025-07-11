package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	BookingID              string  `gorm:"not null;index;type:varchar(36)" json:"booking_id" validate:"required"`
	Amount                 float64 `gorm:"not null" json:"amount" validate:"required,gt=0"`
	PaymentMethod          string  `gorm:"not null" json:"payment_method" validate:"required"`
	TransactionID          string  `gorm:"uniqueIndex" json:"transaction_id"`
	Status                 string  `gorm:"not null;default:'pending'" json:"status"`
	PaymentGatewayResponse string  `gorm:"type:text" json:"payment_gateway_response"`
}

type ProcessPaymentRequest struct {
	BookingID     string  `json:"booking_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" validate:"required"`

	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
}

type PaymentResponse struct {
	ID            uint      `json:"id"`
	BookingID     string    `json:"booking_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (p *Payment) ToPaymentResponse() PaymentResponse {
	return PaymentResponse{
		ID:            p.ID,
		BookingID:     p.BookingID,
		Amount:        p.Amount,
		PaymentMethod: p.PaymentMethod,
		TransactionID: p.TransactionID,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

type UpdateBookingStatusInternalRequest struct {
	Status    string `json:"status" validate:"required"`
	PaymentID uint   `json:"payment_id" validate:"required"`
}
