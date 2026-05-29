package session

import (
	"context"
	"mizu/internal/domain/iam"
)

type Store interface {
	Add(ctx context.Context, session *Session) error
	Get(ctx context.Context, id SessionID) (*Session, error)
	ListByUser(ctx context.Context, uID iam.UserID) ([]*Session, error)
	Delete(ctx context.Context, id SessionID) error
}
