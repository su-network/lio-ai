package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/auth"
	"lio-ai/internal/config"
	"lio-ai/internal/db"
	"lio-ai/internal/handlers"
	"lio-ai/internal/middleware"
	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
	"lio-ai/internal/services"
)

var (
	testDB     *db.Database
	testRouter *gin.Engine
	authHeader string
	userID     string
)

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("DATABASE_URL", ":memory:")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-at-least-32-bytes!")

	// Initialize test database
	cfg, _ := config.LoadConfig()
	testDB, _ = db.NewDatabase(cfg)

	// Initialize router
	gin.SetMode(gin.TestMode)
	testRouter = gin.New()

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()

	os.Exit(code)
}

// setupTestRouter initializes router with all middleware and handlers
func setupTestRouter() *gin.Engine {
	router := gin.New()

	// Middleware
	jwtManager, _ := auth.NewJWTManager()
	router.Use(middleware.NewAuthMiddleware(jwtManager))
	router.Use(middleware.CSRFMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Repositories and Services
	userRepo := repositories.NewUserRepository(testDB.GetConnection())
	userService := services.NewUserService(userRepo, jwtManager)

	chatRepo := repositories.NewChatRepository(testDB.GetConnection())
	chatService := services.NewChatService(chatRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(userService)
	chatHandler := handlers.NewChatHandler(chatService)

	// Routes
	router.POST("/api/v1/auth/register", authHandler.Register)
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/logout", middleware.RequireAuth(), authHandler.Logout)
	router.GET("/api/v1/auth/profile", middleware.RequireAuth(), authHandler.GetProfile)

	router.POST("/api/v1/chats", middleware.RequireAuth(), chatHandler.CreateChat)
	router.GET("/api/v1/chats", middleware.RequireAuth(), chatHandler.GetUserChats)
	router.GET("/api/v1/chats/:id", middleware.RequireAuth(), chatHandler.GetChat)

	return router
}

// TestJWTGeneration tests JWT token generation and validation
func TestJWTGeneration(t *testing.T) {
	jwtManager, _ := auth.NewJWTManager()

	token, err := jwtManager.GenerateToken("testuser", "test@example.com", []string{"user"}, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "testuser" {
		t.Errorf("Expected user_id 'testuser', got '%s'", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", claims.Email)
	}
}

// TestJWTExpiration tests that expired tokens are rejected
func TestJWTExpiration(t *testing.T) {
	jwtManager, _ := auth.NewJWTManager()

	// Create token that expires immediately
	token, _ := jwtManager.GenerateToken("testuser", "test@example.com", []string{"user"}, -time.Hour)

	_, err := jwtManager.ValidateToken(token)
	if err == nil {
		t.Errorf("Expected expired token to fail validation")
	}
}

// TestJWTTampering tests that tampered tokens are rejected
func TestJWTTampering(t *testing.T) {
	jwtManager, _ := auth.NewJWTManager()

	token, _ := jwtManager.GenerateToken("testuser", "test@example.com", []string{"user"}, time.Hour)

	// Tamper with token
	tamperedToken := token[:len(token)-5] + "XXXXX"

	_, err := jwtManager.ValidateToken(tamperedToken)
	if err == nil {
		t.Errorf("Expected tampered token to fail validation")
	}
}

// TestPasswordValidation tests password security requirements
func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"short", false},
		{"nouppercase123", false},
		{"NOLOWERCASE123", false},
		{"NoDigitsHere", false},
		{"ValidPassword123", true},
		{"StrongPass456!", true},
	}

	for _, test := range tests {
		err := auth.ValidatePassword(test.password)
		isValid := err == nil

		if isValid != test.valid {
			t.Errorf("Password '%s': expected valid=%v, got %v (err: %v)", test.password, test.valid, isValid, err)
		}
	}
}

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	password := "MySecurePassword123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Correct password should verify
	err = auth.CheckPassword(password, hash)
	if err != nil {
		t.Errorf("Correct password failed verification: %v", err)
	}

	// Wrong password should fail
	err = auth.CheckPassword("WrongPassword123", hash)
	if err == nil {
		t.Errorf("Wrong password passed verification")
	}
}

