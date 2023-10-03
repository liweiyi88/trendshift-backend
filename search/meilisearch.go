package search

import (
	"fmt"

	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/model"
	"github.com/meilisearch/meilisearch-go"
)

type Meilisearch struct {
	client *meilisearch.Client
}

func NewMeilisearch() *Meilisearch {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.MeilisearchHost,
		APIKey: config.MeilisearchMasterKey,
	})

	return &Meilisearch{
		client: client,
	}
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
		_, err := search.client.Index(repositoryIndex).UpdateDocuments(documents, "id")

		if err != nil {
			return fmt.Errorf("failed to upsert repositories: %v", err)
		}

	}

	return nil
}

func (search *Meilisearch) DeleteAllRepositories() error {
	_, err := search.client.Index(repositoryIndex).DeleteAllDocuments()

	if err != nil {
		return fmt.Errorf("failed to delete all repositories: %v", err)
	}

	return nil
}
