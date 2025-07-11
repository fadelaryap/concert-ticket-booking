package models

import "gorm.io/gorm"

type Buyer struct {
	gorm.Model
	BookingID   string `gorm:"not null;type:varchar(36)" json:"booking_id"`
	FullName    string `gorm:"not null" json:"full_name"`
	PhoneNumber string `gorm:"not null" json:"phone_number"`
	Email       string `gorm:"not null" json:"email"`
	KTPNumber   string `gorm:"not null;unique" json:"ktp_number"`
}

func (b *Buyer) ToBuyerResponse() BuyerResponse {
	return BuyerResponse{
		ID:          b.ID,
		FullName:    b.FullName,
		PhoneNumber: b.PhoneNumber,
		Email:       b.Email,
		KTPNumber:   b.KTPNumber,
	}
}

type TicketHolder struct {
	gorm.Model
	BookingID string `gorm:"not null;type:varchar(36)" json:"booking_id"`
	FullName  string `gorm:"not null" json:"full_name"`
	KTPNumber string `gorm:"not null;unique" json:"ktp_number"`
}

func (th *TicketHolder) ToTicketHolderResponse() TicketHolderResponse {
	return TicketHolderResponse{
		ID:        th.ID,
		FullName:  th.FullName,
		KTPNumber: th.KTPNumber,
	}
}

type BuyerRequest struct {
	FullName    string `json:"full_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,numeric,min=10,max=15"`
	Email       string `json:"email" validate:"required,email"`
	KTPNumber   string `json:"ktp_number" validate:"required,numeric,len=16"`
}

type TicketHolderRequest struct {
	FullName  string `json:"full_name" validate:"required"`
	KTPNumber string `json:"ktp_number" validate:"required,numeric,len=16"`
}

type BuyerResponse struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	KTPNumber   string `json:"ktp_number"`
}

type TicketHolderResponse struct {
	ID        uint   `json:"id"`
	FullName  string `json:"full_name"`
	KTPNumber string `json:"ktp_number"`
}
