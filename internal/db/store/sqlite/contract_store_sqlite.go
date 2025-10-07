package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

// ContractStoreSqlite implements the ContractStore interface using SQLite.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ContractStoreSqlite struct {
	db *sql.DB
}

// NewContractStore constructs a ContractStoreSqlite that persists
// Contract entities in a SQLite database via the provided *sql.DB.
func NewContractStore(db *sql.DB) *ContractStoreSqlite {
	return &ContractStoreSqlite{db: db}
}

// CreateOne is an implementation of CreateOne in store.ContractStore.
// This method creates a new contract record, versions entry and
// system generated messages in relevant channels.
// It converts low-level driver-specific errors to domain specific
// errors defined in the store package.
//
//   - If a foreign key constraint violation occurs, it returns store.ErrForeignKeyViolation.
//   - If a unique constraint violation occurs, it returns store.ErrUniqueViolation.
//   - If a not-null constraint violation occurs, it returns store.ErrNotNullViolation.
//   - If any other errors occur, it returns store.ErrInsertFailed
func (q *ContractStoreSqlite) CreateOne(
	ctx context.Context,
	projectId int64,
	arg *params.ContractCreate) (*agg.ContractCreateResult, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertContractQuery := `
	INSERT INTO contracts(project_id, name, status) 
	VALUES(?, ?, (SELECT id FROM contract_statuses WHERE name = ?)) RETURNING id
	`

	var contractId int64
	if err := tx.QueryRowContext(
		ctx,
		insertContractQuery,
		projectId,
		arg.Name,
		params.ContractStatusPending,
	).Scan(&contractId); err != nil {

		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return nil, store.ErrNotNullViolation
			}

			return nil, store.ErrInsertFailed
		}
	}

	insertVersionQuery := `
	INSERT INTO contract_versions(contract_id, version, status, contract)
	VALUES(?, ?, (SELECT id FROM contract_version_statuses WHERE name = ?), ?)
	`

	if _, err := tx.ExecContext(
		ctx,
		insertVersionQuery,
		contractId,
		arg.Version,
		params.ContractStatusPending,
		arg.Contract); err != nil {

		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return nil, store.ErrNotNullViolation
			}

			return nil, store.ErrInsertFailed
		}
	}

	channelsQuery := `SELECT id FROM channels WHERE project_id = ?`

	rows, err := tx.QueryContext(ctx, channelsQuery, projectId)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer rows.Close()

	var channelIds []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, store.ErrInsertFailed
		}

		channelIds = append(channelIds, id)
	}

	result := agg.ContractCreateResult{
		ContractId: contractId,
		Version:    arg.Version,
		Messages:   make(map[int64]agg.MessageWithUser),
		UserIds:    make(map[int64][]int64),
	}

	userIdsQuery := `SELECT DISTINCT(user_id) FROM channel_users WHERE channel_id = ?`

	for _, channelId := range channelIds {
		usrRows, err := tx.QueryContext(ctx, userIdsQuery, channelId)
		if err != nil {
			return nil, store.ErrInsertFailed
		}

		var userIds []int64
		for usrRows.Next() {
			var id int64
			if err := usrRows.Scan(&id); err != nil {
				if err := usrRows.Close(); err != nil {
					return nil, store.ErrInsertFailed
				}
				return nil, store.ErrInsertFailed
			}

			userIds = append(userIds, id)
		}

		result.UserIds[channelId] = userIds

		if err := usrRows.Close(); err != nil {
			return nil, store.ErrInsertFailed
		}
	}

	insertMsgQuery := `
	INSERT INTO messages(user_id, channel_id, message, type) 
	VALUES(?, ?, ?, (SELECT id FROM message_types WHERE name = ?)) RETURNING id, created_at
	`

	for _, channelId := range channelIds {

		var msgId int64
		var createdAt time.Time

		err := tx.QueryRowContext(
			ctx,
			insertMsgQuery,
			nil, channelId,
			strconv.FormatInt(contractId, 10),
			params.MessageTypeContract,
		).Scan(&msgId, &createdAt)

		if err != nil {
			return nil, store.ErrInsertFailed
		}

		msg := agg.MessageWithUser{
			UserId:    0,
			MessageId: msgId,
			ChannelId: channelId,
			Message:   strconv.FormatInt(contractId, 10),
			CreatedAt: createdAt,
			FirstName: "",
			LastName:  "",
			Type:      params.MessageTypeContract,
			Role:      "",
			Title:     "",
			Image:     nil,
		}

		result.Messages[channelId] = msg
	}

	if err = tx.Commit(); err != nil {
		return nil, store.ErrInsertFailed
	}

	return &result, nil
}

