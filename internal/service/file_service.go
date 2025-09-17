package service

import (
	"context"
	"mime/multipart"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/event"
	"mizu/internal/file"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"path/filepath"

	"github.com/google/uuid"
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

func (fileSrv *FileService) StoreFile(
	ctx context.Context,
	channelId int64,
	mFile multipart.File,
	header *multipart.FileHeader) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		fileSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"channel_id", channelId,
			"err", err)
		return ErrInternalError
	}

	// Random UUID as file name to avoid name collisions
	randomName, err := uuid.NewRandom()
	if err != nil {
		return ErrInternalError
	}

	fileName := randomName.String() + filepath.Ext(header.Filename)

	fileUrl, err := fileSrv.fileStrg.Store(file.TypeFile, mFile, fileName)
	if err != nil {
		return ErrInternalError
	}

	// Ignored variable is the newly created file record's ID
	// Use it to broadcast messages to all connected SSE clients
	_, err = fileSrv.fileSt.CreateOne(ctx, params.FileCreate{
		ChannelId:    channelId,
		UserId:       actor.Id,
		OriginalName: header.Filename,
		SavedName:    fileName,
		Url:          fileUrl,
		Size:         header.Size,
	})

	// TODO: log this error and the resulting error from the file removal
	if err != nil {
		_ = fileSrv.fileStrg.Remove(fileUrl)
	}

	return nil
}
