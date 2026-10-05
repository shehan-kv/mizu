package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("uow.Execute: %w", err)
	}

	txCtx := withTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("uow.Execute rollback: %w, original: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("uow.Execute commit: %w", err)
	}

	return nil
}
