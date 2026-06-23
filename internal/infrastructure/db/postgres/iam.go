package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type IAMRepository struct {
	db *sql.DB
}

func NewIAMRepository(db *sql.DB) *IAMRepository {
	return &IAMRepository{
		db: db,
	}
}

func (r *IAMRepository) Add(ctx context.Context, user *iam.User) error {

	query := `
		INSERT INTO users(
			id,
			first_name,
			last_name,
			title,
			email,
			image,
			image_mime,
			is_active,
			is_verified,
			role,
			version,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var img *string
	var mime *string
	if user.Image() != nil {
		n := user.Image().Name().String()
		img = &n

		m := user.Image().MimeType().String()
		mime = &m
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		user.FirstName(),
		user.LastName(),
		user.Title(),
		user.Email().String(),
		img,
		mime,
		user.IsActive(),
		user.IsVerified(),
		user.Role().String(),
		user.Version(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)

	if err != nil {

		if pqErr, ok := errors.AsType[*pq.Error](err); ok {
			if pqErr.Code == "23505" { // unique_violation

				switch pqErr.Constraint {
				case "users_pk":
					return fmt.Errorf("iam.IAMRepository.Add: duplicate user id: %w", err)

				case "users_email_unique":
					return fmt.Errorf("%w: %w", iam.ErrUserEmailAlreadyExists, err)

				default:
					return fmt.Errorf("iam.IAMRepository.Add: unique constraint violation: %w", err)
				}
			}
		}

		return fmt.Errorf("iam.IAMRepository.Add: %w", err)
	}

	return nil
}

func (r *IAMRepository) Exists(ctx context.Context, id iam.UserID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		id.String(),
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.Exists: %w", err)
	}

	return exists, nil
}

func (r *IAMRepository) ExistsAll(ctx context.Context, ids []iam.UserID) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	seen := make(map[iam.UserID]struct{}, len(ids))
	for _, id := range ids {
		seen[id] = struct{}{}
	}

	uids := make([]string, 0, len(seen))
	for id := range seen {
		uids = append(uids, id.String())
	}

	query := `
		SELECT COUNT(DISTINCT id)
		FROM users
		WHERE id = ANY($1::uuid[])
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		pq.Array(uids),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.ExistsAll: %w", err)
	}

	return count == len(seen), nil
}

func (r *IAMRepository) GetByID(ctx context.Context, id iam.UserID) (*iam.User, error) {
	query := `
		SELECT
			id,
			first_name,
			last_name,
			role,
			title,
			email,
			image,
			image_mime,
			is_active,
			is_verified,
			last_signin_at,
			version,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id.String())

	var (
		rawID        string
		firstName    string
		lastName     string
		role         string
		title        *string
		email        string
		image        *string
		imageMime    *string
		isActive     bool
		isVerified   bool
		lastSignInAt *time.Time
		version      int
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(
		&rawID,
		&firstName,
		&lastName,
		&role,
		&title,
		&email,
		&image,
		&imageMime,
		&isActive,
		&isVerified,
		&lastSignInAt,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
	}

	userID, err := iam.NewUserID(rawID)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
	}

	emailVO, err := iam.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
	}

	name, err := iam.NewName(firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
	}

	roleVO, err := iam.NewRole(role)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
	}

	var imageVO *iam.Image
	if image != nil && imageMime != nil {
		imgName, err := iam.NewImageName(*image)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
		}

		imgMime, err := iam.NewMimeType(*imageMime)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.GetByID: %w", err)
		}

		img := iam.NewImage(imgName, imgMime)
		imageVO = &img
	}

	user := iam.RestoreUser(
		userID,
		name,
		emailVO,
		title,
		roleVO,
		imageVO,
		isActive,
		isVerified,
		lastSignInAt,
		version,
		createdAt,
		updatedAt,
	)

	return user, nil
}

func (r *IAMRepository) GetCredentialsByEmail(ctx context.Context, email iam.Email) (*iam.UserCredentials, error) {

	query := `
		SELECT
			id,
			first_name,
			last_name,
			role,
			title,
			password,
			image,
			image_mime,
			is_active,
			is_verified,
			last_signin_at,
			version,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email.String())

	var (
		rawID        string
		firstName    string
		lastName     string
		role         string
		title        *string
		password     string
		image        *string
		imageMime    *string
		isActive     bool
		isVerified   bool
		lastSignInAt *time.Time
		version      int
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(
		&rawID,
		&firstName,
		&lastName,
		&role,
		&title,
		&password,
		&image,
		&imageMime,
		&isActive,
		&isVerified,
		&lastSignInAt,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrUserNotFound, err)
		}
		return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
	}

	userID, err := iam.NewUserID(rawID)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
	}

	name, err := iam.NewName(firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
	}

	roleVO, err := iam.NewRole(role)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
	}

	var imageVO *iam.Image
	if image != nil && imageMime != nil {
		imgName, err := iam.NewImageName(*image)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
		}

		imgMime, err := iam.NewMimeType(*imageMime)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.GetCredentialsByEmail: %w", err)
		}

		img := iam.NewImage(imgName, imgMime)
		imageVO = &img
	}

	user := iam.RestoreUser(
		userID,
		name,
		email,
		title,
		roleVO,
		imageVO,
		isActive,
		isVerified,
		lastSignInAt,
		version,
		createdAt,
		updatedAt,
	)

	return iam.NewUserCredentials(user, password), nil
}

