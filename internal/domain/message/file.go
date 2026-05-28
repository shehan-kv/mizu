package message

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

type File struct {
	id           FileID
	channelID    ChannelID
	userID       iam.UserID
	originalName FileName
	savedName    FileName
	storageKey   string
	mimeType     string
	size         int64

	events []common.Event

	uploadedAt time.Time
}

func NewFile(
	id FileID,
	channelID ChannelID,
	userID iam.UserID,
	originalName FileName,
	savedName FileName,
	storageKey string,
	mimeType string,
	size int64,
	now time.Time,
) *File {

	f := File{
		id:           id,
		channelID:    channelID,
		userID:       userID,
		originalName: originalName,
		savedName:    savedName,
		storageKey:   storageKey,
		mimeType:     mimeType,
		size:         size,
		uploadedAt:   now,
	}

	f.events = append(f.events, FileCreatedEvent{
		FileID:       id,
		UserID:       userID,
		ChannelID:    channelID,
		OriginalName: originalName,
		StorageKey:   storageKey,
		MimeType:     mimeType,
		Size:         size,
		OccurredAt:   now,
	})

	return &f
}

func RestoreFile(
	id FileID,
	channelID ChannelID,
	userID iam.UserID,
	originalName FileName,
	savedName FileName,
	storageKey string,
	mimeType string,
	size int64,
	uploadedAt time.Time,
) *File {
	return &File{
		id:           id,
		channelID:    channelID,
		userID:       userID,
		originalName: originalName,
		savedName:    savedName,
		storageKey:   storageKey,
		mimeType:     mimeType,
		size:         size,
		uploadedAt:   uploadedAt,
	}
}

func (f *File) ID() FileID {
	return f.id
}

func (f *File) ChannelID() ChannelID {
	return f.channelID
}

func (f *File) UserID() iam.UserID {
	return f.userID
}

func (f *File) OriginalName() FileName {
	return f.originalName
}

func (f *File) SavedName() FileName {
	return f.savedName
}

func (f *File) StorageKey() string {
	return f.storageKey
}

func (f *File) MimeType() string {
	return f.mimeType
}

func (f *File) Size() int64 {
	return f.size
}

func (f *File) UploadedAt() time.Time {
	return f.uploadedAt
}

func (f *File) PullEvents() []common.Event {
	events := f.events
	f.events = nil

	return events
}
