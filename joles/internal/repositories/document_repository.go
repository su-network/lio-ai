package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"lio-ai/internal/models"
)

// DocumentRepository handles document database operations
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document
func (r *DocumentRepository) Create(doc *models.Document) error {
	query := `INSERT INTO documents (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, doc.Title, doc.Content, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	doc.ID = uint(id)
	return nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(id uint) (*models.Document, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM documents WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var doc models.Document
	err := row.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return &doc, nil
}

// GetAll retrieves all documents with pagination
func (r *DocumentRepository) GetAll(skip, limit int) ([]*models.Document, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM documents`
	var total int64
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get paginated results
	query := `SELECT id, title, content, created_at, updated_at FROM documents LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, limit, skip)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return docs, total, nil
}

// Update updates an existing document
func (r *DocumentRepository) Update(id uint, updates *models.Document) (*models.Document, error) {
	doc, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	if updates.Title != "" {
		doc.Title = updates.Title
	}
	if updates.Content != "" {
		doc.Content = updates.Content
	}
	doc.UpdatedAt = time.Now()

	query := `UPDATE documents SET title = ?, content = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, doc.Title, doc.Content, doc.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return doc, nil
}

// Delete deletes a document
func (r *DocumentRepository) Delete(id uint) error {
	query := `DELETE FROM documents WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}
