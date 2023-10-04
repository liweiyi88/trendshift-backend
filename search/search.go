package search

import (
	"github.com/liweiyi88/gti/model"
)

const repositoryIndex = "repositories"

type Search interface {
	UpsertRepositories(repositories ...model.GhRepository) error
	DeleteAllRepositories() error
	SearchRepositories(query string, opt ...any) ([]map[string]interface{}, error)
}

func NewSearch() Search {
	return NewAlgoliasearch()
}
