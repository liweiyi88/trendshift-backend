package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/model/opt"
	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
	"github.com/spf13/cobra"
)

var start string
var end string
var limit int
var recurring bool

// If run as cronjob, a suggested command to avoid sending too many requests to GitHub is
// `sync [repository|developer] --recurring=true --limit=500` and run it hourly.
// This makes sure that we update the latest repositories/developers details once a week.
func init() {
	rootCmd.AddCommand(gihtubSyncCmd)

	gihtubSyncCmd.Flags().StringVarP(&start, "start", "s", "", "--start \"2023-01-06 14:35:00\" ")
	gihtubSyncCmd.Flags().StringVarP(&end, "end", "e", "", "--end \"2023-10-06 14:35:00\" ")
	gihtubSyncCmd.Flags().IntVarP(&limit, "limit", "l", 0, "--limit=100")
	gihtubSyncCmd.Flags().BoolVarP(&recurring, "recurring", "r", false, "--recurring=true")
}

var gihtubSyncCmd = &cobra.Command{
	Use:   "sync [repository|developer]",
	Short: "Sync the latest repositories or developers details from GitHub",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		action := args[0]
		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)
		gh := github.NewClient(config.GitHubToken)

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
				sentry.CaptureException(err)
			}

			stop()
			sentry.Flush(2 * time.Second)
		}()

		if start != "" {
			_, err := time.Parse("2006-01-02 15:04:05", start)
			if err != nil {
				slog.Error("failed to parse start time", slog.Any("error", err))
				sentry.CaptureException(err)
				return
			}
		}

		// When running as cronjob, we only update entries whose last updated day is before the start of this week.
		// This avoid sending too much requests to GitHub.
		if recurring {
			startDayOfWeek := startDayOfWeek(time.Now())
			end = startDayOfWeek.Format("2006-01-02 15:04:05")
		} else {
			if end != "" {
				_, err := time.Parse("2006-01-02 15:04:05", end)
				if err != nil {
					slog.Error("failed to parse end time", slog.Any("error", err))
					sentry.CaptureException(err)
					return
				}
			}
		}

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		handler := github.NewSyncHandler(db, model.NewGhRepositoryRepo(db, dbutils.NewQueryBuilder()), gh)
		err := handler.Handle(ctx, action, opt.Start(start), opt.End(end), opt.Limit(limit))

		if err != nil {
			slog.Error("failed to handle sync action", slog.Any("error", err))
			sentry.CaptureException(err)
		}
	},
}

// Get the start day of the week. e.g. 2023-12-04 00:00:00
func startDayOfWeek(tm time.Time) time.Time {
	weekday := time.Duration(tm.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := tm.Date()
	currentZeroDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return currentZeroDay.Add(-1 * (weekday - 1) * 24 * time.Hour)
}
