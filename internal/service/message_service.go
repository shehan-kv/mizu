package service

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
)

// Handles message related operations.
type MessageService struct {
	lg    logger.Logger
	msgSt store.MessageStore
}

// Creates a new instance of MessageService.
// It takes a logger, a MessageStore for message-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - msgSt: project store that implements the MessageStore interface
//
// Returns:
//   - a pointer to a new MessageService
func NewMessageService(lg logger.Logger, msgSt store.MessageStore) *MessageService {
	return &MessageService{
		lg:    lg,
		msgSt: msgSt,
	}
}
