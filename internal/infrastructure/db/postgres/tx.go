package postgres

import (
	"context"
	"database/sql"
)

type txKey struct{}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
