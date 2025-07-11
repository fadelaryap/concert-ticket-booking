package services

import (
	"errors"
	"time"

	"backend/user-service/models"
	"backend/user-service/repositories"
	"backend/user-service/utils"

	"gorm.io/gorm"
)

type UserService struct {
	UserRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{UserRepo: userRepo}
}

func (s *UserService) RegisterUser(req *models.UserRegisterRequest) (*models.UserResponse, error) {

	_, errUsername := s.UserRepo.FindUserByUsername(req.Username)
	if errUsername == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(errUsername, gorm.ErrRecordNotFound) {
		return nil, errUsername
	}

	_, errEmail := s.UserRepo.FindUserByEmail(req.Email)
	if errEmail == nil {
		return nil, errors.New("email already exists")
	}
	if !errors.Is(errEmail, gorm.ErrRecordNotFound) {
		return nil, errEmail
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.LogError("Failed to hash password during registration: %v", err)
		return nil, errors.New("failed to process password")
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := s.UserRepo.CreateUser(user); err != nil {
		utils.LogError("Failed to create user in database: %v", err)
		return nil, errors.New("failed to register user")
	}

	response := user.ToUserResponse()
	return &response, nil
}

func (s *UserService) LoginUser(req *models.UserLoginRequest) (string, *models.UserResponse, error) {
	user, err := s.UserRepo.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid credentials")
		}
		utils.LogError("Database error finding user '%s' during login: %v", req.Username, err)
		return "", nil, errors.New("internal server error during login")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return "", nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLogin = &now
	if err := s.UserRepo.UpdateUser(user); err != nil {

		utils.LogWarning("Failed to update last login for user %s (ID: %d): %v", user.Username, user.ID, err)
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		utils.LogError("Failed to generate JWT for user %s (ID: %d): %v", user.Username, user.ID, err)
		return "", nil, errors.New("failed to authenticate user")
	}

	response := user.ToUserResponse()
	return token, &response, nil
}

func (s *UserService) GetUserProfile(userID uint) (*models.UserResponse, error) {
	user, err := s.UserRepo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		utils.LogError("Database error fetching user profile for ID %d: %v", userID, err)
		return nil, errors.New("internal server error fetching profile")
	}
	response := user.ToUserResponse()
	return &response, nil
}

func (s *UserService) UpdateUserProfile(userID uint, updatedData *models.UserResponse) (*models.UserResponse, error) {
	user, err := s.UserRepo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		utils.LogError("Database error finding user for update ID %d: %v", userID, err)
		return nil, errors.New("internal server error updating profile")
	}

	if updatedData.ID != userID {
		utils.LogWarning("Unauthorized attempt to update profile with mismatched user ID. Auth ID: %d, Request ID: %d", userID, updatedData.ID)
		return nil, errors.New("unauthorized: cannot update another user's profile")
	}

	if updatedData.Email != "" && updatedData.Email != user.Email {

		existingUserWithEmail, errEmail := s.UserRepo.FindUserByEmail(updatedData.Email)
		if errEmail == nil && existingUserWithEmail.ID != user.ID {
			return nil, errors.New("email already taken by another user")
		}
		if !errors.Is(errEmail, gorm.ErrRecordNotFound) {
			return nil, errEmail
		}
		user.Email = updatedData.Email
	}

	if err := s.UserRepo.UpdateUser(user); err != nil {
		utils.LogError("Failed to update user profile for ID %d: %v", userID, err)
		return nil, errors.New("failed to update profile")
	}

	response := user.ToUserResponse()
	return &response, nil
}