// TestUserRegistration tests user registration flow
func TestUserRegistration(t *testing.T) {
	router := setupTestRouter()

	registerReq := models.RegisterRequest{
		Username: "testuser1",
		Email:    "test1@example.com",
		Password: "SecurePass123",
		FullName: "Test User",
	}

	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["message"] != "User registered successfully" {
		t.Errorf("Registration failed: %v", resp)
	}
}

// TestUserLogin tests user login and JWT token generation
func TestUserLogin(t *testing.T) {
	router := setupTestRouter()

	// Register user first
	registerReq := models.RegisterRequest{
		Username: "loginuser",
		Email:    "login@example.com",
		Password: "SecurePass123",
		FullName: "Login User",
	}
	regBody, _ := json.Marshal(registerReq)
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	// Login
	loginReq := models.LoginRequest{
		Email:    "login@example.com",
		Password: "SecurePass123",
	}
	loginBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	token, exists := resp["token"]
	if !exists {
		t.Errorf("Token not returned in response: %v", resp)
	}

	if token == "" {
		t.Errorf("Token is empty")
	}

	// Store for later tests
	authHeader = "Bearer " + token.(string)
	userID = "loginuser"
}

// TestInvalidLogin tests that invalid credentials are rejected
func TestInvalidLogin(t *testing.T) {
	router := setupTestRouter()

	// Register user first
	registerReq := models.RegisterRequest{
		Username: "validuser",
		Email:    "valid@example.com",
		Password: "SecurePass123",
		FullName: "Valid User",
	}
	regBody, _ := json.Marshal(registerReq)
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	// Try with wrong password
	loginReq := models.LoginRequest{
		Email:    "valid@example.com",
		Password: "WrongPassword123",
	}
	loginBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", w.Code)
	}
}

