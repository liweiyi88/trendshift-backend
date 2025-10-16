package search

import (
	"fmt"

	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/meilisearch/meilisearch-go"
)

type Meilisearch struct {
	client meilisearch.ServiceManager
}

func NewMeilisearch() *Meilisearch {
	serviceManager := meilisearch.New(config.MeilisearchHost, meilisearch.WithAPIKey(config.MeilisearchMasterKey))
	return &Meilisearch{
		client: serviceManager,
	}
}

func (search *Meilisearch) Search(query string, opts ...any) (SearchResults, error) {
	var searchResults SearchResults
	// @TODO
	return searchResults, nil
}

func (search *Meilisearch) UpsertDevelopers(developers ...model.Developer) error {
	var documents []map[string]any

	for _, developer := range developers {
		document := make(map[string]any, 1)
		document["id"] = developer.Id
		document["username"] = developer.Username

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		primayKey := "id"
		_, err := search.client.Index(developerIndex).UpdateDocuments(documents, &primayKey)

		if err != nil {
			return fmt.Errorf("failed to upsert developers: %v", err)
		}
	}

	return nil
}

func (search *Meilisearch) UpsertRepositories(repositories ...model.GhRepository) error {
	var documents []map[string]any

	for _, repository := range repositories {
		document := make(map[string]any, 1)
		document["id"] = repository.Id
		document["full_name"] = repository.FullName

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		primayKey := "id"
		_, err := search.client.Index(repositoryIndex).UpdateDocuments(documents, &primayKey)

		if err != nil {
			return fmt.Errorf("failed to upsert repositories: %v", err)
		}

	}

	return nil
}

func (search *Meilisearch) DeleteAll() error {
	_, err := search.client.Index(repositoryIndex).DeleteAllDocuments()

	if err != nil {
		return fmt.Errorf("failed to delete all repositories: %v", err)
	}

	_, err = search.client.Index(developerIndex).DeleteAllDocuments()

	if err != nil {
		return fmt.Errorf("failed to delete all developers: %v", err)
	}

	return nil
}
