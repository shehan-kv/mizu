package service

import (
	"mizu/internal/db/store"
	"mizu/internal/event"
	"mizu/internal/file"
	"mizu/internal/logger"
)

// FileService handles file related business logic.
// It relies on the provided FileStore for database operations
// and uses the Logger for audit and debugging.
type FileService struct {
	lg       logger.Logger
	evtSndr  *event.EventSender
	fileSt   store.FileStore
	fileStrg file.FileStorage
}

// NewFileService constructs a FileService that
// handles file related business logic. It needs a non-nil logger
// for audit and debugging, and a FileStore for persistence.
func NewFileService(
	lg logger.Logger,
	evtSndr *event.EventSender,
	fileSt store.FileStore,
	fileStrg file.FileStorage) *FileService {

	return &FileService{
		lg:       lg,
		evtSndr:  evtSndr,
		fileSt:   fileSt,
		fileStrg: fileStrg,
	}
}
