package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	// Rate limiting middleware
	limiter := middleware.NewRateLimiter()
	router.Use(middleware.RateLimitMiddleware(limiter))

	// Initialize repositories, services, and handlers
	docRepo := repositories.NewDocumentRepository(database.GetConnection())
	chatRepo := repositories.NewChatRepository(database.GetConnection())
	usageRepo := repositories.NewUsageRepository(database.GetConnection())
	providerKeyRepo := repositories.NewProviderKeyRepository(database.GetConnection())
	
	docService := services.NewDocumentService(docRepo)
	chatService := services.NewChatService(chatRepo)
	usageService := services.NewUsageService(usageRepo)
	
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
			"message": "Welcome to Lio AI Gateway",
			"version": cfg.App.Version,
			"status":  "operational",
		})
	})

	// Health check with backend verification
	router.GET("/health", systemHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		// Document routes
		documents := api.Group("/documents")
		{
			documents.POST("", docHandler.CreateDocument)
			documents.GET("", docHandler.GetDocuments)
			documents.GET("/:id", docHandler.GetDocument)
			documents.PUT("/:id", docHandler.UpdateDocument)
			documents.DELETE("/:id", docHandler.DeleteDocument)
		}

		// Chat routes
		chats := api.Group("/chats")
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

		// Usage routes
		usage := api.Group("/usage")
		{
			usage.GET("/quota", usageHandler.GetQuotaStatus)
			usage.GET("/summary", usageHandler.GetUsageSummary)
			usage.POST("/track", usageHandler.TrackUsage)
			usage.POST("/check-quota", usageHandler.CheckQuota)
			usage.GET("/dashboard", usageHandler.GetDashboard)
		}

		// System routes
		system := api.Group("/system")
		{
			system.GET("/metrics", systemHandler.GetMetrics)
			system.GET("/info", systemHandler.GetInfo)
			system.GET("/stats", systemHandler.GetStats)
		}

		// Provider API Key routes
		apiKeys := api.Group("/api-keys")
		{
			apiKeys.GET("", providerKeyHandler.GetAllKeys)
			apiKeys.POST("", providerKeyHandler.CreateOrUpdateKey)
			apiKeys.DELETE("/:provider", providerKeyHandler.DeleteKey)
			apiKeys.GET("/:provider", providerKeyHandler.GetProviderKey)
		}
	}

	// Proxy routes for code generation service
	codeGen := router.Group("/api/v1/codegen")
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

	// Stats endpoint (separate from codegen)
	router.GET("/api/v1/stats", func(c *gin.Context) {
		proxyHandler.ProxyRequest(c)
	})

	// Proxy routes for model management
	models := router.Group("/api/v1/models")
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
