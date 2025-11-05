package handlers

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/utils"
)

// SearchHandler handles search operations
type SearchHandler struct {
	db *sql.DB
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// SearchAll performs a global search across documents, chats, and messages
func (h *SearchHandler) SearchAll(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestError(c, "Search query 'q' is required")
		return
	}

	userID := c.Query("user_id")
	searchTerm := "%" + strings.ToLower(query) + "%"

	results := gin.H{}

	// Search documents
	docQuery := `
		SELECT id, user_id, title, created_at
		FROM documents
		WHERE (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)
	`
	args := []interface{}{searchTerm, searchTerm}
	
	if userID != "" {
		docQuery += " AND user_id = ?"
		args = append(args, userID)
	}
	docQuery += " ORDER BY created_at DESC LIMIT 10"

	docRows, err := h.db.Query(docQuery, args...)
	if err == nil {
		defer docRows.Close()
		var documents []gin.H
		for docRows.Next() {
			var id int64
			var uid, title, createdAt string
			docRows.Scan(&id, &uid, &title, &createdAt)
			documents = append(documents, gin.H{
				"id":         id,
				"user_id":    uid,
				"title":      title,
				"created_at": createdAt,
			})
		}
		results["documents"] = documents
	}

	// Search chats
	chatQuery := `
		SELECT id, user_id, title, created_at
		FROM chats
		WHERE LOWER(title) LIKE ?
	`
	chatArgs := []interface{}{searchTerm}
	
	if userID != "" {
		chatQuery += " AND user_id = ?"
		chatArgs = append(chatArgs, userID)
	}
	chatQuery += " ORDER BY created_at DESC LIMIT 10"

	chatRows, err := h.db.Query(chatQuery, chatArgs...)
	if err == nil {
		defer chatRows.Close()
		var chats []gin.H
		for chatRows.Next() {
			var chat models.Chat
			chatRows.Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.CreatedAt)
			chats = append(chats, gin.H{
				"id":         chat.ID,
				"user_id":    chat.UserID,
				"title":      chat.Title,
				"created_at": chat.CreatedAt,
			})
		}
		results["chats"] = chats
	}

	// Search messages
	msgQuery := `
		SELECT m.id, m.chat_id, m.role, m.content, m.created_at, c.title as chat_title
		FROM messages m
		JOIN chats c ON m.chat_id = c.id
		WHERE LOWER(m.content) LIKE ?
	`
	msgArgs := []interface{}{searchTerm}
	
	if userID != "" {
		msgQuery += " AND c.user_id = ?"
		msgArgs = append(msgArgs, userID)
	}
	msgQuery += " ORDER BY m.created_at DESC LIMIT 10"

	msgRows, err := h.db.Query(msgQuery, msgArgs...)
	if err == nil {
		defer msgRows.Close()
		var messages []gin.H
		for msgRows.Next() {
			var msg models.Message
			var chatTitle string
			msgRows.Scan(&msg.ID, &msg.ChatID, &msg.Role, &msg.Content, &msg.CreatedAt, &chatTitle)
			
			// Truncate content for search results
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			
			messages = append(messages, gin.H{
				"id":         msg.ID,
				"chat_id":    msg.ChatID,
				"chat_title": chatTitle,
				"role":       msg.Role,
				"content":    content,
				"created_at": msg.CreatedAt,
			})
		}
		results["messages"] = messages
	}

	utils.SuccessResponse(c, gin.H{
		"query":   query,
		"results": results,
	})
}

// SearchDocuments performs advanced document search with filters
func (h *SearchHandler) SearchDocuments(c *gin.Context) {
	query := c.Query("q")
	userID := c.Query("user_id")

	var conditions []string
	var args []interface{}

	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		conditions = append(conditions, "(LOWER(title) LIKE ? OR LOWER(content) LIKE ?)")
		args = append(args, searchTerm, searchTerm)
	}

	if userID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userID)
	}

	sqlQuery := "SELECT id, user_id, title, created_at, updated_at FROM documents"
	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	sqlQuery += " ORDER BY updated_at DESC LIMIT 50"

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		utils.InternalError(c, "Failed to search documents")
		return
	}
	defer rows.Close()

	var documents []gin.H
	for rows.Next() {
		var id int64
		var uid, title, createdAt, updatedAt string
		err := rows.Scan(&id, &uid, &title, &createdAt, &updatedAt)
		if err != nil {
			continue
		}
		documents = append(documents, gin.H{
			"id":         id,
			"user_id":    uid,
			"title":      title,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	utils.SuccessResponse(c, gin.H{
		"count":     len(documents),
		"documents": documents,
	})
}

// SearchChats performs advanced chat search
func (h *SearchHandler) SearchChats(c *gin.Context) {
	query := c.Query("q")
	userID := c.Query("user_id")

	var conditions []string
	var args []interface{}

	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		conditions = append(conditions, "LOWER(title) LIKE ?")
		args = append(args, searchTerm)
	}

	if userID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userID)
	}

	sqlQuery := "SELECT id, user_id, title, created_at, updated_at FROM chats"
	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	sqlQuery += " ORDER BY updated_at DESC LIMIT 50"

	rows, err := h.db.Query(sqlQuery, args...)
	if err != nil {
		utils.InternalError(c, "Failed to search chats")
		return
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		err := rows.Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			continue
		}
		chats = append(chats, chat)
	}

	utils.SuccessResponse(c, gin.H{
		"count": len(chats),
		"chats": chats,
	})
}

// GetRecentActivity returns recent user activity
func (h *SearchHandler) GetRecentActivity(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		utils.BadRequestError(c, "user_id is required")
		return
	}

	limit := 20

	// Recent chats
	chatRows, _ := h.db.Query(`
		SELECT id, title, created_at, updated_at
		FROM chats
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`, userID, limit)

	var recentChats []gin.H
	if chatRows != nil {
		defer chatRows.Close()
		for chatRows.Next() {
			var id int64
			var title string
			var createdAt, updatedAt string
			chatRows.Scan(&id, &title, &createdAt, &updatedAt)
			recentChats = append(recentChats, gin.H{
				"id":         id,
				"title":      title,
				"type":       "chat",
				"created_at": createdAt,
				"updated_at": updatedAt,
			})
		}
	}

	// Recent documents
	docRows, _ := h.db.Query(`
		SELECT id, title, created_at, updated_at
		FROM documents
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ?
	`, userID, limit)

	var recentDocs []gin.H
	if docRows != nil {
		defer docRows.Close()
		for docRows.Next() {
			var id int64
			var title string
			var createdAt, updatedAt string
			docRows.Scan(&id, &title, &createdAt, &updatedAt)
			recentDocs = append(recentDocs, gin.H{
				"id":         id,
				"title":      title,
				"type":       "document",
				"created_at": createdAt,
				"updated_at": updatedAt,
			})
		}
	}

	utils.SuccessResponse(c, gin.H{
		"recent_chats":     recentChats,
		"recent_documents": recentDocs,
	})
}
