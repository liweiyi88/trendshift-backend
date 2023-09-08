package cmd

import (
	"embed"
	"errors"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/liweiyi88/gti/config"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
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
			log.Fatal("failed to create new iofs", err)
		}

		m, err := migrate.NewWithSourceInstance("iofs", d, "mysql://"+config.DatabaseDSN)
		if err != nil {
			log.Fatal("failed to create new migration instance", err)
		}

		defer func() {
			sourceErr, databaseErr := m.Close()

			if sourceErr != nil {
				slog.Error("failed to close resource, source error: %v", sourceErr)
			}

			if databaseErr != nil {
				slog.Error("failed to close resource, database error error: %v", databaseErr)
			}
		}()

		if action == "up" {
			err = m.Up()
		} else if action == "down" {
			err = m.Down()
		} else {
			err = errors.New("unsupported action, only 'up' or 'down' are valida action")
		}

		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal("migration failed: ", err)
		}
	},
}
