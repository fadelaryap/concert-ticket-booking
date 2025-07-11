package controllers

import (
	"net/http"
	"strings"
	"time"

	"backend/user-service/models"
	"backend/user-service/services"
	"backend/user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserController struct {
	UserService *services.UserService
	Validate    *validator.Validate
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		UserService: userService,
		Validate:    validator.New(),
	}
}

// @Summary Register a new user
// @Description Register a new user with username, email, and password.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.UserRegisterRequest true "User registration data"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} map[string]string "Bad Request - Invalid input or validation errors"
// @Failure 409 {object} map[string]string "Conflict - Username or email already exists"
// @Failure 500 {object} map[string]string "Internal Server Error - Failed to process request"
// @Router /register [post]
func (ctrl *UserController) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Invalid JSON body for registration: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for registration: %v", validationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(validationErrors)})
		return
	}

	userResponse, err := ctrl.UserService.RegisterUser(&req)
	if err != nil {
		utils.LogError("Failed to register user: %v", err)
		if strings.Contains(err.Error(), "username already exists") || strings.Contains(err.Error(), "email already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, userResponse)
}

// @Summary Log in a user
// @Description Authenticate a user and set a JWT token as HttpOnly cookie.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.UserLoginRequest true "User login credentials"
// @Success 200 {object} map[string]string "message: Login successful"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Invalid JSON body for login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for login: %v", validationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(validationErrors)})
		return
	}

	token, userResponse, err := ctrl.UserService.LoginUser(&req)
	if err != nil {
		utils.LogError("Failed to login user %s: %v", req.Username, err)
		if err.Error() == "invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login: " + err.Error()})
		return
	}

	c.SetCookie("token", token, int((time.Hour * 24).Seconds()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    userResponse,
	})
}

// @Summary Get user profile
// @Description Retrieve the profile of the authenticated user.
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} map[string]string "Unauthorized - Missing or invalid token"
// @Failure 404 {object} map[string]string "Not Found - User not found (shouldn't happen for authenticated user)"
// @Failure 500 {object} map[string]string "Internal Server Error - Failed to retrieve profile"
// @Router /profile [get]
func (ctrl *UserController) GetProfile(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("User ID not found in context after auth middleware (possible internal error)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	profile, err := ctrl.UserService.GetUserProfile(userID.(uint))
	if err != nil {
		utils.LogError("Failed to get user profile for ID %d: %v", userID, err)
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// @Summary Update user profile
// @Description Update the email of the authenticated user.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body models.UserResponse true "User profile data to update (only email is currently allowed)"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]string "Bad Request - Invalid input or validation errors"
// @Failure 401 {object} map[string]string "Unauthorized - Missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - Attempt to update another user's profile"
// @Failure 404 {object} map[string]string "Not Found - User not found (shouldn't happen for authenticated user)"
// @Failure 409 {object} map[string]string "Conflict - Email already taken by another user"
// @Failure 500 {object} map[string]string "Internal Server Error - Failed to update profile"
// @Router /profile [put]
func (ctrl *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.LogError("User ID not found in context after auth middleware (possible internal error)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	var updatedUserReq models.UserResponse
	if err := c.ShouldBindJSON(&updatedUserReq); err != nil {
		utils.LogError("Invalid JSON body for profile update: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := ctrl.Validate.StructPartial(updatedUserReq, "Email"); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.LogError("Validation error for profile update: %v", validationErrors)
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(validationErrors)})
		return
	}

	profile, err := ctrl.UserService.UpdateUserProfile(userID.(uint), &updatedUserReq)
	if err != nil {
		utils.LogError("Failed to update user profile for ID %d: %v", userID, err)
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "unauthorized: cannot update another user's profile":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "email already taken by another user":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

// @Summary Log out a user
// @Description Clears the JWT token HttpOnly cookie.
// @Tags Users
// @Produce json
// @Success 200 {object} map[string]string "message: Logout successful"
// @Router /logout [post]
func (ctrl *UserController) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
