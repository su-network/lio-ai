package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// MockDocumentService is a mock implementation of DocumentService
type MockDocumentService struct {
	createFunc    func(*models.CreateDocumentRequest) (*models.DocumentResponse, error)
	getFunc       func(uint) (*models.DocumentResponse, error)
	getAllFunc    func(int, int) ([]*models.DocumentResponse, int64, error)
	updateFunc    func(uint, *models.UpdateDocumentRequest) (*models.DocumentResponse, error)
	deleteFunc    func(uint) error
}

func (m *MockDocumentService) CreateDocument(req *models.CreateDocumentRequest) (*models.DocumentResponse, error) {
	return m.createFunc(req)
}

func (m *MockDocumentService) GetDocument(id uint) (*models.DocumentResponse, error) {
	return m.getFunc(id)
}

func (m *MockDocumentService) GetDocuments(skip, limit int) ([]*models.DocumentResponse, int64, error) {
	return m.getAllFunc(skip, limit)
}

func (m *MockDocumentService) UpdateDocument(id uint, req *models.UpdateDocumentRequest) (*models.DocumentResponse, error) {
	return m.updateFunc(id, req)
}

func (m *MockDocumentService) DeleteDocument(id uint) error {
	return m.deleteFunc(id)
}

func TestCreateDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockDocumentService{
		createFunc: func(req *models.CreateDocumentRequest) (*models.DocumentResponse, error) {
			return &models.DocumentResponse{
				ID:      1,
				Title:   req.Title,
				Content: req.Content,
			}, nil
		},
	}

	handler := NewDocumentHandler((*services.DocumentService)(nil))
	handler.service = mockService

	router := gin.New()
	router.POST("/documents", handler.CreateDocument)

	body := models.CreateDocumentRequest{
		Title:   "Test Document",
		Content: "Test Content",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/documents", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetDocuments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockDocumentService{
		getAllFunc: func(skip, limit int) ([]*models.DocumentResponse, int64, error) {
			return []*models.DocumentResponse{
				{ID: 1, Title: "Doc 1", Content: "Content 1"},
				{ID: 2, Title: "Doc 2", Content: "Content 2"},
			}, 2, nil
		},
	}

	handler := NewDocumentHandler((*services.DocumentService)(nil))
	handler.service = mockService

	router := gin.New()
	router.GET("/documents", handler.GetDocuments)

	req, _ := http.NewRequest("GET", "/documents?skip=0&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
