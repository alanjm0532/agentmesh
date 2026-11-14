package infra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"go.uber.org/zap"
)

func NewDatabase(cfg config.Config, logger *zap.Logger) (*ent.Client, func(), error) {
	var dialectName string
	var connector driver.Connector
	var err error
	switch cfg.Database.Driver {
	case "mysql":
		dialectName = dialect.MySQL
		connector, err = newMySQLConnector(cfg.Database.DSN)
	case "postgres":
		dialectName = dialect.Postgres
		connector, err = newPostgresConnector(cfg.Database.DSN)
	case "sqlite":
		dialectName = dialect.SQLite
		connector, err = newSQLiteConnector(cfg.Database.DSN)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q", cfg.Database.Driver)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("configure database connection: %w", err)
	}
	// The connector applies UTC settings to every new connection in the pool.
	db := sql.OpenDB(connector)
	cleanup := func() { _ = db.Close() }
	initialized := false
	defer func() {
		// Close the pool if initialization fails or panics before ownership transfers.
		if !initialized {
			cleanup()
		}
	}()
	if cfg.Database.Driver == "sqlite" {
		// Keep in-memory databases and SQLite writes on the same connection.
		db.SetMaxOpenConns(1)
	}
	// Each wrapper takes ownership of the same pool, so keep only its current closer.
	sqlDriver := entsql.OpenDB(dialectName, db)
	cleanup = func() { _ = sqlDriver.Close() }
	opts := []ent.Option{ent.Driver(sqlDriver)}
	if cfg.Server.Mode == "debug" {
		opts = append(opts, ent.Debug(), ent.Log(logger.Named("sql").Sugar().Info))
	}
	client := ent.NewClient(opts...)
	cleanup = func() { _ = client.Close() }
	if err = client.Schema.Create(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("migrate database: %w", err)
	}
	// Wire owns the client after successful initialization and invokes its cleanup.
	initialized = true
	return client, cleanup, nil
}
