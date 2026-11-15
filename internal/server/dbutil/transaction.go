package dbutil

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"go.uber.org/zap"
)

// Transaction runs action once, committing on success and rolling back on error
// or panic. Database failures are returned to the caller without retrying.
func Transaction(ctx context.Context, client *ent.Client, logger *zap.Logger, action func(*ent.Tx) error) error {
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
func Rollback(tx *ent.Tx, logger *zap.Logger) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		logger.Error("failed to roll back database transaction", zap.Error(err))
	}
}
