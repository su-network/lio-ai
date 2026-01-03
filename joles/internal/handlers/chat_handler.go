package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// ChatHandler handles HTTP requests for chats
type ChatHandler struct {
	service *services.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(service *services.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

// CreateChat handles POST /api/v1/chats
func (h *ChatHandler) CreateChat(c *gin.Context) {
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Use authenticated user's ID, NOT client-provided one
	chat, err := h.service.CreateChat(userID.(string), req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create chat",
			"code":  "CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

// GetChat handles GET /api/v1/chats/:id
func (h *ChatHandler) GetChat(c *gin.Context) {
	// Get authenticated user
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid chat id",
			"code":  "INVALID_ID",
		})
		return
	}

	chat, err := h.service.GetChat(id, userID.(string))
	if err != nil {
		if err == services.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied",
				"code":  "FORBIDDEN",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error": "chat not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// GetChatByUUID handles GET /api/v1/chats/uuid/:uuid
func (h *ChatHandler) GetChatByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat uuid"})
		return
	}

	chat, err := h.service.GetChatByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// GetUserChats handles GET /api/v1/chats
func (h *ChatHandler) GetUserChats(c *gin.Context) {
	// Get authenticated user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("❌ GetUserChats: user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Log the userID for debugging
	log.Printf("✓ GetUserChats: userID from context: %v (type: %T)", userID, userID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate limit
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	// Convert userID to string safely
	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("❌ GetUserChats: userID type assertion failed, got type: %T", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user context",
			"code":  "INVALID_USER_CONTEXT",
		})
		return
	}

	log.Printf("✓ GetUserChats: calling service with userID=%s, limit=%d, offset=%d", userIDStr, limit, offset)

	// Use authenticated user's ID, NOT query parameter
	chats, total, err := h.service.GetUserChats(userIDStr, limit, offset)
	if err != nil {
		// Log detailed error
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch chats",
			"code":  "FETCH_FAILED",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   chats,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateChat handles PUT /api/v1/chats/:id
func (h *ChatHandler) UpdateChat(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.UpdateChat(id, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// DeleteChat handles DELETE /api/v1/chats/:id
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	if err := h.service.DeleteChat(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "chat deleted successfully"})
}

// SendMessage handles POST /api/v1/chats/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.service.SendMessage(id, req.Role, req.Content, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// SendMessageByUUID handles POST /api/v1/chats/uuid/:uuid/messages
func (h *ChatHandler) SendMessageByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat uuid"})
		return
	}

	var req models.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.service.SendMessageByUUID(uuid, req.Role, req.Content, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages handles GET /api/v1/chats/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	messages, err := h.service.GetChatMessages(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"total": len(messages),
	})
}

// GetMessagesByUUID handles GET /api/v1/chats/uuid/:uuid/messages
func (h *ChatHandler) GetMessagesByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat uuid"})
		return
	}

	messages, err := h.service.GetChatMessagesByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"total": len(messages),
	})
}

// ChatCompletion handles POST /api/v1/chat/completions
func (h *ChatHandler) ChatCompletion(c *gin.Context) {
	var req models.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CreateChatCompletion(&req)
	if err != nil {
		// Preserve upstream AI service status codes (e.g., 429 rate limit)
		var aiErr *services.AIServiceError
		if errors.As(err, &aiErr) && aiErr != nil {
			status := aiErr.StatusCode
			if status == 0 {
				status = http.StatusBadGateway
			}

			detail := aiErr.Error()
			// Try extracting a clean "detail" from the upstream JSON
			var upstream struct {
				Detail  string `json:"detail"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if json.Unmarshal([]byte(aiErr.Body), &upstream) == nil {
				if upstream.Detail != "" {
					detail = upstream.Detail
				} else if upstream.Message != "" {
					detail = upstream.Message
				} else if upstream.Error != "" {
					detail = upstream.Error
				}
			}

			c.JSON(status, gin.H{"detail": detail})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
