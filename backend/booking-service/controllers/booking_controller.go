package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"backend/booking-service/models"
	"backend/booking-service/services"
	"backend/booking-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type BookingController struct {
	BookingService *services.BookingService
	Validate       *validator.Validate
}

func NewBookingController(bs *services.BookingService) *BookingController {
	return &BookingController{
		BookingService: bs,
		Validate:       validator.New(),
	}
}

// @Summary Create a new booking
// @Description Creates a new booking for a concert with specified tickets by class.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param createBookingRequest body models.CreateBookingRequest true "Booking request details"
// @Security ApiKeyAuth
// @Success 201 {object} models.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings [post]
func (ctrl *BookingController) CreateBooking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("UserID not found in context for CreateBooking")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogWarning("Invalid request body for CreateBooking: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for CreateBooking: %v", validationErrors)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: utils.FormatValidationErrors(validationErrors)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	bookingResp, err := ctrl.BookingService.CreateBooking(ctx, userID.(uint), &req)
	if err != nil {
		utils.LogError("Failed to create booking for user %d: %v", userID.(uint), err)
		if strings.Contains(err.Error(), "not enough overall seats available") ||
			strings.Contains(err.Error(), "ticket class ID not found") ||
			strings.Contains(err.Error(), "failed to reserve tickets for class") ||
			strings.Contains(err.Error(), "concert is not active for booking") ||
			strings.Contains(err.Error(), "invalid total number of tickets requested") ||
			strings.Contains(err.Error(), "you already have an active (pending or confirmed) booking for this concert") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bookingResp)
}

// @Summary Get booking details by ID
// @Description Retrieves the details of a specific booking.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID (UUID)"
// @Security ApiKeyAuth
// @Success 200 {object} models.BookingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id} [get]
func (ctrl *BookingController) GetBookingByID(c *gin.Context) {

	bookingID := c.Param("id")
	if bookingID == "" {
		utils.LogWarning("Invalid booking ID format: empty ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid booking ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("UserID not found in context for GetBookingDetails")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	bookingResp, err := ctrl.BookingService.GetBookingDetails(ctx, bookingID, userID.(uint))
	if err != nil {
		utils.LogError("Failed to get booking %s details for user %d: %v", bookingID, userID.(uint), err)
		if err.Error() == "booking not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err.Error() == "unauthorized: you can only view your own bookings" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookingResp)
}

// @Summary Get user's bookings
// @Description Retrieves all bookings made by the authenticated user.
// @Tags Bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.BookingResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /my-bookings [get]
func (ctrl *BookingController) GetMyBookings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("UserID not found in context for GetMyBookings")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	bookingsResp, err := ctrl.BookingService.GetBookingsByUserID(ctx, userID.(uint))
	if err != nil {
		utils.LogError("Failed to get bookings for user %d: %v", userID.(uint), err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookingsResp)
}

// @Summary Update booking status (Internal)
// @Description Internal endpoint for payment service to update booking status.
// @Tags Bookings (Internal)
// @Accept json
// @Produce json
// @Param id path string true "Booking ID (UUID)"
// @Param updateBookingStatusRequest body models.UpdateBookingStatusRequest true "Update booking status request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /internal/bookings/{id}/status [put]
func (ctrl *BookingController) UpdateBookingStatusInternal(c *gin.Context) {

	bookingID := c.Param("id")
	if bookingID == "" {
		utils.LogWarning("Invalid booking ID format for internal status update: empty ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid booking ID"})
		return
	}

	var req models.UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogWarning("Invalid request body for internal status update: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for internal status update: %v", validationErrors)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: utils.FormatValidationErrors(validationErrors)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	err := ctrl.BookingService.UpdateBookingStatusFromPayment(ctx, bookingID, req.Status, req.PaymentID)
	if err != nil {
		utils.LogError("Failed to update booking %s status internally: %v", bookingID, err)
		if err.Error() == "booking not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Booking status updated successfully"})
}

// @Summary Cancel a pending booking
// @Description Allows a user to cancel their pending booking.
// @Tags Bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID (UUID)"
// @Security ApiKeyAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /bookings/{id}/cancel [put]
func (ctrl *BookingController) CancelBooking(c *gin.Context) {

	bookingID := c.Param("id")
	if bookingID == "" {
		utils.LogWarning("Invalid booking ID format for CancelBooking: empty ID")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid booking ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("UserID not found in context for CancelBooking")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	err := ctrl.BookingService.CancelBooking(ctx, bookingID, userID.(uint))
	if err != nil {
		utils.LogError("Failed to cancel booking %s for user %d: %v", bookingID, userID.(uint), err)
		if err.Error() == "booking not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}

		if strings.Contains(err.Error(), "cannot be cancelled") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Booking cancelled successfully"})
}
