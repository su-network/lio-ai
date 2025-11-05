package models

import "time"

// Document represents a document in the system
// @Description Document model with timestamps
type Document struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateDocumentRequest represents the request payload for creating a document
type CreateDocumentRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required,min=1"`
}

// UpdateDocumentRequest represents the request payload for updating a document
type UpdateDocumentRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1,max=255"`
	Content *string `json:"content" binding:"omitempty,min=1"`
}

// DocumentResponse represents the response payload for a document
type DocumentResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Document model to DocumentResponse
func (d *Document) ToResponse() *DocumentResponse {
	return &DocumentResponse{
		ID:        d.ID,
		Title:     d.Title,
		Content:   d.Content,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
