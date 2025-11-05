package main

import (
	"fmt"
	"log"
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

	// Auth middleware (permissive - sets context but doesn't block)
	router.Use(middleware.AuthMiddleware())

	// Initialize repositories, services, and handlers
	docRepo := repositories.NewDocumentRepository(database.GetConnection())
	docService := services.NewDocumentService(docRepo)
	docHandler := handlers.NewDocumentHandler(docService)

	chatRepo := repositories.NewChatRepository(database.GetConnection())
	chatService := services.NewChatService(chatRepo)
	chatHandler := handlers.NewChatHandler(chatService)

	usageRepo := repositories.NewUsageRepository(database.GetConnection())
	usageService := services.NewUsageService(usageRepo)
	usageHandler := handlers.NewUsageHandler(usageService)

	// System handlers
	systemHandler := handlers.NewSystemHandler(database.GetConnection())
	searchHandler := handlers.NewSearchHandler(database.GetConnection())
	batchHandler := handlers.NewBatchHandler(docService, chatService, database.GetConnection())

	// Apply usage tracking middleware (after initialization)
	router.Use(middleware.UsageTracking(usageService))

	// Initialize proxy handler for FastAPI backend
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}
	proxyHandler := handlers.NewProxyHandler(backendURL)

	// Root endpoint
	router.GET("/", systemHandler.GetInfo)

	// Health check with backend verification
	router.GET("/health", systemHandler.HealthCheck)

	// System endpoints
	router.GET("/stats", systemHandler.GetStats)

	// Document API routes (direct)
	api := router.Group("/api/v1")
	{
		// Metrics endpoint
		api.GET("/metrics", systemHandler.GetMetrics)

		// Search endpoints
		search := api.Group("/search")
		{
			search.GET("", searchHandler.SearchAll)
			search.GET("/documents", searchHandler.SearchDocuments)
			search.GET("/chats", searchHandler.SearchChats)
			search.GET("/activity", searchHandler.GetRecentActivity)
		}

		documents := api.Group("/documents")
		{
			documents.POST("", docHandler.CreateDocument)
			documents.GET("", docHandler.GetDocuments)
			documents.GET("/:id", docHandler.GetDocument)
			documents.PUT("/:id", docHandler.UpdateDocument)
			documents.DELETE("/:id", docHandler.DeleteDocument)
		}

		// Chat API routes
		chats := api.Group("/chats")
		{
			chats.POST("", chatHandler.CreateChat)
			chats.GET("", chatHandler.GetUserChats)
			chats.GET("/:id", chatHandler.GetChat)
			chats.PUT("/:id", chatHandler.UpdateChat)
			chats.DELETE("/:id", chatHandler.DeleteChat)
			chats.POST("/:id/messages", chatHandler.SendMessage)
			chats.GET("/:id/messages", chatHandler.GetMessages)
		}

		// Chat completion endpoint
		api.POST("/chat/completions", chatHandler.ChatCompletion)

		// Usage tracking API routes
		usage := api.Group("/usage")
		{
			usage.GET("/quota", usageHandler.GetQuotaStatus)
			usage.GET("/summary", usageHandler.GetUsageSummary)
			usage.GET("/dashboard", usageHandler.GetDashboard)
			usage.POST("/track", usageHandler.TrackUsage)
			usage.POST("/check-quota", usageHandler.CheckQuota)
			usage.PUT("/quota/:user_id", usageHandler.UpdateQuota)
		}

		// Batch operations
		batch := api.Group("/batch")
		{
			batch.POST("/documents", batchHandler.BatchCreateDocuments)
			batch.DELETE("/documents", batchHandler.BatchDeleteDocuments)
			batch.DELETE("/chats", batchHandler.BatchDeleteChats)
			batch.PUT("/documents/tags", batchHandler.BulkUpdateTags)
			batch.GET("/export", batchHandler.ExportData)
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
		codeGen.GET("/stats", func(c *gin.Context) {
			proxyHandler.ProxyRequest(c)
		})
	}

	// Proxy all unmatched routes to backend
	router.NoRoute(func(c *gin.Context) {
		proxyHandler.ProxyRequest(c)
	})

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("✓ Starting server at http://%s", addr)
	log.Printf("✓ API Documentation available at http://%s/docs", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
