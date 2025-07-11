package repositories

import (
	"backend/booking-service/models"

	"gorm.io/gorm"
)

type TicketClassRepository struct {
	DB *gorm.DB
}

func NewTicketClassRepository(db *gorm.DB) *TicketClassRepository {
	return &TicketClassRepository{DB: db}
}

func (r *TicketClassRepository) CreateTicketClass(db *gorm.DB, ticketClass *models.TicketClass) error {
	return db.Create(ticketClass).Error
}

func (r *TicketClassRepository) GetTicketClassesByConcertID(concertID uint) ([]models.TicketClass, error) {
	var ticketClasses []models.TicketClass
	err := r.DB.Where("concert_id = ?", concertID).Find(&ticketClasses).Error
	return ticketClasses, err
}

func (r *TicketClassRepository) GetTicketClassByID(id uint) (*models.TicketClass, error) {
	var ticketClass models.TicketClass
	err := r.DB.First(&ticketClass, id).Error
	return &ticketClass, err
}

func (r *TicketClassRepository) UpdateTicketClass(db *gorm.DB, ticketClass *models.TicketClass) error {
	return db.Save(ticketClass).Error
}
