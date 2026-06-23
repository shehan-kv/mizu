package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ContractRepository struct {
	db *sql.DB
}

func NewContractRepository(db *sql.DB) *ContractRepository {
	return &ContractRepository{
		db: db,
	}
}

func (r *ContractRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *ContractRepository) Add(ctx context.Context, c *contract.Contract) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO contracts(
            id,
            project_id,
            name,
            status,
            terms,
            version,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID().String(),
		c.ProjectID().String(),
		c.Name().String(),
		c.Status().String(),
		c.Terms().String(),
		c.Version(),
		c.CreatedAt(),
		c.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("contract.ContractRepository.Add: %w", err)
	}

	signatories := c.Signatories()

	if len(signatories) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(signatories)*4)

	sb.WriteString(`
        INSERT INTO contract_signatories(
            contract_id,
            user_id,
            status,
            updated_at
        ) VALUES 
    `)

	for i := range signatories {
		if i > 0 {
			sb.WriteString(", ")
		}

		start := len(args)

		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf("$%d, $%d, $%d, $%d",
			start+1, start+2, start+3, start+4))
		sb.WriteString(")")

		args = append(
			args,
			c.ID().String(),
			signatories[i].UserID().String(),
			signatories[i].Status().String(),
			signatories[i].UpdatedAt(),
		)
	}

	_, err = ex.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("contract.ContractRepository.Add: %w", err)
	}

	return nil
}

func (r *ContractRepository) Get(ctx context.Context, id contract.ContractID) (*contract.Contract, error) {
	ex := r.executor(ctx)

	var (
		rawID        string
		rawProjectID string
		name         string
		status       string
		terms        string
		version      int
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
            id,
            project_id,
            name,
            status,
            terms,
            version,
            created_at,
            updated_at
        FROM contracts
        WHERE id = $1`,
		id.String(),
	).Scan(
		&rawID,
		&rawProjectID,
		&name,
		&status,
		&terms,
		&version,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"%w: %w",
				contract.ErrContractNotFound,
				err,
			)
		}

		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	rows, err := ex.QueryContext(
		ctx,
		`SELECT
            user_id,
            status,
            updated_at
        FROM contract_signatories
        WHERE contract_id = $1`,
		rawID,
	)

	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}
	defer rows.Close()

	signatories := make([]contract.Signatory, 0)

	for rows.Next() {
		var (
			rawUserID   string
			rawStatus   string
			signatoryAt time.Time
		)

		if err := rows.Scan(
			&rawUserID,
			&rawStatus,
			&signatoryAt,
		); err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
		}

		userID, err := iam.NewUserID(rawUserID)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
		}

		statusVO, err := contract.NewSignatoryStatus(rawStatus)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
		}

		signatory := contract.RestoreSignatory(userID, statusVO, signatoryAt)

		signatories = append(signatories, signatory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	contractID, err := contract.NewContractID(rawID)
	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	projectID, err := project.NewProjectID(rawProjectID)
	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	nameVO, err := contract.NewName(name)
	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	statusVO, err := contract.NewStatus(status)
	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	termsVO, err := contract.NewTerms(terms)
	if err != nil {
		return nil, fmt.Errorf("contract.ContractRepository.Get: %w", err)
	}

	return contract.RestoreContract(
		contractID,
		projectID,
		nameVO,
		statusVO,
		termsVO,
		signatories,
		version,
		createdAt,
		updatedAt,
	), nil
}

func (r *ContractRepository) GetStatsByProject(ctx context.Context, pID project.ProjectID) (contract.Stats, error) {
	ex := r.executor(ctx)

	var (
		total  int
		signed int
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = $1) AS signed
		FROM contracts
		WHERE project_id = $2`,
		contract.StatusSigned.String(),
		pID.String(),
	).Scan(
		&total,
		&signed,
	)

	if err != nil {
		return contract.Stats{}, fmt.Errorf(
			"contract.ContractRepository.GetStatsByProject: %w",
			err,
		)
	}

	return contract.NewStats(total, signed), nil
}

