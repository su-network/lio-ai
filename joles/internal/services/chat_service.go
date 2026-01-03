package services

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"lio-ai/internal/models"
	"lio-ai/internal/repositories"
)

// ChatService handles business logic for chats
type ChatService struct {
	repo *repositories.ChatRepository
}

// NewChatService creates a new chat service
func NewChatService(repo *repositories.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

// CreateChat creates a new chat
func (s *ChatService) CreateChat(userID, title string) (*models.Chat, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if title == "" {
		title = "New Chat"
	}

	chat := &models.Chat{
		UserID: userID,
		Title:  title,
	}

	if err := s.repo.CreateChat(chat); err != nil {
		return nil, err
	}

	return chat, nil
}

// GetChat retrieves a chat by ID with its messages (with ownership check)
func (s *ChatService) GetChat(id int64, userID string) (*models.ChatWithMessages, error) {
	chat, err := s.repo.GetChatByID(id)
	if err != nil {
		return nil, err
	}

	// CRITICAL: Verify ownership
	if chat.UserID != userID {
		return nil, ErrUnauthorized
	}

	messages, err := s.repo.GetMessagesByChatID(id)
	if err != nil {
		return nil, err
	}

	return &models.ChatWithMessages{
		Chat:     *chat,
		Messages: messages,
	}, nil
}

// GetChatByUUID retrieves a chat by UUID with its messages
func (s *ChatService) GetChatByUUID(uuid string) (*models.ChatWithMessages, error) {
	chat, err := s.repo.GetChatByUUID(uuid)
	if err != nil {
		return nil, err
	}

	messages, err := s.repo.GetMessagesByChatID(chat.ID)
	if err != nil {
		return nil, err
	}

	return &models.ChatWithMessages{
		Chat:     *chat,
		Messages: messages,
	}, nil
}

// GetUserChats retrieves all chats for a user
func (s *ChatService) GetUserChats(userID string, limit, offset int) ([]models.Chat, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	chats, err := s.repo.GetChatsByUserID(userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountChatsByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	return chats, total, nil
}

// UpdateChat updates a chat's title
func (s *ChatService) UpdateChat(id int64, title string) (*models.Chat, error) {
	chat, err := s.repo.GetChatByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		chat.Title = title
	}

	if err := s.repo.UpdateChat(chat); err != nil {
		return nil, err
	}

	return chat, nil
}

// DeleteChat deletes a chat
func (s *ChatService) DeleteChat(id int64) error {
	return s.repo.DeleteChat(id)
}

// SendMessage sends a message in a chat
func (s *ChatService) SendMessage(chatID int64, role, content, model string) (*models.Message, error) {
	// Validate chat exists
	_, err := s.repo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if role == "" {
		role = "user"
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var modelPtr *string
	if model != "" {
		modelPtr = &model
	}

	message := &models.Message{
		ChatID:  chatID,
		Role:    role,
		Content: content,
		Model:   modelPtr,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, err
	}

	return message, nil
}

// SendMessageByUUID sends a message in a chat identified by UUID
func (s *ChatService) SendMessageByUUID(uuid, role, content, model string) (*models.Message, error) {
	// Validate chat exists and get ID
	chat, err := s.repo.GetChatByUUID(uuid)
	if err != nil {
		return nil, err
	}

	if role == "" {
		role = "user"
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var modelPtr *string
	if model != "" {
		modelPtr = &model
	}

	message := &models.Message{
		ChatID:  chat.ID,
		Role:    role,
		Content: content,
		Model:   modelPtr,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetChatMessages retrieves all messages for a chat
func (s *ChatService) GetChatMessages(chatID int64) ([]models.Message, error) {
	// Validate chat exists
	_, err := s.repo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetMessagesByChatID(chatID)
}

// GetChatMessagesByUUID retrieves all messages for a chat identified by UUID
func (s *ChatService) GetChatMessagesByUUID(uuid string) ([]models.Message, error) {
	// Validate chat exists and get ID
	chat, err := s.repo.GetChatByUUID(uuid)
	if err != nil {
		return nil, err
	}

	return s.repo.GetMessagesByChatID(chat.ID)
}

// CreateChatCompletion creates a new chat or adds to existing one and gets AI response
func (s *ChatService) CreateChatCompletion(req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	var chatID int64
	var err error

	// Create new chat if chatID not provided
	if req.ChatID == 0 {
		userID := req.UserID
		if userID == "" {
			userID = "anonymous"
		}
		title := req.Title
		if title == "" {
			title = truncateText(req.Message, 50)
		}

		chat, err := s.CreateChat(userID, title)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat: %w", err)
		}
		chatID = chat.ID
	} else {
		chatID = req.ChatID
	}

	// Save user message
	_, err = s.SendMessage(chatID, "user", req.Message, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get chat history for context
	messages, err := s.repo.GetMessagesByChatID(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history: %w", err)
	}

	// Build messages array for AI service
	aiMessages := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		aiMessages = append(aiMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Call Python AI service for completion
	aiResponse, err := s.callAIService(req.Model, aiMessages, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Save AI response
	aiMessage, err := s.SendMessage(chatID, "assistant", aiResponse.Content, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	// Update tokens if available
	if aiResponse.Tokens > 0 {
		aiMessage.Tokens = aiResponse.Tokens
	}

	return &models.ChatCompletionResponse{
		ChatID:    chatID,
		MessageID: aiMessage.ID,
		Role:      aiMessage.Role,
		Content:   aiMessage.Content,
		Model:     aiMessage.Model,
		Tokens:    aiMessage.Tokens,
		CreatedAt: aiMessage.CreatedAt,
	}, nil
}

// callAIService calls the Python AI service for chat completion
func (s *ChatService) callAIService(model string, messages []map[string]interface{}, userID string) (*AIServiceResponse, error) {
	// Get AI service URL from environment
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8000"
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"user_id":  userID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make HTTP request to AI service
	resp, err := http.Post(
		aiServiceURL+"/api/v1/chat/completions",
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &AIServiceError{StatusCode: resp.StatusCode, Body: string(bodyBytes)}
	}

	// Parse response
	var result struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	// Extract content
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI service")
	}

	return &AIServiceResponse{
		Content: result.Choices[0].Message.Content,
		Tokens:  result.Usage.TotalTokens,
	}, nil
}

// AIServiceResponse represents the response from AI service
type AIServiceResponse struct {
	Content string
	Tokens  int
}

// AIServiceError captures non-200 responses from the Python AI service.
// This lets handlers preserve status codes (e.g., 429 rate limit, 401 auth).
type AIServiceError struct {
	StatusCode int
	Body       string
}

func (e *AIServiceError) Error() string {
	if e == nil {
		return "AI service error"
	}
	if e.Body != "" {
		return e.Body
	}
	return fmt.Sprintf("AI service returned error (status=%d)", e.StatusCode)
}

func IsAIServiceError(err error) (*AIServiceError, bool) {
	var aiErr *AIServiceError
	if errors.As(err, &aiErr) {
		return aiErr, true
	}
	return nil, false
}

// truncateText truncates text to specified length
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