// SignVersion adds a signature for the given contract version and user.
// It checks if the user is attempting to sign a
// contract/version that's already been signed/rejected.
// It returns a boolean, distinguishing between
// a newly created signature and existing signatures.
//
// If the contract or version was already signed, the method returns true.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) SignVersion(ctx context.Context, versionId int64, userId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrInsertFailed
	}

	defer tx.Rollback()

	// The signature could be added to the database
	// safely without trying to check status first.
	// But this method has to distinguish between a new
	// signature and an existing signature in the database.
	// So, the method checks if the specified contract version
	// is already signed or rejected.

	signed, err := q.isContractSignedOrRejected(ctx, tx, versionId)
	if err != nil {
		return false, err
	}

	if signed {
		return true, nil
	}

	usrSigned, err := q.hasUserSignedOrRejected(ctx, tx, versionId, userId)
	if err != nil {
		return false, err
	}

	if usrSigned {
		return true, nil
	}

	// If the user has not signed or rejected,
	// insert a new signature record.

	insertSignatureQuery := `
	INSERT OR IGNORE INTO contract_signatures(version_id, user_id, status)
	VALUES(?, ?, (SELECT id FROM contract_signature_statuses WHERE name = ?))
	`

	if _, err := tx.ExecContext(
		ctx,
		insertSignatureQuery,
		versionId,
		userId,
		params.ContractStatusSigned,
	); err != nil {

		return false, store.ErrInsertFailed
	}

	// Mark version as signed if the user is the last user to sign the contract.
	setVersionStatusQuery := `
	UPDATE contract_versions
    SET status = (SELECT id FROM contract_version_statuses WHERE name = ?)
    WHERE id = ?
    AND NOT EXISTS (
    	SELECT 1
        FROM project_users pu
        LEFT JOIN contract_signatures cs
        ON cs.user_id = pu.user_id
        AND cs.version_id = ?
        WHERE pu.project_id = (
        SELECT c.project_id FROM contracts c WHERE c.id = ?
    	)
    	AND cs.user_id IS NULL
    );
	`

	if _, err := tx.ExecContext(
		ctx,
		setVersionStatusQuery,
		params.ContractStatusSigned,
		versionId,
		versionId, // cs.version_id
		versionId, // c.id for project lookup
	); err != nil {

		return false, store.ErrInsertFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrInsertFailed
	}

	return false, nil
}

// isContractSignedOrRejected is helper method that checks
// if any contract version is signed or rejected.
// Considers a contract is signed/rejected if any version
// is signed/rejected.
//
//   - It returns true if the contract has been signed already.
//   - If an error occurs, store.ErrQueryFailed is returned.
func (q *ContractStoreSqlite) isContractSignedOrRejected(ctx context.Context, tx *sql.Tx, versionId int64) (bool, error) {

	query := `
	SELECT 
		CASE 
			WHEN cs.name IN (?, ?) 
			THEN 1 ELSE 0 
		END
	FROM contract_versions cv
	JOIN contracts c ON c.id = cv.contract_id
	JOIN contract_statuses cs ON cs.id = c.status
	WHERE cv.id = ?
	`

	var isSignedOrRejected bool
	if err := tx.QueryRowContext(
		ctx,
		query,
		params.ContractStatusSigned,
		params.ContractStatusRejected,
		versionId,
	).Scan(&isSignedOrRejected); err != nil {

		return false, store.ErrQueryFailed
	}
	return isSignedOrRejected, nil
}

// hasUserSignedOrRejected is a helper method that checks
// if a contract and version is signed or rejected by the
// specified user.
//
//   - It returns true if the contract has been signed already
//     by the specified user.
//   - If an error occurs, store.ErrQueryFailed is returned.
func (q *ContractStoreSqlite) hasUserSignedOrRejected(ctx context.Context, tx *sql.Tx, versionId int64, userId int64) (bool, error) {

	userSignedVersionQuery := `
	SELECT CASE
    		WHEN COUNT(*) > 0 THEN TRUE
    		ELSE FALSE
  		END AS has_signed
	FROM contract_signatures
	WHERE version_id = ? AND user_id = ?
	`

	var hasUserSigned bool
	if err := tx.QueryRowContext(
		ctx, userSignedVersionQuery,
		versionId,
		userId,
	).Scan(&hasUserSigned); err != nil {

		return false, store.ErrQueryFailed
	}

	if hasUserSigned {
		return true, nil
	}

	return false, nil
}

