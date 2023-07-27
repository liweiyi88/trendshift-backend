package cmd

import (
	"embed"
	"errors"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var fs embed.FS

var dsn string

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&dsn, "dsn", "d", "mysql://root@tcp(localhost:3306)/gti?parseTime=true", "the database connection dsn")
}

var migrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Args:  cobra.ExactArgs(1),
	Short: "Run database migration, valid args is either up or down",
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]

		d, err := iofs.New(fs, "migrations")
		if err != nil {
			log.Fatal(err)
		}

		m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
		if err != nil {
			log.Fatal(err)
		}

		if action == "up" {
			err = m.Up()
		} else if action == "down" {
			err = m.Down()
		} else {
			err = errors.New("unsupported action, only 'up' or 'down' are valida action")
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}
