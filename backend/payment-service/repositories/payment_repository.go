package repositories

import (
	"backend/payment-service/models"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (r *PaymentRepository) CreatePayment(payment *models.Payment) error {
	return r.DB.Create(payment).Error
}

func (r *PaymentRepository) GetPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.DB.First(&payment, id).Error
	return &payment, err
}

func (r *PaymentRepository) GetPaymentByBookingID(bookingID uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.DB.Where("booking_id = ?", bookingID).First(&payment).Error
	return &payment, err
}

func (r *PaymentRepository) UpdatePayment(payment *models.Payment) error {
	return r.DB.Save(payment).Error
}
