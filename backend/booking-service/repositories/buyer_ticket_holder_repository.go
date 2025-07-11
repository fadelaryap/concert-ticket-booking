package repositories

import (
	"backend/booking-service/models"

	"gorm.io/gorm"
)

type BuyerRepository struct {
	DB *gorm.DB
}

func NewBuyerRepository(db *gorm.DB) *BuyerRepository {
	return &BuyerRepository{DB: db}
}

func (r *BuyerRepository) CreateBuyer(db *gorm.DB, buyer *models.Buyer) error {
	return db.Create(buyer).Error
}

func (r *BuyerRepository) GetBuyerByBookingID(bookingID uint) (*models.Buyer, error) {
	var buyer models.Buyer
	err := r.DB.Where("booking_id = ?", bookingID).First(&buyer).Error
	return &buyer, err
}

type TicketHolderRepository struct {
	DB *gorm.DB
}

func NewTicketHolderRepository(db *gorm.DB) *TicketHolderRepository {
	return &TicketHolderRepository{DB: db}
}

func (r *TicketHolderRepository) CreateTicketHolder(db *gorm.DB, ticketHolder *models.TicketHolder) error {
	return db.Create(ticketHolder).Error
}

func (r *TicketHolderRepository) GetTicketHolderByBookingID(bookingID uint) (*models.TicketHolder, error) {
	var ticketHolder models.TicketHolder
	err := r.DB.Where("booking_id = ?", bookingID).First(&ticketHolder).Error
	return &ticketHolder, err
}
