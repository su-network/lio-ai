package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	user, err := h.userService.Register(req.Username, req.Email, req.Password, req.FullName)
	if err != nil {
		// Log the detailed error securely
		log.Printf("[AUTH] Registration failed for %s: %v", req.Email, err)

		// Return generic error to client
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "registration failed",
			"code":  "REGISTRATION_FAILED",
		})
		return
	}

	// Generate JWT token for immediate login after registration
	token, err := h.userService.GenerateTokenForUser(user)
	if err != nil {
		log.Printf("[AUTH] Token generation failed for newly registered user %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "registration succeeded but login failed",
			"code":  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// Log successful registration
	log.Printf("[AUDIT] User registered: %s (ID: %d)", user.Email, user.ID)

	// Set cookie for immediate persistence
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",
		token,
		86400, // 24 hours
		"/",
		"",
		true,  // httpOnly
		false, // secure (false for development)
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.FullName,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	token, user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		// Log failed login attempt
		log.Printf("[AUDIT] Login failed for %s: %v (IP: %s)", req.Email, err, c.ClientIP())

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication failed",
			"code":  "INVALID_CREDENTIALS",
		})
		return
	}

	// Log successful login
	log.Printf("[AUDIT] Login successful: %s (ID: %d, IP: %s)", user.Email, user.ID, c.ClientIP())

	// Set cookie with JWT token for persistence across page refreshes
	// httpOnly=true prevents XSS attacks, secure=false for local development
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"auth_token",
		token,
		86400, // 24 hours
		"/",
		"",      // domain (empty = current domain)
		true,   // httpOnly (prevents JavaScript access for security)
		false,  // secure (false for development, true for production)
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.FullName,
			"role":     user.Role,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract user from JWT (set by middleware)
	userID, exists := c.Get("user_id")
	if exists {
		log.Printf("[AUDIT] Logout: %s", userID)
	}

	// Clear authentication cookie
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user from JWT token (set by middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Get user details
	user, err := h.userService.GetUserByUsername(userIDStr.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	if err := h.userService.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		log.Printf("[AUDIT] Password change failed for user %s: %v", user.Email, err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "password change failed",
			"code":  "PASSWORD_CHANGE_FAILED",
		})
		return
	}

	// Log successful password change
	log.Printf("[AUDIT] Password changed: %s (ID: %d)", user.Email, user.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GetProfile returns current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// user_id from JWT is a string representation of the ID
	userIDStr, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user id format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	// Convert string to int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user id format",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
			"code":  "USER_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.FullName,
		"role":     user.Role,
	})
}
