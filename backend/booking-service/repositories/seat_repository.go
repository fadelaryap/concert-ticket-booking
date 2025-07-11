package repositories

import (
	"backend/booking-service/models"

	"gorm.io/gorm"
)

type SeatRepository struct {
	DB *gorm.DB
}

func NewSeatRepository(db *gorm.DB) *SeatRepository {
	return &SeatRepository{DB: db}
}

func (r *SeatRepository) CreateSeats(db *gorm.DB, seats []models.Seat) error {

	if len(seats) == 0 {
		return nil
	}
	return db.Create(seats).Error
}

func (r *SeatRepository) CreateSeatsInBatches(db *gorm.DB, seats []models.Seat, batchSize int) error {

	if len(seats) == 0 {
		return nil
	}
	return db.CreateInBatches(seats, batchSize).Error
}

func (r *SeatRepository) GetSeatsByConcertID(concertID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.DB.Where("concert_id = ?", concertID).Find(&seats).Error
	return seats, err
}

func (r *SeatRepository) GetSeatsByConcertIDAndNumbers(concertID uint, seatNumbers []string) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.DB.Where("concert_id = ? AND seat_number IN ?", concertID, seatNumbers).Find(&seats).Error
	return seats, err
}

func (r *SeatRepository) UpdateSeat(seat *models.Seat) error {
	return r.DB.Save(seat).Error
}

func (r *SeatRepository) UpdateSeats(seats []*models.Seat) error {

	if len(seats) == 0 {

		return nil
	}

	return r.DB.Save(&seats).Error
}

func (r *SeatRepository) GetSeatsByTicketClassID(ticketClassID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.DB.Where("ticket_class_id = ?", ticketClassID).Find(&seats).Error
	return seats, err
}
