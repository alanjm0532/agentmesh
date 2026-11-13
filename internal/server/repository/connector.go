package repository

import (
	"database/sql/driver"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"modernc.org/sqlite"
)

func newMySQLConnector(dsn string) (driver.Connector, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	// Driver conversion and the server session must agree on the time zone.
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	params := map[string]string{"time_zone": "'+00:00'"}
	for key, value := range cfg.Params {
		if !strings.EqualFold(key, "time_zone") {
			params[key] = value
		}
	}
	cfg.Params = params
	return mysql.NewConnector(cfg)
}

func newPostgresConnector(dsn string) (driver.Connector, error) {
	cfg, err := pq.NewConfig(dsn)
	if err != nil {
		return nil, err
	}
	// Startup parameters apply to every connection and support URL and keyword DSNs.
	params := map[string]string{"timezone": "UTC"}
	for key, value := range cfg.Runtime {
		if !strings.EqualFold(key, "timezone") {
			params[key] = value
		}
	}
	cfg.Runtime = params
	return pq.NewConnectorConfig(cfg)
}

func newSQLiteConnector(dsn string) (driver.Connector, error) {
	name, query, _ := strings.Cut(dsn, "?")
	params, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	// Preserve repeated pragmas while normalizing timestamp encoding and decoding.
	params.Set("_timezone", "UTC")
	return sqlite.NewConnector(name + "?" + params.Encode())
}
