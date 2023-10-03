package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/dbutils"
	"github.com/liweiyi88/gti/model"
	"github.com/liweiyi88/gti/search"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var delete bool

func init() {
	rootCmd.AddCommand(searchImportCmd)

	searchImportCmd.Flags().BoolVarP(&delete, "delete", "d", false, "if true, then delete all repositories.")
}

var searchImportCmd = &cobra.Command{
	Use:   "search-import",
	Short: "Import repositories to full text search",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		ctx, stop := context.WithCancel(context.Background())
		search := search.NewAlgoliasearch()

		db := database.GetInstance(ctx)

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
			}

			stop()
		}()

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		if delete {
			err := search.DeleteAllRepositories()
			if err != nil {
				slog.Error("failed to delete all repositories from full text search", slog.Any("error", err))
				return
			}
		} else {
			repositoryRepo := model.NewGhRepositoryRepo(db, dbutils.NewQueryBuilder())

			var repositories []model.GhRepository
			var err error

			repositories, err = repositoryRepo.FindAll(ctx, "", "", 0)

			if err != nil {
				slog.Error("could not retrieve repositories", slog.Any("error", err))
				return
			}

			err = search.UpsertRepositories(repositories...)

			if err != nil {
				slog.Error("could not import repositories to full text search", slog.Any("error", err))
				return
			}

			slog.Info("repositories have been imported")
		}
	},
}
