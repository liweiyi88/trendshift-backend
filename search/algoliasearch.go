package search

import (
	"fmt"
	"strconv"

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

func (search *Algoliasearch) SearchRepositories(query string, opt ...any) ([]map[string]interface{}, error) {
	res, err := search.client.InitIndex(repositoryIndex).Search(query, opt...)

	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %v", err)
	}

	return res.Hits, nil
}

func (search *Algoliasearch) UpsertDevelopers(developers ...model.Developer) error {
	var documents []developerDocument

	for _, developer := range developers {
		var document developerDocument

		document.ObjectID = strconv.Itoa(developer.Id)
		document.Username = developer.Username

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		_, err := search.client.InitIndex(developerIndex).SaveObjects(documents)

		if err != nil {
			return fmt.Errorf("failed to save developer obejcts to algolia search: %v", err)
		}
	}

	return nil
}

func (search *Algoliasearch) UpsertRepositories(repositories ...model.GhRepository) error {
	var documents []repositoryDocument

	for _, repository := range repositories {
		var document repositoryDocument

		document.ObjectID = strconv.Itoa(repository.Id)
		document.FullName = repository.FullName

		documents = append(documents, document)
	}

	if len(documents) > 0 {
		_, err := search.client.InitIndex(repositoryIndex).SaveObjects(documents)

		if err != nil {
			return fmt.Errorf("failed to save repositoriy obejcts to algolia search: %v", err)
		}
	}

	return nil
}

func (search *Algoliasearch) DeleteAll() error {
	_, err := search.client.InitIndex(repositoryIndex).ClearObjects()

	if err != nil {
		return fmt.Errorf("failed to clear repository obejcts in algolia search: %v", err)
	}

	_, err = search.client.InitIndex(developerIndex).ClearObjects()

	if err != nil {
		return fmt.Errorf("failed to clear developer objects in algolia search: %v", err)
	}

	return nil
}
