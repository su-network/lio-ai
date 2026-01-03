package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"lio-ai/internal/models"
)

// ChatRepository handles database operations for chats
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateChat creates a new chat
func (r *ChatRepository) CreateChat(chat *models.Chat) error {
	// Generate UUID for the chat
	chat.ChatUUID = uuid.New().String()
	
	query := `
		INSERT INTO chats (user_id, title, chat_uuid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	result, err := r.db.Exec(query, chat.UserID, chat.Title, chat.ChatUUID, now, now)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	chat.ID = id
	chat.CreatedAt = now
	chat.UpdatedAt = now
	return nil
}

// GetChatByID retrieves a chat by its ID
func (r *ChatRepository) GetChatByID(id int64) (*models.Chat, error) {
	query := `
		SELECT id, user_id, title, chat_uuid, created_at, updated_at
		FROM chats
		WHERE id = ?
	`

	chat := &models.Chat{}
	err := r.db.QueryRow(query, id).Scan(
		&chat.ID,
		&chat.UserID,
		&chat.Title,
		&chat.ChatUUID,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return chat, nil
}

// GetChatByUUID retrieves a chat by its UUID
func (r *ChatRepository) GetChatByUUID(chatUUID string) (*models.Chat, error) {
	query := `
		SELECT id, user_id, title, chat_uuid, created_at, updated_at
		FROM chats
		WHERE chat_uuid = ?
	`

	chat := &models.Chat{}
	err := r.db.QueryRow(query, chatUUID).Scan(
		&chat.ID,
		&chat.UserID,
		&chat.Title,
		&chat.ChatUUID,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	return chat, nil
}

// GetChatsByUserID retrieves all chats for a user
func (r *ChatRepository) GetChatsByUserID(userID string, limit, offset int) ([]models.Chat, error) {
	query := `
		SELECT id, user_id, title, chat_uuid, created_at, updated_at
		FROM chats
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	log.Printf("✓ GetChatsByUserID: Executing query with userID=%s, limit=%d, offset=%d", userID, limit, offset)
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		log.Printf("❌ GetChatsByUserID: Query failed: %v", err)
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}
	defer rows.Close()

	// Initialize with empty slice to return [] instead of null in JSON
	chats := make([]models.Chat, 0)
	for rows.Next() {
		var chat models.Chat
		err := rows.Scan(
			&chat.ID,
			&chat.UserID,
			&chat.Title,
			&chat.ChatUUID,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, chat)
	}

	log.Printf("✓ GetChatsByUserID: Found %d chats for userID=%s", len(chats), userID)
	return chats, nil
}

// UpdateChat updates a chat
func (r *ChatRepository) UpdateChat(chat *models.Chat) error {
	query := `
		UPDATE chats
		SET title = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, chat.Title, now, chat.ID)
	if err != nil {
		return fmt.Errorf("failed to update chat: %w", err)
	}

	chat.UpdatedAt = now
	return nil
}

// DeleteChat deletes a chat and its messages
func (r *ChatRepository) DeleteChat(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete messages first
	_, err = tx.Exec("DELETE FROM messages WHERE chat_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	// Delete chat
	_, err = tx.Exec("DELETE FROM chats WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return tx.Commit()
}

// CreateMessage creates a new message in a chat
func (r *ChatRepository) CreateMessage(message *models.Message) error {
	query := `
		INSERT INTO messages (chat_id, role, content, model, tokens, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, message.ChatID, message.Role, message.Content, message.Model, message.Tokens, now)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	message.ID = id
	message.CreatedAt = now

	// Update chat's updated_at
	_, err = r.db.Exec("UPDATE chats SET updated_at = ? WHERE id = ?", now, message.ChatID)
	if err != nil {
		return fmt.Errorf("failed to update chat timestamp: %w", err)
	}

	return nil
}

// GetMessagesByChatID retrieves all messages for a chat
func (r *ChatRepository) GetMessagesByChatID(chatID int64) ([]models.Message, error) {
	query := `
		SELECT id, chat_id, role, content, model, tokens, created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.Role,
			&message.Content,
			&message.Model,
			&message.Tokens,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// CountChatsByUserID counts the total number of chats for a user
func (r *ChatRepository) CountChatsByUserID(userID string) (int, error) {
	query := `SELECT COUNT(*) FROM chats WHERE user_id = ?`
	
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count chats: %w", err)
	}

	return count, nil
}
