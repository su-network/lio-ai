package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims with user information
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTManager manages JWT token generation and validation
type JWTManager struct {
	secretKey string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager() (*JWTManager, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET_KEY environment variable not set")
	}

	if len(secretKey) < 32 {
		return nil, errors.New("JWT_SECRET_KEY must be at least 32 characters")
	}

	return &JWTManager{secretKey: secretKey}, nil
}

// GenerateToken creates a new JWT token
func (jm *JWTManager) GenerateToken(userID, email string, roles []string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jm.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken creates a new token from existing claims
func (jm *JWTManager) RefreshToken(claims *Claims, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(expiresIn))
	claims.IssuedAt = jwt.NewNumericDate(now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
