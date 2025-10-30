package store

import (
	"context"
	"mizu/internal/db/models"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
)

// Defines the behavior required for managing users
type UserStore interface {

	// Creates one user
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to UserCreate params
	//
	// Returns:
	//   - int64: id of new user
	//   - store.ErrUniqueViolation: if email already exists
	//   - store.ErrInsertFailed: if create fails
	CreateOne(ctx context.Context, arg *params.UserCreate) (int64, error)

	// Gets a user by id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//
	// Returns:
	//   - *models.User: reference to a models.User instace
	//   - store.ErrRecordNotFound: if not found
	GetById(ctx context.Context, id int64) (*models.User, error)

	// Gets a user by email
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - email: email of the user
	//
	// Returns:
	//   - *models.User: reference to a models.User instace
	//   - store.ErrRecordNotFound: if not found
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Gets user password by id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//
	// Returns:
	//   - string: user's password
	//   - store.ErrRecordNotFound: if not found
	GetPasswordById(ctx context.Context, id int64) (string, error)

	// Counts all users
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//
	// Returns:
	//   - int64: number of users
	//   - store.ErrQueryFailed: if query fails
	CountAll(ctx context.Context) (int64, error)

	// Sets password by user id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//   - password: new password to set
	//
	// Returns:
	//   - store.ErrUpdateFailed: if update fails
	SetPasswordById(ctx context.Context, id int64, password string) error

	// Deletes user by id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//
	// Returns:
	//   - store.ErrDeleteFailed: if delete fails
	DeleteById(ctx context.Context, id int64) error

	// Updates last login by user id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the user
	//
	// Returns:
	//   - store.ErrUpdateFailed: if update fails
	UpdateLastLogin(ctx context.Context, id int64) error

	// Gets role by id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of the role
	//
	// Returns:
	//   - store.ErrRecordNotFound: if role not found
	GetRoleById(ctx context.Context, id int64) (*models.Role, error)

	// Creates an onboard request for a user
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to UserOnboardReqCreate params
	//
	// Returns:
	//   - store.ErrInsertFailed: if create fails
	//	 - store.ErrUniqueViolation: if token isn't unique
	// 	 - store.ErrNotNullViolation: if a required field is null
	CreateOnboardReq(ctx context.Context, arg *params.UserOnboardReqCreate) error

	// Onboards a user.
	// 	- creates a user without password
	// 	- creates an onboard request
	// 	- creates a new channel for messaging
	// 	- adds user to the channel
	// 	- adds the account that created the user to the channel
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: pointer to UserOnboard params
	//
	// Returns:
	//   - store.ErrInsertFailed: if onboard fails
	// 	 - store.ErrUniqueViolation: if email already exists
	//	 - store.ErrNotNullViolation: if a required field is null
	Onboard(ctx context.Context, arg *params.UserOnboard) (int64, error)

	// Gets onboard request by token
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - token: token of onboard request
	//
	// Returns:
	//   - store.ErrRecordNotFound: if request not found
	//	 - store.ErrQueryFailed: if any other error occurs
	GetOnboardReqByToken(ctx context.Context, token string) (*models.UserOnboardReq, error)

	// Deletes onboard request by id
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - id: id of onboard request
	//
	// Returns:
	//   - store.ErrRecordNotFound: if request not found
	//	 - store.ErrQueryFailed: if any other error occurs
	DeleteOnboardReqById(ctx context.Context, id int64) error

	// Verifies an onboard request
	//	 - sets user's hashed password
	// 	 - deletes onboard request
	//
	// Parameters:
	//   - ctx: context to execute the query
	//   - arg: a pointer to UserOnboardVerify params
	//
	// Returns:
	//   - store.ErrUpdateFailed: if set password fails
	//	 - store.ErrDeleteFailed: if onboard request delete fails
	//	 - store.ErrQueryFailed: if transaction fails
	OnboardVerify(ctx context.Context, arg *params.UserOnboardVerify) error

	GetAll(ctx context.Context, arg *params.UserSearch) ([]agg.User, error)
}