func (r *ContractRepository) ListByProject(ctx context.Context, f contract.FilterByProject, p common.Page) ([]*contract.Contract, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 8)

	argPos := 1

	query.WriteString(`
		SELECT
			id,
			project_id,
			name,
			status,
			terms,
			version,
			created_at,
			updated_at
		FROM contracts
		WHERE project_id = $1
	`)

	args = append(args, f.ProjectID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(fmt.Sprintf(` AND name LIKE $%d`, argPos))
		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		query.WriteString(fmt.Sprintf(` AND status = $%d`, argPos))
		args = append(args, f.Status.String())
		argPos++
	}

	query.WriteString(fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, argPos, argPos+1))

	args = append(args, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListByProject: %w",
			err,
		)
	}
	defer rows.Close()

	type rawContract struct {
		id        string
		projectID string
		name      string
		status    string
		terms     string
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	rawContracts := make([]rawContract, 0)
	contractIDs := make([]any, 0)

	for rows.Next() {
		var rc rawContract

		if err := rows.Scan(
			&rc.id,
			&rc.projectID,
			&rc.name,
			&rc.status,
			&rc.terms,
			&rc.version,
			&rc.createdAt,
			&rc.updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListByProject: %w",
				err,
			)
		}

		rawContracts = append(rawContracts, rc)
		contractIDs = append(contractIDs, rc.id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListByProject: %w",
			err,
		)
	}

	if len(rawContracts) == 0 {
		return []*contract.Contract{}, nil
	}

	var signatoryQuery strings.Builder
	signatoryArgs := make([]any, 0, 1)

	signatoryQuery.WriteString(`
		SELECT
			contract_id,
			user_id,
			status,
			updated_at
		FROM contract_signatories
		WHERE contract_id = ANY($1)
	`)

	signatoryArgs = append(signatoryArgs, pq.Array(contractIDs))

	signatoryRows, err := ex.QueryContext(
		ctx,
		signatoryQuery.String(),
		signatoryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListByProject: %w",
			err,
		)
	}
	defer signatoryRows.Close()

	signatoriesByContract := make(map[string][]contract.Signatory, len(rawContracts))

	for signatoryRows.Next() {
		var (
			rawContractID string
			rawUserID     string
			rawStatus     string
			updatedAt     time.Time
		)

		if err := signatoryRows.Scan(
			&rawContractID,
			&rawUserID,
			&rawStatus,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListByProject: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(rawUserID)
		if err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListByProject: %w",
				err,
			)
		}

		statusVO, err := contract.NewSignatoryStatus(rawStatus)
		if err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListByProject: %w",
				err,
			)
		}

		signatory := contract.RestoreSignatory(
			userID,
			statusVO,
			updatedAt,
		)

		signatoriesByContract[rawContractID] = append(
			signatoriesByContract[rawContractID],
			signatory,
		)
	}

	if err := signatoryRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListByProject: %w",
			err,
		)
	}

	contracts := make([]*contract.Contract, 0, len(rawContracts))

	for i := range rawContracts {
		contractID, err := contract.NewContractID(rawContracts[i].id)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListByProject: %w", err)
		}

		projectID, err := project.NewProjectID(rawContracts[i].projectID)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListByProject: %w", err)
		}

		nameVO, err := contract.NewName(rawContracts[i].name)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListByProject: %w", err)
		}

		statusVO, err := contract.NewStatus(rawContracts[i].status)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListByProject: %w", err)
		}

		termsVO, err := contract.NewTerms(rawContracts[i].terms)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListByProject: %w", err)
		}

		c := contract.RestoreContract(
			contractID,
			projectID,
			nameVO,
			statusVO,
			termsVO,
			signatoriesByContract[rawContracts[i].id],
			rawContracts[i].version,
			rawContracts[i].createdAt,
			rawContracts[i].updatedAt,
		)

		contracts = append(contracts, c)
	}

	return contracts, nil
}

