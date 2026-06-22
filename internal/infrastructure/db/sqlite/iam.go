package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
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
		is_active, 
		is_verified, 
		role, 
		version,
		created_at, 
		updated_at
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		user.FirstName(),
		user.LastName(),
		user.Title(),
		user.Email().String(),
		user.Image(),
		user.IsActive(),
		user.IsVerified(),
		user.Role().String(),
		user.Version(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)

	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf("iam.IAMRepository.Add: duplicate user id: %w", err)
			}
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return fmt.Errorf("%w: %w", iam.ErrUserEmailAlreadyExists, err)
			}
		}
		return fmt.Errorf("iam.IAMRepository.Add: %w", err)
	}

	return nil
}

func (r *IAMRepository) Exists(ctx context.Context, id iam.UserID) (bool, error) {

	query := `
    SELECT EXISTS (
        SELECT 1 FROM users WHERE id = ?
    )
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&exists)
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

	args := make([]any, 0, len(seen))
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(DISTINCT id) FROM users WHERE id IN (")
	i := 0
	for id := range seen {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
		args = append(args, id.String())
		i++
	}
	sb.WriteString(")")

	var count int
	err := r.db.QueryRowContext(ctx, sb.String(), args...).Scan(&count)
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
    WHERE id = ?
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
    WHERE email = ?
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
		sb.WriteString(" AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)")
		keyword := "%" + *filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	if filter.Role != nil {
		sb.WriteString(" AND role = ?")
		args = append(args, filter.Role.String())
	}

	if filter.IsActive != nil {
		sb.WriteString(" AND is_active = ?")
		args = append(args, *filter.IsActive)
	}

	if filter.IsVerified != nil {
		sb.WriteString(" AND is_verified = ?")
		args = append(args, *filter.IsVerified)
	}

	sb.WriteString(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
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
	args := make([]any, 0, len(ids)+3)

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
		WHERE id IN (`)

	for i, id := range ids {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
		args = append(args, id.String())
	}

	sb.WriteString(")")

	if f.Keyword != nil && strings.TrimSpace(*f.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(*f.Keyword) + "%"

		sb.WriteString(`
			AND (
				first_name LIKE ?
				OR last_name LIKE ?
				OR title LIKE ?
				OR (first_name || ' ' || last_name) LIKE ?
			)`)

		args = append(args, keyword, keyword, keyword, keyword)
	}

	if f.Role != nil {
		sb.WriteString(` AND role = ?`)
		args = append(args, f.Role.String())
	}

	if f.IsActive != nil {
		sb.WriteString(` AND is_active = ?`)
		args = append(args, *f.IsActive)
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
	args := make([]any, 0, len(ids))

	sb.WriteString(`
    SELECT EXISTS (
        SELECT 1 FROM users
        WHERE role = ?
        AND id IN (`)
	args = append(args, iam.RoleAdministrator.String())

	for i := range ids {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
		args = append(args, ids[i].String())
	}
	sb.WriteString("))")

	var exists bool
	err := r.db.QueryRowContext(ctx, sb.String(), args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.IsAnyAdministrator: %w", err)
	}

	return exists, nil
}

func (r *IAMRepository) HasAdministrator(ctx context.Context) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE role = ?
			LIMIT 1
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, iam.RoleAdministrator.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("iam.IAMRepository.HasAdministrator: %w", err)
	}

	return exists, nil
}

func (r *IAMRepository) SetPassword(ctx context.Context, user *iam.User, hash string) error {
	query := `
		UPDATE users
		SET password = ?, version = version + 1
		WHERE id = ? AND version = ?
	`

	result, err := r.db.ExecContext(ctx, query, hash, user.ID().String(), user.Version())
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.SetPassword: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.IAMRepository.SetPassword: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: user id %s not found or version mismatch", iam.ErrUserNotFound, user.ID())
	}

	return nil
}

func (r *IAMRepository) Count(ctx context.Context, filter iam.UserFilter) (int, error) {
	var sb strings.Builder
	args := make([]any, 0)

	sb.WriteString(`
    SELECT COUNT(*)
    FROM users
    WHERE 1=1
`)

	if filter.Keyword != nil {
		sb.WriteString(" AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)")
		keyword := "%" + *filter.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	if filter.Role != nil {
		sb.WriteString(" AND role = ?")
		args = append(args, filter.Role.String())
	}

	if filter.IsActive != nil {
		sb.WriteString(" AND is_active = ?")
		args = append(args, *filter.IsActive)
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
			first_name = ?,
			last_name = ?,
			role = ?,
			title = ?,
			email = ?,
			image = ?,
			image_mime = ?,
			is_active = ?,
			is_verified = ?,
			last_signin_at = ?,
			version = version + 1,
			updated_at = ?
		WHERE id = ?
		  AND version = ?
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
		user.ID(),
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
		return fmt.Errorf("%w: optimistic lock failed for user %s", iam.ErrUserConcurrentModification, user.ID())
	}

	return nil

}

func (r *IAMRepository) Remove(ctx context.Context, user *iam.User) error {
	query := `
		DELETE FROM users
		WHERE id = ? AND version = ?
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
		return fmt.Errorf("%w: user id %s", iam.ErrUserConcurrentModification, user.ID())
	}

	return nil
}