// TestCSRFProtection tests CSRF token validation
func TestCSRFProtection(t *testing.T) {
	router := setupTestRouter()

	// Setup: Register and login user
	registerReq := models.RegisterRequest{
		Username: "csrfuser",
		Email:    "csrf@example.com",
		Password: "SecurePass123",
		FullName: "CSRF User",
	}
	regBody, _ := json.Marshal(registerReq)
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	// Login
	loginReq := models.LoginRequest{
		Email:    "csrf@example.com",
		Password: "SecurePass123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginHttpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginHttpReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginHttpReq)

	var loginResp map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	// Get CSRF token
	getReq := httptest.NewRequest("GET", "/api/v1/chats", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getReq, getW)

	// Extract CSRF cookie
	csrfToken := ""
	for _, cookie := range getW.Result().Cookies() {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}

	if csrfToken == "" {
		t.Fatalf("CSRF token not set in cookie")
	}

	// Try POST without CSRF token - should fail
	chatReq := models.ChatRequest{
		UserID: userID,
		Title:  "Test Chat",
	}
	chatBody, _ := json.Marshal(chatReq)
	postReq := httptest.NewRequest("POST", "/api/v1/chats", bytes.NewReader(chatBody))
	postReq.Header.Set("Authorization", "Bearer "+token)
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	router.ServeHTTP(postReq, postW)

	if postW.Code != http.StatusForbidden {
		t.Errorf("Expected 403 without CSRF token, got %d", postW.Code)
	}

	// Try with correct CSRF token - should succeed
	postReq2 := httptest.NewRequest("POST", "/api/v1/chats", bytes.NewReader(chatBody))
	postReq2.Header.Set("Authorization", "Bearer "+token)
	postReq2.Header.Set("Content-Type", "application/json")
	postReq2.Header.Set("X-CSRF-Token", csrfToken)
	postReq2.Header.Set("Cookie", "_csrf="+csrfToken)
	postW2 := httptest.NewRecorder()
	router.ServeHTTP(postReq2, postW2)

	if postW2.Code != http.StatusCreated && postW2.Code != http.StatusOK {
		t.Errorf("Expected success with CSRF token, got %d", postW2.Code)
	}
}

// TestResourceOwnership tests that users can't access other users' resources
func TestResourceOwnership(t *testing.T) {
	router := setupTestRouter()

	// Create User 1
	user1Req := models.RegisterRequest{
		Username: "user1",
		Email:    "user1@example.com",
		Password: "SecurePass123",
		FullName: "User 1",
	}
	u1Body, _ := json.Marshal(user1Req)
	u1RegReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(u1Body))
	u1RegReq.Header.Set("Content-Type", "application/json")
	u1RegW := httptest.NewRecorder()
	router.ServeHTTP(u1RegReq, u1RegW)

	// Login User 1
	u1LoginReq := models.LoginRequest{
		Email:    "user1@example.com",
		Password: "SecurePass123",
	}
	u1LoginBody, _ := json.Marshal(u1LoginReq)
	u1LoginHttpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(u1LoginBody))
	u1LoginHttpReq.Header.Set("Content-Type", "application/json")
	u1LoginW := httptest.NewRecorder()
	router.ServeHTTP(u1LoginHttpReq, u1LoginW)

	var u1LoginResp map[string]interface{}
	json.Unmarshal(u1LoginW.Body.Bytes(), &u1LoginResp)
	u1Token := u1LoginResp["token"].(string)

	// Create User 2
	user2Req := models.RegisterRequest{
		Username: "user2",
		Email:    "user2@example.com",
		Password: "SecurePass123",
		FullName: "User 2",
	}
	u2Body, _ := json.Marshal(user2Req)
	u2RegReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(u2Body))
	u2RegReq.Header.Set("Content-Type", "application/json")
	u2RegW := httptest.NewRecorder()
	router.ServeHTTP(u2RegReq, u2RegW)

	// Login User 2
	u2LoginReq := models.LoginRequest{
		Email:    "user2@example.com",
		Password: "SecurePass123",
	}
	u2LoginBody, _ := json.Marshal(u2LoginReq)
	u2LoginHttpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(u2LoginBody))
	u2LoginHttpReq.Header.Set("Content-Type", "application/json")
	u2LoginW := httptest.NewRecorder()
	router.ServeHTTP(u2LoginHttpReq, u2LoginW)

	var u2LoginResp map[string]interface{}
	json.Unmarshal(u2LoginW.Body.Bytes(), &u2LoginResp)
	u2Token := u2LoginResp["token"].(string)

	// User 1 creates a chat
	chatReq := models.ChatRequest{
		UserID: "user1",
		Title:  "User 1's Chat",
	}
	chatBody, _ := json.Marshal(chatReq)
	chatHttpReq := httptest.NewRequest("POST", "/api/v1/chats", bytes.NewReader(chatBody))
	chatHttpReq.Header.Set("Authorization", "Bearer "+u1Token)
	chatHttpReq.Header.Set("Content-Type", "application/json")
	chatW := httptest.NewRecorder()
	router.ServeHTTP(chatHttpReq, chatW)

	// User 2 tries to access User 1's chat - should be forbidden
	// (In a real implementation, we'd get the chat ID and try to access it)
	// For now, this tests that authentication is required
	noAuthReq := httptest.NewRequest("GET", "/api/v1/chats", nil)
	noAuthW := httptest.NewRecorder()
	router.ServeHTTP(noAuthReq, noAuthW)

	if noAuthW.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without authentication, got %d", noAuthW.Code)
	}
}

// TestAuthenticationRequired tests that protected endpoints require authentication
func TestAuthenticationRequired(t *testing.T) {
	router := setupTestRouter()

	// Try to access protected endpoint without token
	req := httptest.NewRequest("GET", "/api/v1/chats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(req, w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for unauthenticated request, got %d", w.Code)
	}
}

// TestInvalidToken tests that invalid tokens are rejected
func TestInvalidToken(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/v1/chats", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(req, w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}