func (r *IAMRepository) List(ctx context.Context, filter iam.UserFilter, page common.Page) ([]*iam.User, error) {

	var sb strings.Builder
	args := make([]any, 0)
	argPos := 1

	sb.WriteString(`
		SELECT
			id,
			first_name,
			last_name,
			role,
			title,
			email,
			image,
			image_mime,
			is_active,
			is_verified,
			last_signin_at,
			version,
			created_at,
			updated_at
		FROM users
		WHERE 1=1
	`)

	if filter.Keyword != nil {
		keyword := "%" + *filter.Keyword + "%"

		sb.WriteString(" AND (first_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(" OR last_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos + 1))
		sb.WriteString(" OR email ILIKE $")
		sb.WriteString(strconv.Itoa(argPos + 2))
		sb.WriteString(")")

		args = append(args, keyword, keyword, keyword)
		argPos += 3
	}

	if filter.Role != nil {
		sb.WriteString(" AND role = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, filter.Role.String())
		argPos++
	}

	if filter.IsActive != nil {
		sb.WriteString(" AND is_active = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, *filter.IsActive)
		argPos++
	}

	if filter.IsVerified != nil {
		sb.WriteString(" AND is_verified = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, *filter.IsVerified)
		argPos++
	}

	sb.WriteString(" ORDER BY created_at DESC LIMIT $")
	sb.WriteString(strconv.Itoa(argPos))
	sb.WriteString(" OFFSET $")
	sb.WriteString(strconv.Itoa(argPos + 1))

	args = append(args, page.Limit(), page.Offset())

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
	}
	defer rows.Close()

	users := make([]*iam.User, 0)

	for rows.Next() {
		var (
			rawID        string
			firstName    string
			lastName     string
			role         string
			title        *string
			email        string
			image        *string
			imageMime    *string
			isActive     bool
			isVerified   bool
			lastSignInAt *time.Time
			version      int
			createdAt    time.Time
			updatedAt    time.Time
		)

		err := rows.Scan(
			&rawID,
			&firstName,
			&lastName,
			&role,
			&title,
			&email,
			&image,
			&imageMime,
			&isActive,
			&isVerified,
			&lastSignInAt,
			&version,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
		}

		userID, err := iam.NewUserID(rawID)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
		}

		emailVO, err := iam.NewEmail(email)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
		}

		name, err := iam.NewName(firstName, lastName)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
		}

		roleVO, err := iam.NewRole(role)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
		}

		var imageVO *iam.Image
		if image != nil && imageMime != nil {
			imgName, err := iam.NewImageName(*image)
			if err != nil {
				return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
			}

			imgMime, err := iam.NewMimeType(*imageMime)
			if err != nil {
				return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
			}

			img := iam.NewImage(imgName, imgMime)
			imageVO = &img
		}

		user := iam.RestoreUser(
			userID,
			name,
			emailVO,
			title,
			roleVO,
			imageVO,
			isActive,
			isVerified,
			lastSignInAt,
			version,
			createdAt,
			updatedAt,
		)

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.List: %w", err)
	}

	return users, nil
}

func (r *IAMRepository) ListByIDs(ctx context.Context, ids []iam.UserID, f iam.UserFilter) ([]*iam.User, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(ids)+10)
	argPos := 1

	sb.WriteString(`
		SELECT
			id,
			first_name,
			last_name,
			role,
			title,
			email,
			image,
			image_mime,
			is_active,
			is_verified,
			last_signin_at,
			version,
			created_at,
			updated_at
		FROM users
		WHERE id = ANY($1)
	`)

	uuidList := make([]string, 0, len(ids))
	for _, id := range ids {
		uuidList = append(uuidList, id.String())
	}

	args = append(args, pq.Array(uuidList))
	argPos++

	if f.Keyword != nil && strings.TrimSpace(*f.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(*f.Keyword) + "%"

		sb.WriteString(" AND (")

		sb.WriteString("first_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(" OR last_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(" OR title ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(" OR (first_name || ' ' || last_name) ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))

		sb.WriteString(")")

		args = append(args, keyword)
		argPos++
	}

	if f.Role != nil {
		sb.WriteString(" AND role = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Role.String())
		argPos++
	}

	if f.IsActive != nil {
		sb.WriteString(" AND is_active = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, *f.IsActive)
		argPos++
	}

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
	}
	defer rows.Close()

	users := make([]*iam.User, 0)

	for rows.Next() {
		var (
			rawID        string
			firstName    string
			lastName     string
			role         string
			title        *string
			email        string
			image        *string
			imageMime    *string
			isActive     bool
			isVerified   bool
			lastSignInAt *time.Time
			version      int
			createdAt    time.Time
			updatedAt    time.Time
		)

		err := rows.Scan(
			&rawID,
			&firstName,
			&lastName,
			&role,
			&title,
			&email,
			&image,
			&imageMime,
			&isActive,
			&isVerified,
			&lastSignInAt,
			&version,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
		}

		userID, err := iam.NewUserID(rawID)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
		}

		emailVO, err := iam.NewEmail(email)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
		}

		name, err := iam.NewName(firstName, lastName)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
		}

		roleVO, err := iam.NewRole(role)
		if err != nil {
			return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
		}

		var imageVO *iam.Image
		if image != nil && imageMime != nil {
			imgName, err := iam.NewImageName(*image)
			if err != nil {
				return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
			}

			imgMime, err := iam.NewMimeType(*imageMime)
			if err != nil {
				return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
			}

			img := iam.NewImage(imgName, imgMime)
			imageVO = &img
		}

		users = append(users, iam.RestoreUser(
			userID,
			name,
			emailVO,
			title,
			roleVO,
			imageVO,
			isActive,
			isVerified,
			lastSignInAt,
			version,
			createdAt,
			updatedAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iam.IAMRepository.ListByIDs: %w", err)
	}

	return users, nil
}

