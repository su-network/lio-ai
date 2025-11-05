package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/services"
	"lio-ai/internal/utils"
)

// BatchHandler handles batch operations
type BatchHandler struct {
	docService  *services.DocumentService
	chatService *services.ChatService
	db          *sql.DB
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(docService *services.DocumentService, chatService *services.ChatService, db *sql.DB) *BatchHandler {
	return &BatchHandler{
		docService:  docService,
		chatService: chatService,
		db:          db,
	}
}

// BatchCreateDocuments creates multiple documents
func (h *BatchHandler) BatchCreateDocuments(c *gin.Context) {
	var req struct {
		Documents []models.CreateDocumentRequest `json:"documents" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if len(req.Documents) == 0 {
		utils.BadRequestError(c, "No documents provided")
		return
	}

	if len(req.Documents) > 100 {
		utils.BadRequestError(c, "Maximum 100 documents per batch")
		return
	}

	var created []models.DocumentResponse
	var failed []gin.H

	for i, docReq := range req.Documents {
		doc, err := h.docService.CreateDocument(&docReq)
		if err != nil {
			failed = append(failed, gin.H{
				"index": i,
				"error": err.Error(),
			})
			continue
		}
		created = append(created, *doc)
	}

	utils.SuccessResponse(c, gin.H{
		"created": created,
		"failed":  failed,
		"summary": gin.H{
			"total":     len(req.Documents),
			"succeeded": len(created),
			"failed":    len(failed),
		},
	})
}

// BatchDeleteDocuments deletes multiple documents
func (h *BatchHandler) BatchDeleteDocuments(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequestError(c, "No document IDs provided")
		return
	}

	if len(req.IDs) > 100 {
		utils.BadRequestError(c, "Maximum 100 documents per batch")
		return
	}

	var deleted []int64
	var failed []gin.H

	for _, id := range req.IDs {
		err := h.docService.DeleteDocument(uint(id))
		if err != nil {
			failed = append(failed, gin.H{
				"id":    id,
				"error": err.Error(),
			})
			continue
		}
		deleted = append(deleted, id)
	}

	utils.SuccessResponse(c, gin.H{
		"deleted": deleted,
		"failed":  failed,
		"summary": gin.H{
			"total":     len(req.IDs),
			"succeeded": len(deleted),
			"failed":    len(failed),
		},
	})
}

// BatchDeleteChats deletes multiple chats
func (h *BatchHandler) BatchDeleteChats(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequestError(c, "No chat IDs provided")
		return
	}

	if len(req.IDs) > 100 {
		utils.BadRequestError(c, "Maximum 100 chats per batch")
		return
	}

	var deleted []int64
	var failed []gin.H

	for _, id := range req.IDs {
		err := h.chatService.DeleteChat(id)
		if err != nil {
			failed = append(failed, gin.H{
				"id":    id,
				"error": err.Error(),
			})
			continue
		}
		deleted = append(deleted, id)
	}

	utils.SuccessResponse(c, gin.H{
		"deleted": deleted,
		"failed":  failed,
		"summary": gin.H{
			"total":     len(req.IDs),
			"succeeded": len(deleted),
			"failed":    len(failed),
		},
	})
}

// ExportData exports user data
func (h *BatchHandler) ExportData(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		utils.BadRequestError(c, "user_id is required")
		return
	}

	// Get all user chats
	chats, _, _ := h.chatService.GetUserChats(userID, 1, 1000)

	// Get all user documents
	docRows, _ := h.db.Query(`
		SELECT id, title, content, created_at, updated_at
		FROM documents
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	var documents []gin.H
	if docRows != nil {
		defer docRows.Close()
		for docRows.Next() {
			var id int64
			var title, content, createdAt, updatedAt string
			docRows.Scan(&id, &title, &content, &createdAt, &updatedAt)
			documents = append(documents, gin.H{
				"id":         id,
				"title":      title,
				"content":    content,
				"created_at": createdAt,
				"updated_at": updatedAt,
			})
		}
	}

	// Get usage summary
	var totalRequests int
	var totalTokens int
	var totalCost float64
	h.db.QueryRow(`
		SELECT 
			COUNT(*) as total_requests,
			COALESCE(SUM(tokens_total), 0) as total_tokens,
			COALESCE(SUM(cost_usd), 0.0) as total_cost
		FROM usage_metrics
		WHERE user_id = ?
	`, userID).Scan(&totalRequests, &totalTokens, &totalCost)

	usageSummary := gin.H{
		"total_requests": totalRequests,
		"total_tokens":   totalTokens,
		"total_cost":     totalCost,
	}

	export := gin.H{
		"user_id":   userID,
		"chats":     chats,
		"documents": documents,
		"usage":     usageSummary,
		"exported_at": gin.H{
			"timestamp": gin.H{},
		},
	}

	utils.SuccessResponse(c, export)
}

// BulkUpdateTags updates tags for multiple documents
func (h *BatchHandler) BulkUpdateTags(c *gin.Context) {
	var req struct {
		IDs  []int64 `json:"ids" binding:"required"`
		Tags string  `json:"tags" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.BadRequestError(c, "No document IDs provided")
		return
	}

	var updated []int64
	var failed []gin.H

	for _, id := range req.IDs {
		_, err := h.db.Exec(`
			UPDATE documents
			SET tags = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Tags, id)

		if err != nil {
			failed = append(failed, gin.H{
				"id":    id,
				"error": err.Error(),
			})
			continue
		}
		updated = append(updated, id)
	}

	utils.SuccessResponse(c, gin.H{
		"updated": updated,
		"failed":  failed,
		"summary": gin.H{
			"total":     len(req.IDs),
			"succeeded": len(updated),
			"failed":    len(failed),
		},
	})
}
