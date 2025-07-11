package repositories

import (
	"backend/booking-service/models"

	"gorm.io/gorm"
)

type ConcertRepository struct {
	DB *gorm.DB
}

func NewConcertRepository(db *gorm.DB) *ConcertRepository {
	return &ConcertRepository{DB: db}
}

func (r *ConcertRepository) CreateConcert(db *gorm.DB, concert *models.Concert) error {

	return db.Create(concert).Error
}

func (r *ConcertRepository) GetConcerts() ([]models.Concert, error) {
	var concerts []models.Concert
	err := r.DB.Preload("TicketClasses").Find(&concerts).Error
	return concerts, err
}

func (r *ConcertRepository) GetConcertByID(id uint) (*models.Concert, error) {
	var concert models.Concert
	err := r.DB.Preload("TicketClasses").First(&concert, id).Error
	return &concert, err
}

func (r *ConcertRepository) UpdateConcert(db *gorm.DB, concert *models.Concert) error {
	return db.Save(concert).Error
}

func (r *ConcertRepository) GetConcertsByStatus(status string) ([]models.Concert, error) {
	var concerts []models.Concert
	err := r.DB.Where("status = ?", status).Preload("TicketClasses").Find(&concerts).Error
	return concerts, err
}
