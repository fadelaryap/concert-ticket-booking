package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"backend/payment-service/models"
	"backend/payment-service/services"
	"backend/payment-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	PaymentService *services.PaymentService
	Validate       *validator.Validate
}

func NewPaymentController(ps *services.PaymentService) *PaymentController {
	return &PaymentController{
		PaymentService: ps,
		Validate:       validator.New(),
	}
}

// @Summary Process a payment for a booking
// @Description Initiate payment processing for a given booking. Sensitive card data is for simulation only.
// @Tags Payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param payment body models.ProcessPaymentRequest true "Payment request data"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request - Invalid input or validation errors"
// @Failure 401 {object} map[string]string "Unauthorized - Missing or invalid token"
// @Failure 500 {object} map[string]string "Internal Server Error - Failed to process payment"
// @Router /payments [post]
func (ctrl *PaymentController) ProcessPayment(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		utils.LogWarning("User ID not found in context for payment processing. Assuming internal call or direct frontend call without prior booking check.")

	} else {
		utils.LogInfo("Processing payment initiated by user ID: %d", userID.(uint))
	}

	var req models.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Invalid JSON body for process payment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for process payment: %v", validationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(validationErrors)})
		return
	}

	resp, err := ctrl.PaymentService.ProcessPayment(c.Request.Context(), &req)
	if err != nil {
		utils.LogError("Failed to process payment for booking %d: %v", req.BookingID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get payment details by ID
// @Description Retrieve details of a specific payment by its ID.
// @Tags Payments
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Payment ID"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request - Invalid payment ID"
// @Failure 401 {object} map[string]string "Unauthorized - Missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - Not authorized to view this payment"
// @Failure 404 {object} map[string]string "Not Found - Payment not found"
// @Failure 500 {object} map[string]string "Internal Server Error - Failed to retrieve payment"
// @Router /payments/{id} [get]
func (ctrl *PaymentController) GetPaymentByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("User ID not found in context for get payment by ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID format"})
		return
	}

	resp, err := ctrl.PaymentService.GetPaymentDetails(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		utils.LogError("Failed to get payment ID %d for user %d: %v", id, userID, err)
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
