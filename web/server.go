package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/global"
	"github.com/liweiyi88/gti/web/controller"
	"golang.org/x/exp/slog"
)

func Server() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	db := database.GetInstance(ctx)
	repositories := global.InitRepositories(db)
	repositoryController := controller.NewRepositoryController(repositories.GhRepositoryRepo)

	defer func() {
		err := db.Close()

		if err != nil {
			slog.Error("failed to close db: %v", err)
		}

		stop()
	}()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api")

	//TODO: create tag, search tags.
	//TODO: basic authentication for create tag and attach tags to repo.
	api.POST("/tags/", postTag)
	api.POST("/repositories/:id/tags", repositoryController.AttachTagsToRepository)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	stop()

	slog.Info("shutting down gracefully...")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	slog.Info("server exit")
}
