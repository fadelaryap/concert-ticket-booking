package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	ConcertStatusPendingSeatCreation = "pending_seat_creation"
	ConcertStatusActive              = "active"
	ConcertStatusFailed              = "failed"
)

type Concert struct {
	gorm.Model
	Name           string        `gorm:"not null" json:"name" validate:"required,min=3"`
	Artist         string        `json:"artist" validate:"required"`
	Date           time.Time     `gorm:"not null" json:"date" validate:"required"`
	Venue          string        `gorm:"not null" json:"venue" validate:"required"`
	TotalSeats     int           `gorm:"not null" json:"total_seats" validate:"required,min=1"`
	AvailableSeats int           `gorm:"not null" json:"available_seats"`
	Description    string        `json:"description"`
	Status         string        `gorm:"default:'pending_seat_creation'" json:"status"`
	ImageUrl       string        `json:"image_url"`
	TicketClasses  []TicketClass `gorm:"foreignKey:ConcertID" json:"-"`
}

type CreateConcertRequest struct {
	Name          string                     `json:"name" validate:"required,min=3"`
	Artist        string                     `json:"artist" validate:"required"`
	Date          time.Time                  `json:"date" validate:"required"`
	Venue         string                     `json:"venue" validate:"required"`
	Description   string                     `json:"description"`
	ImageUrl      string                     `json:"image_url" validate:"url"`
	TicketClasses []CreateTicketClassRequest `json:"ticket_classes" validate:"required,min=1,dive"`
}

type ConcertResponse struct {
	ID             uint                  `json:"id"`
	Name           string                `json:"name"`
	Artist         string                `json:"artist"`
	Date           time.Time             `json:"date"`
	SetDateISO     string                `json:"date_iso"`
	Venue          string                `json:"venue"`
	TotalSeats     int                   `json:"total_seats"`
	AvailableSeats int                   `json:"available_seats"`
	Description    string                `json:"description"`
	Status         string                `json:"status"`
	ImageUrl       string                `json:"image_url"`
	TicketClasses  []TicketClassResponse `json:"ticket_classes"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

func (c *Concert) ToConcertResponse() ConcertResponse {
	var tcResponses []TicketClassResponse
	for _, tc := range c.TicketClasses {
		tcResponses = append(tcResponses, tc.ToTicketClassResponse())
	}
	return ConcertResponse{
		ID:             c.ID,
		Name:           c.Name,
		Artist:         c.Artist,
		Date:           c.Date,
		SetDateISO:     c.Date.Format(time.RFC3339),
		Venue:          c.Venue,
		TotalSeats:     c.TotalSeats,
		AvailableSeats: c.AvailableSeats,
		Description:    c.Description,
		Status:         c.Status,
		ImageUrl:       c.ImageUrl,
		TicketClasses:  tcResponses,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

type SeatCreationMessage struct {
	ConcertID     uint                 `json:"concert_id"`
	TotalSeats    int                  `json:"total_seats"`
	TicketClasses []TicketClassMessage `json:"ticket_classes"`
}

type TicketClassMessage struct {
	TicketClassID     uint   `json:"ticket_class_id"`
	Name              string `json:"name"`
	TotalSeatsInClass int    `json:"total_seats_in_class"`
}
