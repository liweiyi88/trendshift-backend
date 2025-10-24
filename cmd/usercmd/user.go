package usercmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/global"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/spf13/cobra"
)

var username, password, role string

func init() {
	UserCmd.Flags().StringVarP(&username, "username", "u", "", "username")
	UserCmd.Flags().StringVarP(&password, "password", "p", "", "password")
	UserCmd.Flags().StringVarP(&role, "role", "r", "user", "role")

	UserCmd.MarkFlagRequired("username")
	UserCmd.MarkFlagRequired("password")
}

var UserCmd = &cobra.Command{
	Use:   "user:create",
	Short: "Create user",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		ctx, stop := context.WithCancel(context.Background())
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

		repositories := global.InitRepositories(db)

		var user model.User
		user.Username = username
		user.Role = role
		user.SetPassword(password)

		_, err := repositories.UserRepo.Save(ctx, user)

		if err != nil {
			log.Fatalf("failed to save user: %v", err)
		}
	},
}