// GetUsersWithSignature retrieves the contract metadata and every project-user’s
// signature for the given contract version. It returns a pointer to
// an *aggregates.ContractUserSignatures struct with the metadata and signatures.
//
//   - If an error occurs, store.ErrQueryFailed is returned.
func (q *ContractStoreSqlite) GetUsersWithSignature(
	ctx context.Context,
	versionId int64) (*agg.ContractUserSignatures, error) {

	contractQuery := `
	SELECT c.name, cv.version, cv.contract FROM contract_versions cv
	JOIN contracts c ON c.id = cv.contract_id
	WHERE cv.id = ?
	`

	var contractUsrSigns agg.ContractUserSignatures
	if err := q.db.QueryRowContext(
		ctx,
		contractQuery,
		versionId,
	).Scan(
		&contractUsrSigns.ContractName,
		&contractUsrSigns.ContractVersion,
		&contractUsrSigns.ContractText); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrRecordNotFound
		}

		return nil, store.ErrQueryFailed
	}

	signaturesQuery := `
	SELECT 
	u.first_name, 
	u.last_name, 
	u.email,
	cs.created_at AS signed_date,
	COALESCE(c_stat.name, 'pending') as status 
	FROM contract_versions cv
	JOIN contracts c ON c.id = cv.contract_id
	JOIN project_users pu ON pu.project_id = c.project_id
	JOIN users u ON u.id = pu.user_id
	LEFT JOIN contract_signatures cs ON cs.user_id = u.id AND cs.version_id = cv.id
	LEFT JOIN contract_signature_statuses c_stat ON c_stat.id = cs.status
	WHERE cv.id = ?
	`

	rows, err := q.db.QueryContext(ctx, signaturesQuery, versionId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	contractUsrSigns.Signatures = make([]agg.ContractSignature, 0)

	for rows.Next() {

		var usrSign agg.ContractSignature
		if err := rows.Scan(
			&usrSign.FirstName,
			&usrSign.LastName,
			&usrSign.Email,
			&usrSign.SignedAt,
			&usrSign.Status,
		); err != nil {
			return nil, store.ErrQueryFailed
		}

		contractUsrSigns.Signatures = append(contractUsrSigns.Signatures, usrSign)
	}

	return &contractUsrSigns, nil
}

// RejectVersion adds a rejected signature for the given contract version and user.
// It checks if the user is attempting to reject a
// contract/version that's already been signed/rejected.
// It returns a boolean, distinguishing between
// a newly rejected signature and existing signatures.
//
// If the contract or version was already rejected or signed, the method returns true.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) RejectVersion(ctx context.Context, versionId int64, userId int64) (bool, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return false, store.ErrInsertFailed
	}

	defer tx.Rollback()

	signed, err := q.isContractSignedOrRejected(ctx, tx, versionId)
	if err != nil {
		return false, err
	}

	if signed {
		return true, nil
	}

	usrSigned, err := q.hasUserSignedOrRejected(ctx, tx, versionId, userId)
	if err != nil {
		return false, err
	}

	if usrSigned {
		return true, nil
	}

	// If the user has not signed or rejected,
	// insert a new signature record.

	insertSignatureQuery := `
	INSERT OR IGNORE INTO contract_signatures(version_id, user_id, status)
	VALUES(?, ?, (SELECT id FROM contract_signature_statuses WHERE name = ?))
	`

	if _, err := tx.ExecContext(
		ctx,
		insertSignatureQuery,
		versionId,
		userId,
		params.ContractStatusRejected,
	); err != nil {

		return false, store.ErrInsertFailed
	}

	setVersionStatusQuery := `
	UPDATE contract_versions
    SET status = (SELECT id FROM contract_version_statuses WHERE name = ?)
    WHERE id = ?
	`

	if _, err := tx.ExecContext(
		ctx,
		setVersionStatusQuery,
		params.ContractStatusRejected,
		versionId,
	); err != nil {

		return false, store.ErrInsertFailed
	}

	if err = tx.Commit(); err != nil {
		return false, store.ErrInsertFailed
	}

	return false, nil
}

// CreateRevision creates a contract revision request for a specified contract.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) CreateRevision(ctx context.Context, arg *params.ContractRevisionCreate) error {

	// TODO: set contract status to revision-requested or something similar.
	// Check if the contract is already signed or rejected.
	// If signed or rejected, deny revision.

	query := `
	INSERT INTO contract_revisions(contract_id, req_user_id, title, description, status)
	VALUES(?, ?, ?, ?, (SELECT id FROM contract_revision_statuses WHERE name = ?))
	`

	_, err := q.db.ExecContext(
		ctx,
		query,
		arg.ContractId,
		arg.UserId,
		arg.Title,
		arg.Description,
		params.ContractRevisionPending,
	)

	if err != nil {
		log.Println(err)
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return store.ErrNotNullViolation
			}
		}

		return store.ErrInsertFailed
	}

	return nil
}

