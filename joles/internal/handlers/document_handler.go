package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
	"lio-ai/internal/services"
)

// DocumentHandler handles document HTTP requests
type DocumentHandler struct {
	service *services.DocumentService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(service *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// CreateDocument handles POST /api/v1/documents
// @Summary Create a new document
// @Description Create a new document with title and content
// @Accept json
// @Produce json
// @Param document body models.CreateDocumentRequest true "Document data"
// @Success 201 {object} models.DocumentResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/documents [post]
func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req models.CreateDocumentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.service.CreateDocument(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetDocuments handles GET /api/v1/documents
// @Summary Get all documents
// @Description Retrieve all documents with pagination
// @Produce json
// @Param skip query int false "Number of documents to skip" default(0)
// @Param limit query int false "Maximum documents to return" default(100)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/documents [get]
func (h *DocumentHandler) GetDocuments(c *gin.Context) {
	skip := 0
	limit := 100

	if s := c.Query("skip"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val >= 0 {
			skip = val
		}
	}

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	docs, total, err := h.service.GetDocuments(skip, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"total": total,
		"skip":  skip,
		"limit": limit,
	})
}

// GetDocument handles GET /api/v1/documents/:id
// @Summary Get a specific document
// @Description Retrieve a document by ID
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} models.DocumentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/documents/{id} [get]
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, err := h.service.GetDocument(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// UpdateDocument handles PUT /api/v1/documents/:id
// @Summary Update a document
// @Description Update an existing document
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param document body models.UpdateDocumentRequest true "Document updates"
// @Success 200 {object} models.DocumentResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/documents/{id} [put]
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var req models.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.service.UpdateDocument(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// DeleteDocument handles DELETE /api/v1/documents/:id
// @Summary Delete a document
// @Description Delete a document by ID
// @Param id path int true "Document ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	if err := h.service.DeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
