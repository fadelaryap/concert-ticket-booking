package models

import (
	"gorm.io/gorm"
)

const (
	SeatStatusAvailable = "available"
	SeatStatusReserved  = "reserved"
	SeatStatusBooked    = "booked"
)

type Seat struct {
	gorm.Model
	ConcertID     uint   `gorm:"not null" json:"concert_id"`
	TicketClassID uint   `gorm:"not null" json:"ticket_class_id"`
	SeatNumber    string `gorm:"not null" json:"seat_number" validate:"required"`
	Status        string `gorm:"not null;default:'available'" json:"status"`
	UserID        *uint  `json:"user_id"`
	BookingID     *uint  `json:"booking_id"`
}

type SeatResponse struct {
	ID              uint   `json:"id"`
	SeatNumber      string `json:"seat_number"`
	Status          string `json:"status"`
	ConcertID       uint   `json:"concert_id"`
	TicketClassID   uint   `json:"ticket_class_id"`
	TicketClassName string `json:"ticket_class_name"`
}

func (s *Seat) ToSeatResponse() SeatResponse {
	return SeatResponse{
		ID:            s.ID,
		SeatNumber:    s.SeatNumber,
		Status:        s.Status,
		ConcertID:     s.ConcertID,
		TicketClassID: s.TicketClassID,
	}
}
