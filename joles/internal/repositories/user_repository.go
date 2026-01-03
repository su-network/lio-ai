package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"lio-ai/internal/auth"
	"lio-ai/internal/models"
	"time"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
		now,
		now,
	)

	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.email" {
			return errors.New("email already registered")
		}
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return errors.New("username already taken")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		user.ID = id
	}

	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE email = ? AND is_active = 1
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE username = ? AND is_active = 1
	`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		WHERE id = ? AND is_active = 1
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// VerifyPassword checks if password matches user's hash
func (r *UserRepository) VerifyPassword(user *models.User, password string) error {
	return auth.CheckPassword(password, user.PasswordHash)
}

// UpdatePassword updates user's password hash
func (r *UserRepository) UpdatePassword(userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// UpdateLastLogin updates user's last login time
func (r *UserRepository) UpdateLastLogin(userID int64) error {
	query := `UPDATE users SET updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}
