package store

import (
	"context"
	agg "mizu/internal/db/models/aggregates"
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
	CreateOne(ctx context.Context, projectId int64, arg *params.ContractCreate) (*agg.ContractCreateResult, error)

	// SignVersion adds a signature for the given contract version and user.
	// It checks if the user is attempting to sign a
	// contract/version that's already been signed/rejected.
	// It returns a boolean, distinguishing between
	// a newly created signature and existing signatures.
	//
	// If the contract or version was already signed, the method returns true.
	//
	// If any error occurs, store.ErrInsertFailed is returned.
	SignVersion(ctx context.Context, versionId int64, userId int64) (bool, error)

	// GetUsersWithSignature retrieves the contract metadata and every project-user’s
	// signature for the given contract version. It returns a pointer to
	// an *aggregates.ContractUserSignatures struct with the metadata and signatures.
	//
	//   - If an error occurs, store.ErrQueryFailed is returned.
	GetUsersWithSignature(ctx context.Context, versionId int64) (*agg.ContractUserSignatures, error)

	// RejectVersion adds a rejected signature for the given contract version and user.
	// It checks if the user is attempting to reject a
	// contract/version that's already been signed/rejected.
	// It returns a boolean, distinguishing between
	// a newly rejected signature and existing signatures.
	//
	// If the contract or version was already rejected or signed, the method returns true.
	//
	// If any error occurs, store.ErrInsertFailed is returned.
	RejectVersion(ctx context.Context, versionId int64, userId int64) (bool, error)

	// CreateRevision creates a contract revision request for a specified contract.
	//
	// If any error occurs, store.ErrInsertFailed is returned.
	CreateRevision(ctx context.Context, arg *params.ContractRevisionCreate) error

	// AcceptRevision marks a contract revision as accepted.
	// If the contract revision is already marked as accepted,
	// it doesn't make any changes.
	//
	// If any error occurs, store.ErrInsertFailed is returned.
	AcceptRevision(ctx context.Context, revisionId int64, userId int64) error

	// RejectRevision marks a contract revision as rejected.
	// If the contract revision is already marked as rejected,
	// it doesn't make any changes.
	//
	// If any error occurs, store.ErrInsertFailed is returned.
	RejectRevision(ctx context.Context, revisionId int64, userId int64) error

	// GetRevisions returns the total number of revisions found and
	// a list of contract revision information with user data for
	// a specified contract using the contract ID.
	// The search criteria parameter can be used to filter results
	// by a keyword, limit and offset results.
	//
	// If any error occurs, store.ErrQueryFailed is returned.
	GetRevisions(
		ctx context.Context,
		contractId int64,
		arg *params.ContractRevisionSearch) (*agg.WithCount[agg.ContractRevisionWithUser], error)

	// GetContractStatsByProject returns the total number of contracts found and
	// a list of contract with metrics such as number of versions and revisions
	// for a specified project.
	// The search criteria parameter can be used to filter results
	// by a keyword, limit and offset results.
	//
	// If any error occurs, store.ErrQueryFailed is returned.
	GetContractStatsByProject(
		ctx context.Context,
		projectId int64,
		arg *params.ContractSearch) (*agg.WithCount[agg.ContractWithStats], error)

	GetVersionsByContractId(ctx context.Context, contractId int64) ([]agg.ContractVersion, error)

	CountByProjectId(ctx context.Context, projectId int64, arg *params.ContractSearch) (int64, error)

	GetByUserId(
		ctx context.Context,
		userId int64,
		arg *params.ContractSearch) ([]agg.ContractWithStats, error)

	CountByUserId(ctx context.Context, userId int64, arg *params.ContractSearch) (int64, error)
}
