package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"backend/booking-service/models"
	"backend/booking-service/services"
	"backend/booking-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ConcertController struct {
	ConcertService *services.ConcertService
	Validate       *validator.Validate
}

func NewConcertController(cs *services.ConcertService) *ConcertController {
	return &ConcertController{
		ConcertService: cs,
		Validate:       validator.New(),
	}
}

// @Summary Create a new concert
// @Description Create a new concert event (Admin only). Seat creation is asynchronous.
// @Tags Concerts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param concert body models.CreateConcertRequest true "Concert creation data"
// @Success 201 {object} models.ConcertResponse "Concert created successfully, seat creation in background"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid input or validation errors"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 403 {object} ErrorResponse "Forbidden - Requires admin role"
// @Failure 500 {object} ErrorResponse "Internal Server Error - Failed to create concert or offload seat creation"
// @Router /admin/concerts [post]
func (ctrl *ConcertController) CreateConcert(c *gin.Context) {
	var req models.CreateConcertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Invalid JSON body for create concert: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for create concert: %v", validationErrors)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: utils.FormatValidationErrors(validationErrors)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	resp, err := ctrl.ConcertService.CreateConcert(ctx, &req)
	if err != nil {
		utils.LogError("Failed to create concert: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create concert: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// @Summary Get all concerts
// @Description Retrieve a list of all available concerts.
// @Tags Concerts
// @Produce json
// @Success 200 {array} models.ConcertResponse
// @Failure 500 {object} ErrorResponse "Internal Server Error - Failed to retrieve concerts"
// @Router /concerts [get]
func (ctrl *ConcertController) GetConcerts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	concerts, err := ctrl.ConcertService.GetConcerts(ctx)
	if err != nil {
		utils.LogError("Failed to get concerts: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve concerts: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, concerts)
}

// @Summary Get concert by ID
// @Description Retrieve details of a specific concert by its ID.
// @Tags Concerts
// @Produce json
// @Param id path int true "Concert ID"
// @Success 200 {object} models.ConcertResponse
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid concert ID"
// @Failure 404 {object} ErrorResponse "Not Found - Concert not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error - Failed to retrieve concert"
// @Router /concerts/{id} [get]
func (ctrl *ConcertController) GetConcertByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid concert ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := ctrl.ConcertService.GetConcertByID(ctx, uint(id))
	if err != nil {
		utils.LogError("Failed to get concert ID %d: %v", id, err)
		if err.Error() == "concert not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve concert: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get seats for a concert
// @Description Retrieves the status of all seats for a specific concert.
// @Tags Concerts
// @Produce json
// @Param id path int true "Concert ID"
// @Success 200 {array} models.SeatResponse
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid concert ID"
// @Failure 404 {object} ErrorResponse "Not Found - Concert not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error - Failed to retrieve seats"
// @Router /concerts/{id}/seats [get]
func (ctrl *ConcertController) GetConcertSeats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid concert ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	seats, err := ctrl.ConcertService.GetSeatsForConcert(ctx, uint(id))
	if err != nil {
		utils.LogError("Failed to get seats for concert ID %d: %v", id, err)
		if err.Error() == "concert not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve seats: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, seats)
}
