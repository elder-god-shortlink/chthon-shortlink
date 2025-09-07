package service

import (
	"errors"
	"time"

	"github.com/chthon/shortlink/pkg/auth"
	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/services/user-management/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo       *repository.UserRepository
	config     *config.Config
	jwtManager *auth.JWTManager
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	User      *models.User `json:"user"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

func NewUserService(repo *repository.UserRepository, config *config.Config) *UserService {
	jwtManager := auth.NewJWTManager(&config.JWT)

	return &UserService{
		repo:       repo,
		config:     config,
		jwtManager: jwtManager,
	}
}

func (s *UserService) Register(req *RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to check existing user", "error", err)
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Create user
	user := &models.User{
		Email:    req.Email,
		Role:     role,
		IsActive: true,
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return nil, err
	}
	user.PasswordHash = string(hashedPassword)

	if err := s.repo.CreateUser(user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	logger.Info("User registered successfully", "email", user.Email)
	return user, nil
}

func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		logger.Error("Failed to get user", "error", err)
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return nil, err
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(s.config.JWT.ExpiresIn)

	logger.Info("User logged in successfully", "email", user.Email)

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		logger.Error("Failed to get user", "error", err)
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		logger.Error("Failed to get user", "error", err)
		return nil, err
	}

	// Update fields if provided
	if req.Email != "" {
		// Check if email is already taken by another user
		existingUser, err := s.repo.GetUserByEmail(req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("Failed to check existing email", "error", err)
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already taken")
		}
		user.Email = req.Email
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateUser(user); err != nil {
		logger.Error("Failed to update user", "error", err)
		return nil, err
	}

	logger.Info("User updated successfully", "id", id)
	return user, nil
}

func (s *UserService) DeleteUser(id uint) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		logger.Error("Failed to get user", "error", err)
		return err
	}

	if err := s.repo.DeleteUser(id); err != nil {
		logger.Error("Failed to delete user", "error", err)
		return err
	}

	logger.Info("User deleted successfully", "id", id, "email", user.Email)
	return nil
}

func (s *UserService) GetAllUsers(page, limit int) ([]models.User, int64, error) {
	offset := (page - 1) * limit

	users, err := s.repo.GetAllUsers(limit, offset)
	if err != nil {
		logger.Error("Failed to get users", "error", err)
		return nil, 0, err
	}

	total, err := s.repo.CountUsers()
	if err != nil {
		logger.Error("Failed to count users", "error", err)
		return nil, 0, err
	}

	return users, total, nil
}
