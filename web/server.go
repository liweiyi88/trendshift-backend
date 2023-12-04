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

	"log/slog"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/global"
	"github.com/liweiyi88/trendshift-backend/web/controller"
	"github.com/liweiyi88/trendshift-backend/web/middleware"
)

type Controllers struct {
	developerController  *controller.DeveloperController
	repositoryController *controller.RepositoryController
	tagController        *controller.TagController
	securityController   *controller.SecurityController
	statsController      *controller.StatsController
	searchController     *controller.SearchController
}

func initControllers(repositories *global.Repositories) *Controllers {
	return &Controllers{
		developerController:  controller.NewDeveloperController(&repositories.DeveloperRepo),
		repositoryController: controller.NewRepositoryController(repositories.GhRepositoryRepo),
		tagController:        controller.NewTagController(repositories.TagRepo),
		securityController:   controller.NewSecurityController(repositories.UserRepo),
		statsController:      controller.NewStatsController(repositories.StatsRepo),
		searchController:     controller.NewSearchController(),
	}
}

func setupRouter(ctx context.Context) (*gin.Engine, *sql.DB) {
	db := database.GetInstance(ctx)
	repositories := global.InitRepositories(db)
	controllers := initControllers(repositories)

	gin.SetMode(config.GinMode)
	router := gin.Default()

	// Use sentry to capture errors.
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	router.UseRawPath = true

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.POST("/login", controllers.securityController.Login)

	router.POST("/api/search", controllers.searchController.Search)
	router.GET("/api/trending-developers", controllers.developerController.GetTrendingDevelopers)
	router.GET("/api/trending-repositories", controllers.repositoryController.GetTrendingRepositories)
	router.GET("/api/developers/:id", controllers.developerController.Get)
	router.GET("/api/repositories", controllers.repositoryController.List)
	router.GET("/api/repositories/:id", controllers.repositoryController.Get)
	router.GET("/api/tags", controllers.tagController.List)
	router.GET("/api/stats/trending-topics", controllers.statsController.GetTrendingTopicsStats)

	// Protected routes.
	auth := router.Group("/api")
	auth.Use(middleware.JwtAuth())
	auth.POST("/tags", controllers.tagController.Save)
	auth.PUT("/repositories/:id/tags", controllers.repositoryController.SaveTags)

	return router, db
}

func Server() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	router, db := setupRouter(ctx)

	defer func() {
		err := db.Close()

		if err != nil {
			slog.Error("failed to close db", slog.Any("error", err))
		}

		stop()
	}()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
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
