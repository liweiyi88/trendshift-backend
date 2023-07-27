package database

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/liweiyi88/gti/internal/config"
	"golang.org/x/exp/slog"
)

var db *sql.DB
var once sync.Once

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Database struct {
	client *sql.DB
}

func (d *Database) ExecContext(ctx context.Context, query string, values ...any) (sql.Result, error) {
	statement, err := d.client.PrepareContext(ctx, query)

	if err != nil {
		return nil, err
	}

	result, err := statement.ExecContext(ctx, values)
	statement.Close()

	return result, err
}

func (d *Database) QueryRowContext(ctx context.Context, query string, values ...any) *sql.Row {
	return d.client.QueryRowContext(ctx, query, values...)
}

func (d *Database) QueryContext(ctx context.Context, query string, values ...any) (*sql.Rows, error) {
	rows, err := d.client.QueryContext(ctx, query, values...)
	return rows, err
}

func GetInstance() *sql.DB {
	once.Do(func() {
		var err error
		db, err = sql.Open("mysql", config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}

		pingErr := db.Ping()

		if pingErr != nil {
			log.Fatal(pingErr)
		}

		slog.Info("database connected.")

		// Avoid closing bad idle connection: unexpected read from socket, driver: bad connection error
		// Reference: https://github.com/go-sql-driver/mysql/issues/1120#issuecomment-636795680
		db.SetConnMaxLifetime(3 * time.Minute)
	})

	return db
}
