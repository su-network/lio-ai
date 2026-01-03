package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/auth"
	"lio-ai/internal/config"
	"lio-ai/internal/db"
	"lio-ai/internal/handlers"
	"lio-ai/internal/middleware"
	"lio-ai/internal/repositories"
	"lio-ai/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize JWT manager (must happen before handlers)
	jwtManager, err := auth.NewJWTManager()
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.ErrorRecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// SECURITY: Add JWT auth middleware
	router.Use(middleware.NewAuthMiddleware(jwtManager))

	// SECURITY: Add CSRF protection middleware
	router.Use(middleware.CSRFMiddleware())

	// Rate limiting middleware
	limiter := middleware.NewRateLimiter()
	router.Use(middleware.RateLimitMiddleware(limiter))

	// Initialize repositories
	userRepo := repositories.NewUserRepository(database.GetConnection())
	docRepo := repositories.NewDocumentRepository(database.GetConnection())
	chatRepo := repositories.NewChatRepository(database.GetConnection())
	usageRepo := repositories.NewUsageRepository(database.GetConnection())
	providerKeyRepo := repositories.NewProviderKeyRepository(database.GetConnection())
	
	// Initialize services
	userService := services.NewUserService(userRepo, jwtManager)
	docService := services.NewDocumentService(docRepo)
	chatService := services.NewChatService(chatRepo)
	usageService := services.NewUsageService(usageRepo)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	docHandler := handlers.NewDocumentHandler(docService)
	chatHandler := handlers.NewChatHandler(chatService)
	usageHandler := handlers.NewUsageHandler(usageService)
	systemHandler := handlers.NewSystemHandler(database.GetConnection())
	providerKeyHandler := handlers.NewProviderKeyHandler(providerKeyRepo)

	// Initialize proxy handler for FastAPI backend
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}
	proxyHandler := handlers.NewProxyHandler(backendURL)

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Lio AI Gateway (Secured)",
			"version": cfg.App.Version,
			"status":  "operational",
			"security": "jwt-enabled csrf-protected",
		})
	})

	// Health check with backend verification
	router.GET("/health", systemHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		// SECURITY: Authentication routes (NO JWT required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.RequireAuth(), authHandler.Logout)
			auth.GET("/profile", middleware.RequireAuth(), authHandler.GetProfile)
		}

		// Document routes (JWT required)
		documents := api.Group("/documents")
		documents.Use(middleware.RequireAuth())
		{
			documents.POST("", docHandler.CreateDocument)
			documents.GET("", docHandler.GetDocuments)
			documents.GET("/:id", docHandler.GetDocument)
			documents.PUT("/:id", docHandler.UpdateDocument)
			documents.DELETE("/:id", docHandler.DeleteDocument)
		}

		// Chat routes (JWT required)
		chats := api.Group("/chats")
		chats.Use(middleware.RequireAuth())
		{
			chats.POST("", chatHandler.CreateChat)
			chats.GET("", chatHandler.GetUserChats)
			chats.GET("/:id", chatHandler.GetChat)
			chats.PUT("/:id", chatHandler.UpdateChat)
			chats.DELETE("/:id", chatHandler.DeleteChat)
			chats.POST("/:id/messages", chatHandler.SendMessage)
			chats.GET("/:id/messages", chatHandler.GetMessages)
			
			// UUID-based routes
			chats.GET("/uuid/:uuid", chatHandler.GetChatByUUID)
			chats.POST("/uuid/:uuid/messages", chatHandler.SendMessageByUUID)
			chats.GET("/uuid/:uuid/messages", chatHandler.GetMessagesByUUID)
		}

		// Chat completion endpoint (JWT required)
		api.POST("/chat/completions", middleware.RequireAuth(), chatHandler.ChatCompletion)

		// Usage routes (JWT required)
		usage := api.Group("/usage")
		usage.Use(middleware.RequireAuth())
		{
			usage.GET("/quota", usageHandler.GetQuotaStatus)
			usage.GET("/summary", usageHandler.GetUsageSummary)
			usage.POST("/track", usageHandler.TrackUsage)
			usage.POST("/check-quota", usageHandler.CheckQuota)
			usage.GET("/dashboard", usageHandler.GetDashboard)
		}

		// System routes (JWT required)
		system := api.Group("/system")
		system.Use(middleware.RequireAuth())
		{
			system.GET("/metrics", systemHandler.GetMetrics)
			system.GET("/info", systemHandler.GetInfo)
			system.GET("/stats", systemHandler.GetStats)
		}

		// Provider API Key routes (JWT required)
		apiKeys := api.Group("/api-keys")
		apiKeys.Use(middleware.RequireAuth())
		{
			apiKeys.GET("", providerKeyHandler.GetAllKeys)
			apiKeys.POST("", providerKeyHandler.CreateOrUpdateKey)
			apiKeys.POST("/sync", providerKeyHandler.SyncAllKeys)
			apiKeys.DELETE("/:provider", providerKeyHandler.DeleteKey)
			apiKeys.GET("/:provider", providerKeyHandler.GetProviderKey)
		}
	}

	// Proxy routes for code generation service (JWT required)
	codeGen := router.Group("/api/v1/codegen")
	codeGen.Use(middleware.RequireAuth())
	{
		codeGen.POST("/generate", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		codeGen.POST("/validate", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		codeGen.POST("/rag/search", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
	}

	// Stats endpoint (JWT required)
	router.GET("/api/v1/stats", middleware.RequireAuth(), func(c *gin.Context) {
		proxyHandler.ProxyRequest(c)
	})

	// Proxy routes for model management (JWT required)
	models := router.Group("/api/v1/models")
	models.Use(middleware.RequireAuth())
	{
		models.GET("", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		models.GET("/status", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		models.GET("/:model_id", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		models.POST("/:model_id/health", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		models.GET("/recommend", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
		models.POST("/recommend", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
	}

	// Proxy all unmatched routes to backend
	router.NoRoute(func(c *gin.Context) {
		proxyHandler.ProxyRequest(c)
	})

	// Build server address
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	// Start server
	log.Printf("✓ Starting Go Gateway at http://%s", addr)
	log.Printf("✓ Python AI Service: http://localhost:%s", cfg.Backend.AIServicePort)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
