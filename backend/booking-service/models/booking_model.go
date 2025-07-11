package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	BookingStatusFailed    = "failed"
)

type Booking struct {
	gorm.Model
	ID         string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	ConcertID  uint       `gorm:"not null" json:"concert_id"`
	SeatIDs    string     `gorm:"type:text;not null" json:"seat_ids"`
	TotalPrice float64    `gorm:"not null" json:"total_price"`
	Status     string     `gorm:"not null;default:'pending'" json:"status"`
	PaymentID  *uint      `json:"payment_id"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Concert    Concert    `gorm:"foreignKey:ConcertID" json:"-"`
	Seats      []*Seat    `gorm:"many2many:booking_seats;foreignKey:ID;joinForeignKey:booking_id;References:ID;joinReferences:seat_id" json:"-"`

	Buyer        *Buyer        `gorm:"foreignKey:BookingID;references:ID" json:"-"`
	TicketHolder *TicketHolder `gorm:"foreignKey:BookingID;references:ID" json:"-"`
}

type CreateBookingRequest struct {
	ConcertID        uint                    `json:"concert_id" validate:"required"`
	TicketsByClass   []TicketQuantityByClass `json:"tickets_by_class" validate:"required,min=1,dive"`
	BuyerInfo        BuyerRequest            `json:"buyer_info" validate:"required"`
	TicketHolderInfo *TicketHolderRequest    `json:"ticket_holder_info"`
}

type TicketQuantityByClass struct {
	TicketClassID uint `json:"ticket_class_id" validate:"required"`
	Quantity      int  `json:"quantity" validate:"required,min=1"`
}

type BookingResponse struct {
	ID               string                `json:"id"`
	UserID           uint                  `json:"user_id"`
	ConcertID        uint                  `json:"concert_id"`
	TotalPrice       float64               `json:"total_price"`
	Status           string                `json:"status"`
	PaymentID        *uint                 `json:"payment_id"`
	ExpiresAt        *time.Time            `json:"expires_at"`
	BookedSeats      []SeatResponse        `json:"booked_seats"`
	ConcertName      string                `json:"concert_name"`
	ConcertDate      time.Time             `json:"concert_date"`
	BuyerInfo        *BuyerResponse        `json:"buyer_info"`
	TicketHolderInfo *TicketHolderResponse `json:"ticket_holder_info"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type UpdateBookingStatusRequest struct {
	Status    string `json:"status" validate:"required"`
	PaymentID uint   `json:"payment_id" validate:"required"`
}
