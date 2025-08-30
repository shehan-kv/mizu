package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strconv"
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
	arg *params.ContractCreate) (*aggregates.ContractCreateResult, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertContractQuery := `INSERT INTO contracts(project_id, name) VALUES(?, ?) RETURNING id`

	var contractId int64
	if err := tx.QueryRowContext(ctx, insertContractQuery, projectId, arg.Name).Scan(&contractId); err != nil {
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
	VALUES(?, ?, (SELECT id FROM contract_statuses WHERE name = ?), ?)
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

	result := aggregates.ContractCreateResult{
		ContractId: contractId,
		Version:    arg.Version,
		Messages:   make(map[int64]aggregates.MessageWithUser),
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

		msg := aggregates.MessageWithUser{
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
	VALUES(?, ?, (SELECT id FROM contract_statuses WHERE name = ?))
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
    SET status = (SELECT id FROM contract_statuses WHERE name = ?)
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

	isContractSignedQuery := `
	SELECT
  		CASE
    		WHEN COUNT(*) > 0 THEN TRUE
    		ELSE FALSE
  		END AS is_signed
	FROM contract_versions cv
	JOIN contract_versions cv2
  	ON cv2.contract_id = cv.contract_id
	JOIN contract_statuses cs
  	ON cv2.status = cs.id
	WHERE cv.id = ?
  	AND cs.name = ? OR cs.name = ?
	`

	var isContractSigned bool
	if err := tx.QueryRowContext(
		ctx,
		isContractSignedQuery,
		versionId,
		params.ContractStatusSigned,
		params.ContractStatusRejected,
	).Scan(&isContractSigned); err != nil {

		return false, store.ErrQueryFailed
	}

	if isContractSigned {
		return true, nil
	}

	return false, nil
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
	versionId int64) (*aggregates.ContractUserSignatures, error) {

	contractQuery := `
	SELECT c.name, cv.version, cv.contract FROM contract_versions cv
	JOIN contracts c ON c.id = cv.contract_id
	WHERE cv.id = ?
	`

	var contractUsrSigns aggregates.ContractUserSignatures
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
	LEFT JOIN contract_statuses c_stat ON c_stat.id = cs.status
	WHERE cv.id = ?
	`

	rows, err := q.db.QueryContext(ctx, signaturesQuery, versionId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	contractUsrSigns.Signatures = make([]aggregates.ContractSignature, 0)

	for rows.Next() {

		var usrSign aggregates.ContractSignature
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
	VALUES(?, ?, (SELECT id FROM contract_statuses WHERE name = ?))
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
    SET status = (SELECT id FROM contract_statuses WHERE name = ?)
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
