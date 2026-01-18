package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mizu/internal/db/models"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strings"

	"github.com/mattn/go-sqlite3"
)

// SQLite implementation of UserStore interface
type UserStore struct {
	db *sql.DB
}

// Creates a new instance of a UserStore
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *UserStore
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Implementation of CreateOne defined in UserStore interface
func (q *UserStore) CreateOne(ctx context.Context, arg *params.UserCreate) (int64, error) {

	query := `
	INSERT INTO users(first_name, last_name, title, email, image, is_active, is_verified, role)
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
		arg.IsVerified,
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

// Implementation of GetById defined in UserStore interface
func (q *UserStore) GetById(ctx context.Context, id int64) (*models.User, error) {

	query := `
	SELECT id, first_name, last_name, title, email, image, is_active, is_verified, role, 
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
		&user.IsVerified,
		&user.Role,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		return &user, store.ErrRecordNotFound
	}

	return &user, nil
}

// Implementation of GetByEmail defined in UserStore interface
func (q *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {

	query := `
	SELECT id, first_name, last_name, title, email, image, is_active, is_verified, role, 
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
		&user.IsVerified,
		&user.Role,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		log.Println(err)
		return &user, store.ErrRecordNotFound
	}

	return &user, nil
}

// Implementation of GetPasswordById defined in UserStore interface
func (q *UserStore) GetPasswordById(ctx context.Context, id int64) (string, error) {

	query := `SELECT password FROM users WHERE id = ?`

	var password string

	err := q.db.QueryRowContext(ctx, query, id).Scan(&password)

	if err != nil {
		return password, store.ErrRecordNotFound
	}

	return password, nil
}

// Implementation of CountAll defined in UserStore interface
func (q *UserStore) CountAll(ctx context.Context, arg *params.UserSearch) (int64, error) {

	var query strings.Builder
	query.WriteString(`
	SELECT 
		COUNT(*) 
	FROM users u 
	JOIN roles r ON r.id = u.role
	`)

	queryArgs := []any{}

	var conditions []string

	if arg != nil {
		if len(arg.Keyword) > 0 {
			conditions = append(conditions, "( u.first_name LIKE ? OR u.last_name LIKE ? OR u.title LIKE ? )")
			keyword := "%" + arg.Keyword + "%"
			queryArgs = append(queryArgs, keyword, keyword, keyword)
		}

		if len(arg.Role) > 0 {
			conditions = append(conditions, "r.name = ?")
			queryArgs = append(queryArgs, arg.Role)
		}

		if len(conditions) > 0 {
			query.WriteString("WHERE ")
			query.WriteString(strings.Join(conditions, " AND "))
		}
	}

	var count int64
	err := q.db.QueryRowContext(ctx, query.String(), queryArgs...).Scan(&count)

	if err != nil {
		return 0, store.ErrQueryFailed
	}

	return count, nil
}

// Implementation of SetPasswordById defined in UserStore interface
func (q *UserStore) SetPasswordById(ctx context.Context, id int64, password string) error {

	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, password, id)

	if err != nil {
		return store.ErrUpdateFailed
	}

	return nil
}

// Implementation of DeleteById defined in UserStore interface
func (q *UserStore) DeleteById(ctx context.Context, id int64) error {

	query := `DELETE FROM users WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)

	if err != nil {
		return store.ErrDeleteFailed
	}

	return nil
}

// Implementation of UpdateLastLogin defined in UserStore interface
func (q *UserStore) UpdateLastLogin(ctx context.Context, id int64) error {

	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := q.db.ExecContext(ctx, query, id)
	if err != nil {
		return store.ErrUpdateFailed
	}

	return nil
}

// Implementation of GetRoleById defined in UserStore interface
func (q *UserStore) GetRoleById(ctx context.Context, id int64) (*models.Role, error) {

	query := `SELECT id, name FROM roles WHERE id = ?`

	role := models.Role{}

	err := q.db.QueryRowContext(ctx, query, id).Scan(&role.Id, &role.Name)
	if err != nil {
		return nil, store.ErrRecordNotFound
	}

	return &role, nil
}

