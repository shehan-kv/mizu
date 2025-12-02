package service

import (
	"context"
	"mime/multipart"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/file"
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

func (s *FileService) StoreFile(
	ctx context.Context,
	channelId int64,
	mFile multipart.File,
	header *multipart.FileHeader) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
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

	fileUrl, err := s.fileStrg.Store(file.TypeFile, mFile, fileName)
	if err != nil {
		return ErrInternalError
	}

	// Ignored variable is the newly created file record's ID
	// Use it to broadcast messages to all connected SSE clients
	_, err = s.fileSt.CreateOne(ctx, params.FileCreate{
		ChannelId:    channelId,
		UserId:       actor.Id,
		OriginalName: header.Filename,
		SavedName:    fileName,
		Url:          fileUrl,
		Size:         header.Size,
	})

	// TODO: log this error and the resulting error from the file removal
	if err != nil {
		_ = s.fileStrg.Remove(fileUrl)
	}

	return nil
}

func (s *FileService) GetByChannelId(
	ctx context.Context,
	channelId int64,
	query *dto.FileSearch) (*common.Page[[]dto.FileResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"channel_id", channelId,
			"err", err)
		return nil, ErrInternalError
	}

	result, err := s.fileSt.GetByChannelId(ctx, channelId, &params.FileSearch{
		Keyword: query.Keyword,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

	if err != nil {
		s.lg.Error("could not retrieve files for channel",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := common.Page[[]dto.FileResponse]{
		Count: result.Total,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  make([]dto.FileResponse, len(result.Items)),
	}

	for i, file := range result.Items {
		resp.Data[i] = dto.FileResponse{
			Id:        file.Id,
			ChannelId: file.ChannelId,
			User: dto.FileUserResponse{
				Id:        file.UserId,
				FirstName: file.UserFirstName,
				LastName:  file.UserLastName,
			},
			OriginalName: file.OriginalName,
			SavedName:    file.SavedName,
			UploadedAt:   file.UploadedAt,
			Url:          file.Url,
			Size:         file.Size,
		}
	}

	return &resp, nil
}

func (s *FileService) GetByProjectId(
	ctx context.Context,
	projectId int64,
	query *dto.FileSearch) (*common.Page[[]dto.FileResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"project_id", projectId,
			"err", err)
		return nil, ErrInternalError
	}

	search := params.FileSearch{
		Keyword: query.Keyword,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	}

	files, err := s.fileSt.GetByProjectId(ctx, projectId, &search)

	if err != nil {
		s.lg.Error("could not retrieve files for project",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"project_id", projectId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := s.fileSt.CountByProjectId(ctx, projectId, &search)
	if err != nil {
		s.lg.Error("could not retrieve file count for project",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "file_service",
			"project_id", projectId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := common.Page[[]dto.FileResponse]{
		Count: count,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  make([]dto.FileResponse, len(files)),
	}

	for i, file := range files {
		resp.Data[i] = dto.FileResponse{
			Id:        file.Id,
			ChannelId: file.ChannelId,
			User: dto.FileUserResponse{
				Id:        file.UserId,
				FirstName: file.UserFirstName,
				LastName:  file.UserLastName,
			},
			OriginalName: file.OriginalName,
			SavedName:    file.SavedName,
			UploadedAt:   file.UploadedAt,
			Url:          file.Url,
			Size:         file.Size,
		}
	}

	return &resp, nil
}