// AcceptRevision sets contract revision status to the "accepted" state
// regardless of the previously set state.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) AcceptRevision(ctx context.Context, revisionId int64, userId int64) error {

	return q.setRevisionStatus(ctx, revisionId, userId, params.ContractRevisionAccepted)
}

// RejectRevision sets contract revision status to the "rejected" state
// regardless of the previously set state.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) RejectRevision(ctx context.Context, revisionId int64, userId int64) error {
	return q.setRevisionStatus(ctx, revisionId, userId, params.ContractRevisionRejected)
}

// setRevisionStatus is a private method that sets the status
// of a specified contract revision. This method sets the state
// regardless of the previous state of the contract revision.
//
// If any error occurs, store.ErrInsertFailed is returned.
func (q *ContractStoreSqlite) setRevisionStatus(
	ctx context.Context,
	revisionId int64,
	userId int64,
	status string) error {

	query := `
	UPDATE contract_revisions 
	SET res_user_id = ?, 
	updated_at = CURRENT_TIMESTAMP, 
	status = (SELECT id FROM contract_revision_statuses WHERE name = ?)
	WHERE id = ?
	`

	_, err := q.db.ExecContext(ctx, query, userId, status, revisionId)
	if err != nil {
		return store.ErrInsertFailed
	}

	return nil
}

