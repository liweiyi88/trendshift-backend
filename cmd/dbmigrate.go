package cmd

import (
	"embed"
	"errors"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var fs embed.FS

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Args:  cobra.ExactArgs(1),
	Short: "Run database migration, valid args is either up or down",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		action := args[0]

		d, err := iofs.New(fs, "migrations")
		if err != nil {
			slog.Error("failed to create new iofs", slog.Any("error", err))
			sentry.CaptureException(err)
			return
		}

		m, err := migrate.NewWithSourceInstance("iofs", d, "mysql://"+config.DatabaseDSN)
		if err != nil {
			slog.Error("failed to create new migration instance", slog.Any("error", err))
			sentry.CaptureException(err)
			return
		}

		defer func() {
			sourceErr, databaseErr := m.Close()

			if sourceErr != nil {
				slog.Error("failed to close source resource", slog.Any("source error", sourceErr))
				sentry.CaptureException(sourceErr)
			}

			if databaseErr != nil {
				slog.Error("failed to close database resource", slog.Any("database error", databaseErr))
				sentry.CaptureException(databaseErr)
			}

			sentry.Flush(2 * time.Second)
		}()

		switch action {
		case "up":
			err = m.Up()
		case "down":
			err = m.Steps(-1)
		default:
			err = errors.New("unsupported action, only 'up' or 'down' are valid action")
		}

		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			slog.Error("migration failed", slog.Any("error", err))
			sentry.CaptureException(err)
		}
	},
}
