package database

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/liweiyi88/trendshift-backend/config"
)

var db *sql.DB
var once sync.Once

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func GetInstance(ctx context.Context) *sql.DB {
	once.Do(func() {
		var err error
		db, err = sql.Open("mysql", config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}

		// Avoid closing bad idle connection: unexpected read from socket, driver: bad connection error
		// Reference: https://github.com/go-sql-driver/mysql/issues/1120#issuecomment-636795680
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetMaxIdleConns(5)

		// The current managed database only support 75 concurrent connections
		// See https://docs.digitalocean.com/products/databases/mysql/details/limits/
		db.SetMaxOpenConns(25)

		ping(ctx)
		slog.Info("database connected.")
	})

	return db
}

func ping(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
}
