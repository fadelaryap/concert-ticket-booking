package models

import (
	"gorm.io/gorm"
)

type TicketClass struct {
	gorm.Model
	ConcertID             uint    `gorm:"not null" json:"concert_id"`
	Name                  string  `gorm:"not null" json:"name" validate:"required,min=2"`
	Price                 float64 `gorm:"not null" json:"price" validate:"required,gt=0"`
	TotalSeatsInClass     int     `gorm:"not null" json:"total_seats_in_class" validate:"required,min=1"`
	AvailableSeatsInClass int     `gorm:"not null" json:"available_seats_in_class"`
}

type CreateTicketClassRequest struct {
	Name              string  `json:"name" validate:"required,min=2"`
	Price             float64 `json:"price" validate:"required,gt=0"`
	TotalSeatsInClass int     `json:"total_seats_in_class" validate:"required,min=1"`
}

type TicketClassResponse struct {
	ID                    uint    `json:"id"`
	Name                  string  `json:"name"`
	Price                 float64 `json:"price"`
	TotalSeatsInClass     int     `json:"total_seats_in_class"`
	AvailableSeatsInClass int     `json:"available_seats_in_class"`
}

func (tc *TicketClass) ToTicketClassResponse() TicketClassResponse {
	return TicketClassResponse{
		ID:                    tc.ID,
		Name:                  tc.Name,
		Price:                 tc.Price,
		TotalSeatsInClass:     tc.TotalSeatsInClass,
		AvailableSeatsInClass: tc.AvailableSeatsInClass,
	}
}