// GetRevisions returns the total number of revisions found and
// a list of contract revision information with user data for
// a specified contract using the contract ID.
// The search criteria parameter can be used to filter results
// by a keyword, limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *ContractStoreSqlite) GetRevisions(
	ctx context.Context,
	contractId int64,
	arg *params.ContractRevisionSearch) (*agg.WithCount[agg.ContractRevisionWithUser], error) {

	var revisionQuery strings.Builder
	revisionQuery.WriteString(`
	SELECT 
		cr.id, 
		cr.contract_id, 
		cr.created_at,
		cr.updated_at, 
		cr.title, 
		cr.description,
		crs.name,
		req_u.first_name AS req_first_name,
		req_u.last_name AS req_last_name,
		res_u.first_name AS res_first_name,
		res_u.last_name AS res_last_name
	FROM contract_revisions cr
	JOIN contract_revision_statuses crs ON crs.id = cr.status
	LEFT JOIN users req_u ON req_u.id = cr.req_user_id
	LEFT JOIN users res_u ON res_u.id = cr.res_user_id 
	WHERE cr.contract_id = ?
	`)

	var revisionCountQuery strings.Builder
	revisionCountQuery.WriteString(`
	SELECT COUNT(cr.id) FROM contract_revisions cr
	JOIN contract_revision_statuses crs ON crs.id = cr.status
	WHERE cr.contract_id = ?
	`)

	revisionQueryArgs := []any{contractId}
	revisionCountArgs := []any{contractId}

	if len(arg.Keyword) != 0 {
		revisionQuery.WriteString(" AND ( cr.title LIKE ? OR cr.description LIKE ? )")
		revisionCountQuery.WriteString(" AND ( cr.title LIKE ? OR cr.description LIKE ? )")

		revisionQueryArgs = append(revisionQueryArgs, "%"+arg.Keyword+"%", "%"+arg.Keyword+"%")
		revisionCountArgs = append(revisionCountArgs, "%"+arg.Keyword+"%", "%"+arg.Keyword+"%")
	}

	if len(arg.Status) != 0 {
		revisionQuery.WriteString(" AND crs.name = ?")
		revisionCountQuery.WriteString(" AND crs.name = ?")

		revisionQueryArgs = append(revisionQueryArgs, arg.Status)
		revisionCountArgs = append(revisionCountArgs, arg.Status)
	}

	revisionQuery.WriteString(" ORDER BY cr.created_at DESC")
	revisionQuery.WriteString(" LIMIT ? OFFSET ?")
	revisionQueryArgs = append(revisionQueryArgs, arg.Limit, arg.Offset)

	var totalRevisions int64
	err := q.db.QueryRowContext(ctx, revisionCountQuery.String(), revisionCountArgs...).Scan(&totalRevisions)

	if err != nil {
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, revisionQuery.String(), revisionQueryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.ContractRevisionWithUser]{
		Total: totalRevisions,
		Items: make([]agg.ContractRevisionWithUser, 0),
	}

	for rows.Next() {
		var row agg.ContractRevisionWithUser
		err := rows.Scan(
			&row.Id,
			&row.ContractId,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.Title,
			&row.Description,
			&row.Status,
			&row.ReqUserFirstName,
			&row.ReqUserLastName,
			&row.ResUserFirstName,
			&row.ResUserLastName,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}

// GetContractStatsByProject returns the total number of contracts found and
// a list of contract with metrics such as number of versions and revisions
// for a specified project.
// The search criteria parameter can be used to filter results
// by a keyword, limit and offset results.
//
// If any error occurs, store.ErrQueryFailed is returned.
func (q *ContractStoreSqlite) GetContractStatsByProject(
	ctx context.Context,
	projectId int64,
	arg *params.ContractSearch) (*agg.WithCount[agg.ContractWithStats], error) {

	var query strings.Builder
	var countQuery strings.Builder
	query.WriteString(`
	WITH version_counts AS (
  		SELECT
    		contract_id,
    		COUNT(*) AS versions
  		FROM contract_versions
  		GROUP BY contract_id
	),
	revision_counts AS (
  		SELECT
    		contract_id,
    		COUNT(*) AS revisions
  		FROM contract_revisions
  		GROUP BY contract_id
	),
	accepted_revisions AS (
		SELECT 
			cr.contract_id AS contract_id,
			COUNT(*) AS accepted
		FROM contract_revisions cr
		JOIN contract_revision_statuses crs ON crs.id = cr.status
		WHERE crs.name = ?
		GROUP BY cr.contract_id
	)

	SELECT
  		c.id,
  		c.name,
  		cs.name AS status,
  		c.created_at,
  		COALESCE(v.versions,  0) AS versions,
  		COALESCE(r.revisions, 0) AS revisions,
  		COALESCE(ar.accepted, 0) AS accepted_revisions
	FROM contracts c
	JOIN contract_statuses cs ON cs.id = c.status
	LEFT JOIN version_counts v ON v.contract_id = c.id
	LEFT JOIN revision_counts r ON r.contract_id = c.id
	LEFT JOIN accepted_revisions ar ON ar.contract_id = c.id
	WHERE c.project_id = ?
	`)

	countQuery.WriteString(`
	SELECT COUNT(c.id) FROM contracts c
	JOIN contract_statuses cs ON cs.id = c.status
	WHERE c.project_id = ?
	`)

	queryArgs := []any{params.ContractRevisionAccepted, projectId}
	countQueryArgs := []any{projectId}

	if len(arg.Keyword) > 0 {
		query.WriteString(" AND c.name LIKE ?")
		countQuery.WriteString(" AND c.name LIKE ?")

		queryArgs = append(queryArgs, "%"+arg.Keyword+"%")
		countQueryArgs = append(countQueryArgs, "%"+arg.Keyword+"%")
	}

	if len(arg.Status) > 0 {
		query.WriteString(" AND cs.name = ?")
		countQuery.WriteString(" AND cs.name = ?")

		queryArgs = append(queryArgs, arg.Status)
		countQueryArgs = append(countQueryArgs, arg.Status)
	}

	query.WriteString(" LIMIT ? OFFSET ?")
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	var totalContracts int64
	err := q.db.QueryRowContext(ctx, countQuery.String(), countQueryArgs...).Scan(&totalContracts)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, query.String(), queryArgs...)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.ContractWithStats]{
		Total: totalContracts,
		Items: make([]agg.ContractWithStats, 0),
	}

	for rows.Next() {

		var row agg.ContractWithStats
		if err := rows.Scan(
			&row.Id,
			&row.Name,
			&row.Status,
			&row.CreatedAt,
			&row.Versions,
			&row.Revisions,
			&row.AcceptedRevisions,
		); err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}

func (q *ContractStoreSqlite) GetVersionsByContractId(
	ctx context.Context,
	contractId int64) ([]agg.ContractVersion, error) {

	query := `
	SELECT 
		cv.id,
		cv.created_at,
		cv.version,
		cv.contract,
		cs.name AS status
	FROM contract_versions cv
	JOIN contract_statuses cs ON cs.id = cv.status
	WHERe cv.contract_id = ?
	ORDER BY cv.created_at DESC
	`

	rows, err := q.db.QueryContext(ctx, query, contractId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := make([]agg.ContractVersion, 0)
	for rows.Next() {
		var row agg.ContractVersion
		err := rows.Scan(
			&row.Id,
			&row.CreatedAt,
			&row.Version,
			&row.Contract,
			&row.Status,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result = append(result, row)
	}

	return result, nil
}
