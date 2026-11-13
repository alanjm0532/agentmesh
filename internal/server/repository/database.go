package repository

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func NewDatabase(cfg config.Config) (*ent.Client, func(), error) {
	var dialectName string
	switch cfg.Database.Driver {
	case "mysql":
		dialectName = dialect.MySQL
	case "postgres":
		dialectName = dialect.Postgres
	case "sqlite":
		dialectName = dialect.SQLite
	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q", cfg.Database.Driver)
	}

	// modernc registers as sqlite; Ent uses the sqlite3 dialect for its SQL.
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
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
