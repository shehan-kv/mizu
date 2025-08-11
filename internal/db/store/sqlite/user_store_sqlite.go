package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/models"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/mattn/go-sqlite3"
)

// SQLite implementation of UserStore interface
type UserStoreSqlite struct {
	db *sql.DB
}

// Creates a new instance of a UserStoreSqlite
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *UserStoreSqlite
func NewUserStore(db *sql.DB) *UserStoreSqlite {
	return &UserStoreSqlite{db: db}
}

// Creates one user
//
// Parameters:
//   - ctx: context to execute the query
//   - arg: pointer to UserCreateParams
//
// Returns:
//   - int64: id of new user
//   - store.ErrUniqueViolation: if email already exists
//   - store.ErrInsertFailed: if create fails
func (q *UserStoreSqlite) CreateOne(ctx context.Context, arg *params.UserCreate) (int64, error) {

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
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, store.ErrUniqueViolation
			}
		}
		return 0, store.ErrInsertFailed
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
//   - store.ErrRecordNotFound: if not found
func (q *UserStoreSqlite) GetById(ctx context.Context, id int64) (*models.User, error) {

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
		return &user, store.ErrRecordNotFound
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
//   - store.ErrRecordNotFound: if not found
func (q *UserStoreSqlite) GetByEmail(ctx context.Context, email string) (*models.User, error) {

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
		return &user, store.ErrRecordNotFound
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
//   - store.ErrRecordNotFound: if not found
func (q *UserStoreSqlite) GetPasswordById(ctx context.Context, id int64) (string, error) {

	query := `SELECT password FROM users WHERE id = ?`

	var password string

	err := q.db.QueryRowContext(ctx, query, id).Scan(&password)

	if err != nil {
		return password, store.ErrRecordNotFound
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
//   - store.ErrQueryFailed: if query fails
func (q *UserStoreSqlite) CountAll(ctx context.Context) (int64, error) {

	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := q.db.QueryRowContext(ctx, query).Scan(&count)

	if err != nil {
		return 0, store.ErrQueryFailed
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
//   - store.ErrUpdateFailed: if update fails
func (q *UserStoreSqlite) SetPasswordById(ctx context.Context, id int64, password string) error {

	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, password, id)

	if err != nil {
		return store.ErrUpdateFailed
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
//   - store.ErrDeleteFailed: if delete fails
func (q *UserStoreSqlite) DeleteById(ctx context.Context, id int64) error {

	query := `DELETE FROM users WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)

	if err != nil {
		return store.ErrDeleteFailed
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
//   - store.ErrUpdateFailed: if update fails
func (q *UserStoreSqlite) UpdateLastLogin(ctx context.Context, id int64) error {

	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)
	if err != nil {
		return store.ErrUpdateFailed
	}

	return nil
}

func (q *UserStoreSqlite) GetRoleById(ctx context.Context, id int64) (*models.Role, error) {

	query := `SELECT id, name FROM roles WHERE id = ?`

	role := models.Role{}

	err := q.db.QueryRowContext(ctx, query, id).Scan(&role.Id, &role.Name)
	if err != nil {
		return nil, store.ErrRecordNotFound
	}

	return &role, nil
}

// Implementation of CreateVerifyRequest defined in UserStore interface
func (q *UserStoreSqlite) CreateVerifyRequest(ctx context.Context, arg *params.UserVerifyRequestCreate) error {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return store.ErrInsertFailed
	}

	defer tx.Rollback()

	deleteQuery := `DELETE FROM user_verify_requests WHERE user_id = ?`

	_, err = tx.ExecContext(ctx, deleteQuery, arg.UserId)
	if err != nil {
		return store.ErrInsertFailed
	}

	insertQuery := `INSERT INTO user_verify_requests(user_id, token, is_valid) VALUES(?, ?, ?)`

	_, err = tx.ExecContext(ctx, insertQuery, arg.UserId, arg.Token, arg.IsValid)
	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return store.ErrNotNullViolation
			}
		}

		return store.ErrInsertFailed
	}

	if err = tx.Commit(); err != nil {
		return store.ErrInsertFailed
	}

	return nil
}

// Implementation of Onboard defined in UserStore interface
func (q *UserStoreSqlite) Onboard(ctx context.Context, arg *params.UserOnboard) (int64, error) {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertUserQuery := `
	INSERT INTO users(first_name, last_name, title, email, image, is_active, role)
	VALUES(?,?,?,?,?,?, (SELECT id FROM roles WHERE name = ?)) RETURNING id
	`

	var userId int64
	err = tx.QueryRowContext(ctx, insertUserQuery,
		arg.FirstName,
		arg.LastName,
		arg.Title,
		arg.Email,
		arg.Image,
		arg.IsActive,
		arg.Role).Scan(&userId)
	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertOnboardRequestQuery := `
	INSERT INTO user_verify_requests(user_id, token, is_valid) VALUES(?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, insertOnboardRequestQuery, userId, arg.Token, true)
	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}
		}

		return 0, store.ErrInsertFailed
	}

	insertChannelQuery := `INSERT INTO channels(name) VALUES(?) RETURNING id`

	var channelId int64
	err = tx.QueryRowContext(ctx, insertChannelQuery, "general").Scan(&channelId)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	insertUsersQuery := `INSERT INTO channel_users(channel_id, user_id) VALUES(?, ?)`
	for _, userId := range []int64{arg.ActorID, userId} {
		_, err := tx.ExecContext(ctx, insertUsersQuery, channelId, userId)
		if err != nil {
			if sqlite3Err, ok := err.(sqlite3.Error); ok {
				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
					return 0, store.ErrNotNullViolation
				}
			}

			return 0, store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return userId, nil
}
