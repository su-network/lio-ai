package services

import (
	"errors"
	"fmt"
	"log"
	"lio-ai/internal/auth"
	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUnauthorized       = errors.New("user is not authorized to perform this action")
	ErrNotFound           = errors.New("resource not found")
)

// UserService handles user-related business logic
type UserService struct {
	repo       *repositories.UserRepository
	jwtManager *auth.JWTManager
}

// NewUserService creates a new user service
func NewUserService(repo *repositories.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *UserService) Register(username, email, password, fullName string) (*models.User, error) {
	// Validate password
	if err := auth.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create user
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         "user",
		IsActive:     true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns JWT token
func (s *UserService) Login(email, password string) (string, *models.User, error) {
	log.Printf("🔍 Login attempt for: %s", email)
	
	// Find user by email
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		log.Printf("❌ Login: GetByEmail failed: %v", err)
		return "", nil, errors.New("authentication failed")
	}

	if user == nil {
		log.Printf("❌ Login: User not found")
		return "", nil, ErrInvalidCredentials
	}

	log.Printf("✓ Login: User found (ID: %d, active: %v)", user.ID, user.IsActive)

	if !user.IsActive {
		log.Printf("❌ Login: User is not active")
		return "", nil, ErrUserInactive
	}

	log.Printf("🔍 Login: Verifying password (hash: %s...)", user.PasswordHash[:20])
	
	// Verify password
	if err := s.repo.VerifyPassword(user, password); err != nil {
		log.Printf("❌ Login: Password verification failed: %v", err)
		return "", nil, ErrInvalidCredentials
	}

	log.Printf("✓ Login: Password verified successfully")

	// Update last login
	_ = s.repo.UpdateLastLogin(user.ID)

	// Generate JWT token (24-hour expiration)
	// Use string conversion of user.ID as the subject
	token, err := s.jwtManager.GenerateToken(
		fmt.Sprintf("%d", user.ID),
		user.Email,
		[]string{user.Role},
		24*time.Hour,
	)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(username)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id int64) (*models.User, error) {
	return s.repo.GetByID(id)
}

// ChangePassword changes user's password
func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := s.repo.VerifyPassword(user, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := auth.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to process password")
	}

	// Update password in database
	return s.repo.UpdatePassword(userID, hash)
}

// GenerateTokenForUser generates a JWT token for a user
func (s *UserService) GenerateTokenForUser(user *models.User) (string, error) {
	if user == nil {
		return "", errors.New("user cannot be nil")
	}
	
	// Generate JWT token with 24-hour expiration
	// Use string conversion of user.ID as the subject
	token, err := s.jwtManager.GenerateToken(
		fmt.Sprintf("%d", user.ID),
		user.Email,
		[]string{user.Role},
		24*time.Hour,
	)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	
	return token, nil
}
