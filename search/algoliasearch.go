package search

import (
	"fmt"
	"strconv"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/model"
)

type repositoryDocument struct {
	ObjectID string `json:"objectID"`
	FullName string `json:"full_name"`
}

type developerDocument struct {
	ObjectID string `json:"objectID"`
	Username string `json:"username"`
}

type Algoliasearch struct {
	client *search.Client
}

func NewAlgoliasearch() *Algoliasearch {
	return &Algoliasearch{
		client: search.NewClient(config.AlgoliasearchAppId, config.AlgoliasearchApiKey),
	}
}

func (s *Algoliasearch) Search(query string, opts ...any) (SearchResults, error) {
	var searchResults SearchResults

	queries := []search.IndexedQuery{
		search.NewIndexedQuery(repositoryIndex, opt.Query(query), opt.HitsPerPage(5)),
		search.NewIndexedQuery(developerIndex, opt.Query(query), opt.HitsPerPage(5)),
	}

	res, err := s.client.MultipleQueries(queries, "")

	if err != nil {
		return searchResults, fmt.Errorf("failed to search all: %v", err)
	}

	for _, result := range res.Results {
		switch result.Index {
		case repositoryIndex:
			searchResults.Repositories = append(searchResults.Repositories, result.Hits...)
		case developerIndex:
			searchResults.Developers = append(searchResults.Developers, result.Hits...)
		}
	}

	return searchResults, nil
}

func (s *Algoliasearch) UpsertDevelopers(developers ...model.Developer) error {
	var documents []developerDocument

	for _, developer := range developers {
		var document developerDocument

		document.ObjectID = strconv.Itoa(developer.Id)
		document.Username = developer.Username

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		_, err := s.client.InitIndex(developerIndex).SaveObjects(documents)

		if err != nil {
			return fmt.Errorf("failed to save developer obejcts to algolia search: %v", err)
		}
	}

	return nil
}

func (s *Algoliasearch) UpsertRepositories(repositories ...model.GhRepository) error {
	var documents []repositoryDocument

	for _, repository := range repositories {
		var document repositoryDocument

		document.ObjectID = strconv.Itoa(repository.Id)
		document.FullName = repository.FullName

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		_, err := s.client.InitIndex(repositoryIndex).SaveObjects(documents)

		if err != nil {
			return fmt.Errorf("failed to save repositoriy obejcts to algolia search: %v", err)
		}
	}

	return nil
}

func (s *Algoliasearch) DeleteAll() error {
	_, err := s.client.InitIndex(repositoryIndex).ClearObjects()

	if err != nil {
		return fmt.Errorf("failed to clear repository obejcts in algolia search: %v", err)
	}

	_, err = s.client.InitIndex(developerIndex).ClearObjects()

	if err != nil {
		return fmt.Errorf("failed to clear developer objects in algolia search: %v", err)
	}

	return nil
}