func (r *ContractRepository) ListBySignatory(ctx context.Context, f contract.FilterBySignatory, p common.Page) ([]*contract.Contract, error) {
	ex := r.executor(ctx)

	var query strings.Builder

	args := make([]any, 0, 8)
	argPos := 1

	query.WriteString(`
		SELECT DISTINCT
			c.id,
			c.project_id,
			c.name,
			c.status,
			c.terms,
			c.version,
			c.created_at,
			c.updated_at
		FROM contracts c
		INNER JOIN contract_signatories cs
			ON cs.contract_id = c.id
		WHERE cs.user_id = $1
	`)

	args = append(args, f.SignatoryID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(fmt.Sprintf(` AND c.name LIKE $%d`, argPos))
		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		query.WriteString(fmt.Sprintf(` AND c.status = $%d`, argPos))
		args = append(args, f.Status.String())
		argPos++
	}

	query.WriteString(fmt.Sprintf(`
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, argPos, argPos+1))

	args = append(args, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListBySignatory: %w",
			err,
		)
	}
	defer rows.Close()

	type rawContract struct {
		id        string
		projectID string
		name      string
		status    string
		terms     string
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	rawContracts := make([]rawContract, 0)
	contractIDs := make([]any, 0)

	for rows.Next() {
		var rc rawContract

		if err := rows.Scan(
			&rc.id,
			&rc.projectID,
			&rc.name,
			&rc.status,
			&rc.terms,
			&rc.version,
			&rc.createdAt,
			&rc.updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListBySignatory: %w",
				err,
			)
		}

		rawContracts = append(rawContracts, rc)
		contractIDs = append(contractIDs, rc.id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListBySignatory: %w",
			err,
		)
	}

	if len(rawContracts) == 0 {
		return []*contract.Contract{}, nil
	}

	var signatoryQuery strings.Builder

	signatoryQuery.WriteString(`
		SELECT
			contract_id,
			user_id,
			status,
			updated_at
		FROM contract_signatories
		WHERE contract_id = ANY($1)
	`)

	signatoryRows, err := ex.QueryContext(
		ctx,
		signatoryQuery.String(),
		pq.Array(contractIDs),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListBySignatory: %w",
			err,
		)
	}
	defer signatoryRows.Close()

	signatoriesByContract := make(map[string][]contract.Signatory, len(rawContracts))

	for signatoryRows.Next() {
		var (
			rawContractID string
			rawUserID     string
			rawStatus     string
			updatedAt     time.Time
		)

		if err := signatoryRows.Scan(
			&rawContractID,
			&rawUserID,
			&rawStatus,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListBySignatory: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(rawUserID)
		if err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListBySignatory: %w",
				err,
			)
		}

		statusVO, err := contract.NewSignatoryStatus(rawStatus)
		if err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListBySignatory: %w",
				err,
			)
		}

		signatory := contract.RestoreSignatory(
			userID,
			statusVO,
			updatedAt,
		)

		signatoriesByContract[rawContractID] = append(
			signatoriesByContract[rawContractID],
			signatory,
		)
	}

	if err := signatoryRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListBySignatory: %w",
			err,
		)
	}

	contracts := make([]*contract.Contract, 0, len(rawContracts))

	for i := range rawContracts {
		contractID, err := contract.NewContractID(rawContracts[i].id)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListBySignatory: %w", err)
		}

		projectID, err := project.NewProjectID(rawContracts[i].projectID)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListBySignatory: %w", err)
		}

		nameVO, err := contract.NewName(rawContracts[i].name)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListBySignatory: %w", err)
		}

		statusVO, err := contract.NewStatus(rawContracts[i].status)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListBySignatory: %w", err)
		}

		termsVO, err := contract.NewTerms(rawContracts[i].terms)
		if err != nil {
			return nil, fmt.Errorf("contract.ContractRepository.ListBySignatory: %w", err)
		}

		c := contract.RestoreContract(
			contractID,
			projectID,
			nameVO,
			statusVO,
			termsVO,
			signatoriesByContract[rawContracts[i].id],
			rawContracts[i].version,
			rawContracts[i].createdAt,
			rawContracts[i].updatedAt,
		)

		contracts = append(contracts, c)
	}

	return contracts, nil
}

func (r *ContractRepository) ListStatsByProjects(ctx context.Context, projectIDs []project.ProjectID) (map[project.ProjectID]contract.Stats, error) {
	ex := r.executor(ctx)

	statsByProjects := make(map[project.ProjectID]contract.Stats)

	if len(projectIDs) == 0 {
		return statsByProjects, nil
	}

	var query strings.Builder
	args := make([]any, 0, 2)

	query.WriteString(`
		SELECT
			project_id,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = $1) AS signed
		FROM contracts
		WHERE project_id = ANY($2)
		GROUP BY project_id
	`)

	args = append(args, contract.StatusSigned.String())
	args = append(args, pq.Array(projectIDs))

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListStatsByProjects: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rawProjectID string
			total        int
			signed       int
		)

		if err := rows.Scan(
			&rawProjectID,
			&total,
			&signed,
		); err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListStatsByProjects: %w",
				err,
			)
		}

		projectID, err := project.NewProjectID(rawProjectID)
		if err != nil {
			return nil, fmt.Errorf(
				"contract.ContractRepository.ListStatsByProjects: %w",
				err,
			)
		}

		statsByProjects[projectID] = contract.NewStats(total, signed)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"contract.ContractRepository.ListStatsByProjects: %w",
			err,
		)
	}

	return statsByProjects, nil
}

func (r *ContractRepository) CountByProject(ctx context.Context, f contract.FilterByProject) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 4)

	argPos := 1

	query.WriteString(`
		SELECT COUNT(*)
		FROM contracts
		WHERE project_id = $1
	`)

	args = append(args, f.ProjectID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(`
			AND name LIKE $
		`)
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		query.WriteString(`
			AND status = $
		`)
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	var count int

	if err := ex.QueryRowContext(
		ctx,
		query.String(),
		args...,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf(
			"contract.ContractRepository.CountByProject: %w",
			err,
		)
	}

	return count, nil
}

func (r *ContractRepository) CountBySignatory(ctx context.Context, f contract.FilterBySignatory) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 4)

	argPos := 1

	query.WriteString(`
		SELECT COUNT(DISTINCT c.id)
		FROM contracts c
		INNER JOIN contract_signatories cs
			ON cs.contract_id = c.id
		WHERE cs.user_id = $1
	`)

	args = append(args, f.SignatoryID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(`
			AND c.name LIKE $
		`)
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, "%"+*f.Keyword+"%")
		argPos++
	}

	if f.Status != nil {
		query.WriteString(`
			AND c.status = $
		`)
		query.WriteString(strconv.Itoa(argPos))

		args = append(args, f.Status.String())
		argPos++
	}

	var count int

	if err := ex.QueryRowContext(
		ctx,
		query.String(),
		args...,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf(
			"contract.ContractRepository.CountBySignatory: %w",
			err,
		)
	}

	return count, nil
}

func (r *ContractRepository) Save(ctx context.Context, c *contract.Contract) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`UPDATE contracts
		SET
			name = $1,
			status = $2,
			terms = $3,
			version = version + 1,
			updated_at = $4
		WHERE id = $5 AND version = $6`,
		c.Name().String(),
		c.Status().String(),
		c.Terms().String(),
		c.UpdatedAt(),
		c.ID().String(),
		c.Version(),
	)

	if err != nil {
		return fmt.Errorf(
			"contract.ContractRepository.Save: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"contract.ContractRepository.Save: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"%w: version mismatch",
			contract.ErrContractConcurrentModification,
		)
	}

	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM contract_signatories
		WHERE contract_id = $1`,
		c.ID().String(),
	)

	if err != nil {
		return fmt.Errorf(
			"contract.ContractRepository.Save: %w",
			err,
		)
	}

	signatories := c.Signatories()

	if len(signatories) == 0 {
		return nil
	}

	var sb strings.Builder
	args := make([]any, 0, len(signatories)*4)

	argPos := 1

	sb.WriteString(`
		INSERT INTO contract_signatories(
			contract_id,
			user_id,
			status,
			updated_at
		) VALUES 
	`)

	for i := range signatories {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("(")
		sb.WriteString(strconv.Itoa(argPos))
		sb.WriteString(", ")
		sb.WriteString(strconv.Itoa(argPos + 1))
		sb.WriteString(", ")
		sb.WriteString(strconv.Itoa(argPos + 2))
		sb.WriteString(", ")
		sb.WriteString(strconv.Itoa(argPos + 3))
		sb.WriteString(")")

		argPos += 4

		args = append(
			args,
			c.ID().String(),
			signatories[i].UserID().String(),
			signatories[i].Status().String(),
			signatories[i].UpdatedAt(),
		)
	}

	_, err = ex.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf(
			"contract.ContractRepository.Save: %w",
			err,
		)
	}

	return nil
}
