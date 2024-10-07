package search

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/model"
)

const (
	repositoryIndex = "repositories"
	developerIndex  = "developers"
)

type SearchResults struct {
	Repositories []map[string]interface{} `json:"repositories"`
	Developers   []map[string]interface{} `json:"developers"`
}

const (
	sync   = "sync"
	delete = "delete"
)

type Search interface {
	UpsertDevelopers(developers ...model.Developer) error
	UpsertRepositories(repositories ...model.GhRepository) error
	DeleteAll() error
	Search(query string, opts ...any) (SearchResults, error)
}

func NewSearch() Search {
	return NewAlgoliasearch()
}

type SearchHandler struct {
	db     database.DB
	search Search
}

func NewSearchHandler(db database.DB, search Search) *SearchHandler {
	return &SearchHandler{
		db,
		search,
	}
}

func (h *SearchHandler) Handle(ctx context.Context, action string) error {
	switch action {
	case sync:
		return h.sync(ctx)
	case delete:
		return h.deleteAll()
	default:
		return errors.New("invalid search action")
	}
}

func (h *SearchHandler) sync(ctx context.Context) error {
	repositoryRepo := model.NewGhRepositoryRepo(h.db)

	var repositories []model.GhRepository
	var err error

	repositories, err = repositoryRepo.FindAll(ctx)

	if err != nil {
		return fmt.Errorf("could not retrieve repositories: %v", err)
	}

	err = h.search.UpsertRepositories(repositories...)

	if err != nil {
		return fmt.Errorf("could not import repositories to full text search: %v", err)
	}

	slog.Info("repositories have been imported")
	return nil
}

func (h *SearchHandler) deleteAll() error {
	err := h.search.DeleteAll()
	if err != nil {
		slog.Error("failed to delete all repositories and developers from full text search", slog.Any("error", err))
		return err
	}

	slog.Info("repositories and developers have been deleted")
	return nil
}
