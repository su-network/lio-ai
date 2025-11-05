package services

import (
	"fmt"

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

// GetChat retrieves a chat by ID with its messages
func (s *ChatService) GetChat(id int64) (*models.ChatWithMessages, error) {
	chat, err := s.repo.GetChatByID(id)
	if err != nil {
		return nil, err
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

	message := &models.Message{
		ChatID:  chatID,
		Role:    role,
		Content: content,
		Model:   model,
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

	// TODO: Call LLM service to get AI response
	// For now, return a placeholder response
	aiResponse := "This is a placeholder response. Integrate with LLM service for actual AI responses."
	
	// Save AI response
	aiMessage, err := s.SendMessage(chatID, "assistant", aiResponse, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
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

// truncateText truncates text to specified length
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
