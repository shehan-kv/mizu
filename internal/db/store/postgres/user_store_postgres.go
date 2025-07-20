package postgres

import (
	"context"
	"database/sql"
	"mizu/internal/db/models"
	"mizu/internal/db/params"
	"mizu/internal/errdefs"

	"github.com/lib/pq"
)

// Postgres implementation of UserStore interface
type UserStorePostgres struct {
	db *sql.DB
}

// Creates a new instance of a UserStorePostgres
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *UserStorePostgres
func NewUserStore(db *sql.DB) *UserStorePostgres {
	return &UserStorePostgres{db: db}
}

// Creates one user
//
// Parameters:
//   - ctx: context to execute the query
//   - arg: pointer to UserCreateParams
//
// Returns:
//   - int64: id of new user
//   - errdefs.ErrDbUniqueViolation if email already exists
//   - errdefs.ErrDbInsertFailed if create fails
func (q *UserStorePostgres) CreateOne(ctx context.Context, arg *params.UserCreateParams) (int64, error) {

	query := `
	INSERT INTO users(first_name, last_name, title, email, image, is_active, role)
	VALUES(?,?,?,?,?,?, (SELECT id FROM roles WHERE name = ?)) RETURNING id
	`

	var id int64 = 0
	err := q.db.QueryRowContext(
		ctx,
		query,
		arg.FirstName,
		arg.LastName,
		arg.Title,
		arg.Email,
		arg.Image,
		arg.IsActive,
		arg.Role,
	).Scan(&id)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return 0, errdefs.ErrDbUniqueViolation
			}

			return 0, errdefs.ErrDbInsertFailed
		}
	}

	return id, nil
}

// Gets a user by id
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//
// Returns:
//   - *models.User: reference to a models.User instace
//   - errdefs.ErrDbRecordNotFound if not found
func (q *UserStorePostgres) GetById(ctx context.Context, id int64) (*models.User, error) {

	query := `
	SELECT id, first_name, last_name, title, email, image, is_active, role, 
	created_at, last_login FROM users WHERE id = ?
	`

	var user models.User

	err := q.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Title,
		&user.Email,
		&user.Image,
		&user.IsActive,
		&user.Role,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		return &user, errdefs.ErrDbRecordNotFound
	}

	return &user, nil
}

// Gets a user by email
//
// Parameters:
//   - ctx: context to execute the query
//   - email: email of the user
//
// Returns:
//   - *models.User: reference to a models.User instace
//   - errdefs.ErrDbRecordNotFound if not found
func (q *UserStorePostgres) GetByEmail(ctx context.Context, email string) (*models.User, error) {

	query := `
	SELECT id, first_name, last_name, title, email, image, is_active, role, 
	created_at, last_login FROM users WHERE email = ?
	`

	var user models.User

	err := q.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Title,
		&user.Email,
		&user.Image,
		&user.IsActive,
		&user.Role,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		return &user, errdefs.ErrDbRecordNotFound
	}

	return &user, nil
}

// Gets user password id
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//
// Returns:
//   - string: user's password
//   - errdefs.ErrDbRecordNotFound if not found
func (q *UserStorePostgres) GetPasswordById(ctx context.Context, id int64) (string, error) {

	query := `SELECT password FROM users WHERE id = ?`

	var password string

	err := q.db.QueryRowContext(ctx, query, id).Scan(&password)

	if err != nil {
		return password, errdefs.ErrDbRecordNotFound
	}

	return password, nil
}

// Counts all users
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//
// Returns:
//   - int64: number of users
//   - errdefs.ErrDbQueryFailed if query fails
func (q *UserStorePostgres) CountAll(ctx context.Context) (int64, error) {

	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := q.db.QueryRowContext(ctx, query).Scan(&count)

	if err != nil {
		return 0, errdefs.ErrDbQueryFailed
	}

	return count, nil
}

// Sets password by user id
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//   - password: new password to set
//
// Returns:
//   - errdefs.ErrDbUpdateFailed if update fails
func (q *UserStorePostgres) SetPasswordById(ctx context.Context, id int64, password string) error {

	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, password, id)

	if err != nil {
		return errdefs.ErrDbUpdateFailed
	}

	return nil
}

// Deletes user by id
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//
// Returns:
//   - errdefs.ErrDbDeleteFailed if delete fails
func (q *UserStorePostgres) DeleteById(ctx context.Context, id int64) error {

	query := `DELETE FROM users WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)

	if err != nil {
		return errdefs.ErrDbDeleteFailed
	}

	return nil
}

// Updates last login by user id
//
// Parameters:
//   - ctx: context to execute the query
//   - id: id of the user
//
// Returns:
//   - errdefs.ErrDbUpdateFailed if update fails
func (q *UserStorePostgres) UpdateLastLogin(ctx context.Context, id int64) error {

	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)
	if err != nil {
		return errdefs.ErrDbUpdateFailed
	}

	return nil
}
