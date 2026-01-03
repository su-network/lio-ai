package models

import "time"

// User represents a user in the system
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name,omitempty"`
	APIKey       string    `json:"api_key,omitempty"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	IsActive     bool      `json:"is_active"`
	Role         string    `json:"role"` // "admin", "user", "developer"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

// UserProfile represents user profile information
type UserProfile struct {
	User         *User         `json:"user"`
	QuotaStatus  *QuotaStatus  `json:"quota_status"`
	TotalChats   int           `json:"total_chats"`
	TotalDocs    int           `json:"total_documents"`
	UsageSummary *UsageSummary `json:"usage_summary"`
}

// ProviderAPIKey represents a user's API key for an LLM provider
type ProviderAPIKey struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	Provider        string    `json:"provider"` // openai, anthropic, google, cohere
	APIKeyEncrypted string    `json:"-"`        // Never expose in JSON
	APIKey          string    `json:"api_key,omitempty"` // Only for create/update
	ModelsEnabled   string    `json:"models_enabled,omitempty"` // JSON array of model IDs
	IsActive        bool      `json:"is_active"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProviderAPIKeyRequest represents a request to add/update provider API key
type ProviderAPIKeyRequest struct {
	Provider      string   `json:"provider" binding:"required"`
	APIKey        string   `json:"api_key" binding:"required"`
	ModelsEnabled []string `json:"models_enabled,omitempty"`
}

// ProviderAPIKeyResponse represents the response (without sensitive data)
type ProviderAPIKeyResponse struct {
	ID            int64      `json:"id"`
	Provider      string     `json:"provider"`
	ModelsEnabled []string   `json:"models_enabled"`
	IsActive      bool       `json:"is_active"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	HasKey        bool       `json:"has_key"` // Indicates if key is set
}