// Implementation of CreateOnboardReq defined in UserStore interface
func (q *UserStore) CreateOnboardReq(ctx context.Context, arg *params.UserOnboardReqCreate) error {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return store.ErrInsertFailed
	}

	defer tx.Rollback()

	deleteQuery := `DELETE FROM user_onboard_reqs WHERE user_id = ?`

	_, err = tx.ExecContext(ctx, deleteQuery, arg.UserId)
	if err != nil {
		return store.ErrInsertFailed
	}

	insertQuery := `INSERT INTO user_onboard_reqs(user_id, token, is_valid) VALUES(?, ?, ?)`

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
func (q *UserStore) Onboard(ctx context.Context, arg *params.UserOnboard) (int64, error) {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertUserQuery := `
	INSERT INTO users(first_name, last_name, title, email, image, is_active, is_verified, role)
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
		arg.IsVerified,
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

	if len(arg.Projects) > 0 {
		var assignProjectsQuery strings.Builder
		assignProjectsQuery.WriteString("INSERT INTO project_users(user_id, project_id)")
		projectsArgs := []any{}
		valueArgs := []string{}
		for _, val := range arg.Projects {
			valueArgs = append(valueArgs, " VALUES(?, ?)")
			projectsArgs = append(projectsArgs, userId, val)
		}

		assignProjectsQuery.WriteString(strings.Join(valueArgs, ","))

		_, err = tx.ExecContext(ctx, assignProjectsQuery.String(), projectsArgs...)
		if err != nil {
			if sqlite3Err, ok := err.(sqlite3.Error); ok {
				if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
					return 0, store.ErrForeignKeyViolation
				}
			}

			return 0, store.ErrInsertFailed
		}
	}

	insertOnboardRequestQuery := `
	INSERT INTO user_onboard_reqs(user_id, token, is_valid) VALUES(?, ?, ?)
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

// Implementation of GetOnboardReqByToken defined in UserStore interface
func (q *UserStore) GetOnboardReqByToken(ctx context.Context, token string) (*models.UserOnboardReq, error) {

	query := `SELECT id, user_id, token, issued_at, is_valid FROM user_onboard_reqs WHERE token = ?`

	var verifyRequest models.UserOnboardReq
	if err := q.db.QueryRowContext(ctx, query, token).Scan(
		&verifyRequest.Id,
		&verifyRequest.UserId,
		&verifyRequest.Token,
		&verifyRequest.IssuedAt,
		&verifyRequest.IsValid,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrRecordNotFound
		}

		return nil, store.ErrQueryFailed
	}

	return &verifyRequest, nil
}

// Implementation of DeleteOnboardReqById defined in UserStore interface
func (q *UserStore) DeleteOnboardReqById(ctx context.Context, id int64) error {

	query := `DELETE FROM user_onboard_reqs WHERE id = ?`

	if _, err := q.db.ExecContext(ctx, query, id); err != nil {
		return store.ErrDeleteFailed
	}

	return nil
}

// Implementation of OnboardVerify defined in UserStore interface
func (q *UserStore) OnboardVerify(ctx context.Context, arg *params.UserOnboardVerify) error {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return store.ErrQueryFailed
	}

	defer tx.Rollback()

	setPasswordQuery := `UPDATE users SET password = ?, is_verified = 1 WHERE id = ?`
	if _, err := tx.ExecContext(ctx, setPasswordQuery, arg.HashedPassword, arg.UserId); err != nil {
		return store.ErrUpdateFailed
	}

	deleteTokenQuery := `DELETE FROM user_onboard_reqs WHERE user_id = ?`
	if _, err := tx.ExecContext(ctx, deleteTokenQuery, arg.UserId); err != nil {
		return store.ErrDeleteFailed
	}

	if err = tx.Commit(); err != nil {
		return store.ErrQueryFailed
	}

	return nil
}

func (q *UserStore) GetAll(ctx context.Context, arg *params.UserSearch) ([]agg.User, error) {

	var query strings.Builder
	query.WriteString(`
	SELECT 
		u.id,
		u.first_name,
		u.last_name,
		u.email,
		r.name AS role,
		u.title,
		u.image,
		u.created_at,
		u.last_login,
		u.is_active,
		u.is_verified
	FROM users u
	JOIN roles r ON r.id = u.role
	`)

	queryArgs := []any{}

	var conditions []string

	if len(arg.Keyword) > 0 {
		conditions = append(conditions, "( u.first_name LIKE ? OR u.last_name LIKE ? OR u.title LIKE ? )")
		keyword := "%" + arg.Keyword + "%"
		queryArgs = append(queryArgs, keyword, keyword, keyword)
	}

	if len(arg.Role) > 0 {
		conditions = append(conditions, "r.name = ?")
		queryArgs = append(queryArgs, arg.Role)
	}

	if len(conditions) > 0 {
		query.WriteString("WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(" LIMIT ? OFFSET ?")
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.User, 0)

	for rows.Next() {
		var row agg.User
		err := rows.Scan(
			&row.Id,
			&row.FirstName,
			&row.LastName,
			&row.Email,
			&row.Role,
			&row.Title,
			&row.Image,
			&row.CreatedAt,
			&row.LastLogin,
			&row.IsActive,
			&row.IsVerified,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}

func (q *UserStore) SetActive(ctx context.Context, userID int64, isActive bool) error {

	query := `UPDATE users SET is_active = ? WHERE id = ?`

	if _, err := q.db.ExecContext(ctx, query, isActive, userID); err != nil {
		return store.ErrUpdateFailed
	}

	return nil
}
