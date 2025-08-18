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

	// Gets IDs of all channels by user ID
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - userId: id of the user
	//
	// Returns:
	//   - []int64: array of channel ids
	//   - store.ErrQueryFailed: if query fails
	GetChannelsByUserId(ctx context.Context, userId int64) ([]int64, error)
}
