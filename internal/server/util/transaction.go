package util

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

type transaction interface {
	Commit() error
	Rollback() error
}

// Transaction runs action once, committing on success and rolling back on error
// or panic. Database failures are returned to the caller without retrying.
func Transaction[T transaction](ctx context.Context, client interface {
	Tx(context.Context) (T, error)
}, logger *zap.Logger, action func(T) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer Rollback(tx, logger)
	if err = action(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Rollback releases the transaction and logs unexpected failures without hiding
// the original action error or panic. A committed transaction is already closed.
func Rollback(tx interface{ Rollback() error }, logger *zap.Logger) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error("failed to roll back database transaction", zap.Error(err))
	}
}
