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
	docService := services.NewDocumentService(docRepo)
	docHandler := handlers.NewDocumentHandler(docService)

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
	router.GET("/health", proxyHandler.HealthCheck)

	// Document API routes (direct)
	api := router.Group("/api/v1")
	{
		documents := api.Group("/documents")
		{
			documents.POST("", docHandler.CreateDocument)
			documents.GET("", docHandler.GetDocuments)
			documents.GET("/:id", docHandler.GetDocument)
			documents.PUT("/:id", docHandler.UpdateDocument)
			documents.DELETE("/:id", docHandler.DeleteDocument)
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
		codeGen.GET("/stats", func(c *gin.Context) {
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
	log.Printf("✓ Starting server at http://%s", addr)
	log.Printf("✓ API Documentation available at http://%s/docs", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
