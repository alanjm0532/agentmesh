package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
)

func NewDatabase(cfg config.Config) (*ent.Client, func(), error) {
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
	if cfg.Database.Driver == "sqlite" {
		// Keep in-memory databases and SQLite writes on the same connection.
		db.SetMaxOpenConns(1)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialectName, db)))
	cleanup := func() { _ = client.Close() }
	if err := client.Schema.Create(context.Background()); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("migrate database: %w", err)
	}
	return client, cleanup, nil
}
