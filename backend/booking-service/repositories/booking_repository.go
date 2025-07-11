package repositories

import (
	"backend/booking-service/models"
	"time"

	"gorm.io/gorm"
)

type BookingRepository struct {
	DB *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{DB: db}
}

func (r *BookingRepository) CreateBooking(booking *models.Booking) error {
	return r.DB.Create(booking).Error
}

func (r *BookingRepository) GetBookingByID(id string) (*models.Booking, error) {
	var booking models.Booking
	err := r.DB.Preload("Concert").Preload("Seats").Preload("Buyer").Preload("TicketHolder").First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) UpdateBooking(booking *models.Booking) error {
	return r.DB.Save(booking).Error
}

func (r *BookingRepository) GetBookingsByUserID(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.DB.Where("user_id = ?", userID).Preload("Concert").Preload("Seats").Preload("Buyer").Preload("TicketHolder").Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetUserActiveBookingsForConcert(userID, concertID uint) ([]models.Booking, error) {
	var bookings []models.Booking

	err := r.DB.Where("user_id = ? AND concert_id = ? AND (status = ? OR status = ?)",
		userID, concertID, models.BookingStatusPending, models.BookingStatusConfirmed).Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetExpiredPendingBookings() ([]models.Booking, error) {
	var bookings []models.Booking

	err := r.DB.Where("status = ? AND expires_at < ?", models.BookingStatusPending, time.Now()).
		Preload("Seats").
		Find(&bookings).Error
	return bookings, err
}
