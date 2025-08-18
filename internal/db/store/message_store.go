package store

import (
	"context"
	"mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
)

// Defines the behavior required for managing messages
type MessageStore interface {

	// Creates a message
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to MessageCreate params
	//
	// Returns:
	//   - *aggregates.MessageWithUser
	//   - store.ErrInsertFailed: if create fails
	CreateOne(ctx context.Context, arg *params.MessageCreate) (*aggregates.MessageWithUser, error)
}
