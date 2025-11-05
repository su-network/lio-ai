package services

import (
	"fmt"

	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
)

// DocumentService handles document business logic
type DocumentService struct {
	repo *repositories.DocumentRepository
}

// NewDocumentService creates a new document service
func NewDocumentService(repo *repositories.DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(req *models.CreateDocumentRequest) (*models.DocumentResponse, error) {
	doc := &models.Document{
		Title:   req.Title,
		Content: req.Content,
	}

	if err := s.repo.Create(doc); err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	return doc.ToResponse(), nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(id uint) (*models.DocumentResponse, error) {
	doc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	return doc.ToResponse(), nil
}

// GetDocuments retrieves all documents with pagination
func (s *DocumentService) GetDocuments(skip, limit int) ([]*models.DocumentResponse, int64, error) {
	docs, total, err := s.repo.GetAll(skip, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("service error: %w", err)
	}

	responses := make([]*models.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

// UpdateDocument updates an existing document
func (s *DocumentService) UpdateDocument(id uint, req *models.UpdateDocumentRequest) (*models.DocumentResponse, error) {
	updates := &models.Document{}
	if req.Title != nil {
		updates.Title = *req.Title
	}
	if req.Content != nil {
		updates.Content = *req.Content
	}

	doc, err := s.repo.Update(id, updates)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("document not found")
	}

	return doc.ToResponse(), nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("service error: %w", err)
	}
	return nil
}