func (r *IAMRepository) IsAnyAdministrator(ctx context.Context, ids []iam.UserID) (bool, error) {

	if len(ids) == 0 {
		return false, nil
	}

	var sb strings.Builder
	args := make([]any, 0, 2)

	sb.WriteString(`
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE role = $1
			AND id = ANY($2)
		)
	`)

	args = append(args, iam.RoleAdministrator.String())

	uuidList := make([]string, 0, len(ids))
	for _, id := range ids {
		uuidList = append(uuidList, id.String())
	}

	args = append(args, pq.Array(uuidList))

	var exists bool
	err := r.db.QueryRowContext(ctx, sb.String(), args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.IsAnyAdministrator: %w", err)
	}

	return exists, nil
}

func (r *IAMRepository) HasAdministrator(ctx context.Context) (bool, error) {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE role = $1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		query,
		iam.RoleAdministrator.String(),
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.HasAdministrator: %w", err)
	}

	return exists, nil
}

func (r *IAMRepository) SetPassword(ctx context.Context, user *iam.User, hash string) error {

	query := `
		UPDATE users
		SET password = $1,
		    version = version + 1
		WHERE id = $2 AND version = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		hash,
		user.ID().String(),
		user.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.SetPassword: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.SetPassword: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"%w: user id %s not found or version mismatch",
			iam.ErrUserNotFound,
			user.ID(),
		)
	}

	return nil
}

func (r *IAMRepository) Count(ctx context.Context, filter iam.UserFilter) (int, error) {

	var sb strings.Builder
	args := make([]any, 0)
	argPos := 1

	sb.WriteString(`
		SELECT COUNT(*)
		FROM users
		WHERE 1=1
	`)

	if filter.Keyword != nil && strings.TrimSpace(*filter.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(*filter.Keyword) + "%"

		sb.WriteString(" AND (first_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(" OR last_name ILIKE $")
		sb.WriteString(strconv.Itoa(argPos + 1))
		sb.WriteString(" OR email ILIKE $")
		sb.WriteString(strconv.Itoa(argPos + 2))
		sb.WriteString(")")

		args = append(args, keyword, keyword, keyword)
		argPos += 3
	}

	if filter.Role != nil {
		sb.WriteString(" AND role = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, filter.Role.String())
		argPos++
	}

	if filter.IsActive != nil {
		sb.WriteString(" AND is_active = $")
		sb.WriteString(strconv.Itoa(argPos))

		args = append(args, *filter.IsActive)
		argPos++
	}

	var count int

	err := r.db.QueryRowContext(ctx, sb.String(), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("iam.IAMRepository.Count: %w", err)
	}

	return count, nil
}

func (r *IAMRepository) Save(ctx context.Context, user *iam.User) error {

	const query = `
		UPDATE users
		SET
			first_name = $1,
			last_name = $2,
			role = $3,
			title = $4,
			email = $5,
			image = $6,
			image_mime = $7,
			is_active = $8,
			is_verified = $9,
			last_signin_at = $10,
			version = version + 1,
			updated_at = $11
		WHERE id = $12
		  AND version = $13
	`

	var img *string
	var mime *string

	if user.Image() != nil {
		n := user.Image().Name().String()
		img = &n

		m := user.Image().MimeType().String()
		mime = &m
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.FirstName(),
		user.LastName(),
		user.Role().String(),
		user.Title(),
		user.Email().String(),
		img,
		mime,
		user.IsActive(),
		user.IsVerified(),
		user.LastSignInAt(),
		user.UpdatedAt(),
		user.ID().String(),
		user.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.Save: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.Save: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"%w: optimistic lock failed for user %s",
			iam.ErrUserConcurrentModification,
			user.ID(),
		)
	}

	return nil
}

func (r *IAMRepository) Remove(ctx context.Context, user *iam.User) error {

	query := `
		DELETE FROM users
		WHERE id = $1 AND version = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		user.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.Remove: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.Remove: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"%w: user id %s",
			iam.ErrUserConcurrentModification,
			user.ID(),
		)
	}

	return nil
}
