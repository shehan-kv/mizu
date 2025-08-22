package store

import (
	"context"
	"mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
)

// ContractStore abstracts all persistence operations for contracts.
// Lower-level errors should be wrapped in domain-specific error types
// defined in the store package.
// Implementations must be safe for concurrent use.
type ContractStore interface {

	// CreateOne returns a pointer to a aggregates.ContractCreateResult after
	// creating a new contract. It also creates the system generated messages
	// in relevant channels. If any error occurs, nothing is saved in the database.
	//
	//   - If a foreign key constraint violation occurs, it returns store.ErrForeignKeyViolation.
	//   - If a unique constraint violation occurs, it returns store.ErrUniqueViolation.
	//   - If a not-null constraint violation occurs, it returns store.ErrNotNullViolation.
	//   - If any other errors occur, it returns store.ErrInsertFailed
	CreateOne(ctx context.Context, projectId int64, arg *params.ContractCreate) (*aggregates.ContractCreateResult, error)
}
