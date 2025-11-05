package models

import "time"

// Chat represents a chat conversation
type Chat struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a chat
type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Model     string    `json:"model,omitempty"`
	Tokens    int       `json:"tokens,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatWithMessages represents a chat with its messages
type ChatWithMessages struct {
	Chat
	Messages []Message `json:"messages"`
}

// ChatRequest represents the request to create a new chat
type ChatRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

// MessageRequest represents the request to send a message
type MessageRequest struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
	Model   string `json:"model,omitempty"`
}

// ChatCompletionRequest represents a request for chat completion
type ChatCompletionRequest struct {
	ChatID   int64  `json:"chat_id,omitempty"`
	Message  string `json:"message" binding:"required"`
	Model    string `json:"model,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Title    string `json:"title,omitempty"`
}

// ChatCompletionResponse represents the response from chat completion
type ChatCompletionResponse struct {
	ChatID    int64   `json:"chat_id"`
	MessageID int64   `json:"message_id"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Model     string  `json:"model"`
	Tokens    int     `json:"tokens"`
	CreatedAt time.Time `json:"created_at"`
}
