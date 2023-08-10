package web

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/database"
	"github.com/liweiyi88/gti/global"
	"github.com/liweiyi88/gti/web/controller"
	"github.com/liweiyi88/gti/web/middleware"
	"golang.org/x/exp/slog"
)

type Controllers struct {
	repositoryController *controller.RepositoryController
	tagController        *controller.TagController
}

func initControllers(repositories *global.Repositories) *Controllers {
	return &Controllers{
		repositoryController: controller.NewRepositoryController(repositories.GhRepositoryRepo),
		tagController:        controller.NewTagController(repositories.TagRepo),
	}
}

func initialize(ctx context.Context) (*sql.DB, *Controllers) {
	db := database.GetInstance(ctx)
	repositories := global.InitRepositories(db)
	controllers := initControllers(repositories)

	return db, controllers
}

func setupRouter(controllers Controllers) *gin.Engine {
	gin.SetMode(config.GinMode)
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/api/tags", controllers.tagController.List)

	//TODO: login api, shall we return refresh token as well?
	auth := router.Group("/api")

	// JWT authentication protected routes.
	auth.Use(middleware.JwtAuth())
	auth.POST("/tags", controllers.tagController.Save).Use(middleware.JwtAuth())
	auth.POST("/repositories/:id/tags", controllers.repositoryController.SaveTags).Use(middleware.JwtAuth())

	return router
}

func Server() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	db, controllers := initialize(ctx)

	defer func() {
		err := db.Close()

		if err != nil {
			slog.Error("failed to close db: %v", err)
		}

		stop()
	}()

	router := setupRouter(*controllers)

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
		log.Fatal("server forced to shutdown: ", err)
	}

	slog.Info("server exit")
}
