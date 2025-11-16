package github

import (
	"context"
	"testing"

	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/stretchr/testify/assert"
)

func TestParseNextLink(t *testing.T) {
	nextlink := parseNextLink(`<https://api.github.com/repositories/90194616/commits?per_page=1&page=2>; rel="next", <https://api.github.com/repositories/90194616/commits?per_page=1&page=5282>; rel="last"`)
	assert.Equal(t, "https://api.github.com/repositories/90194616/commits?per_page=1&page=2", nextlink)
	assert.Empty(t, parseNextLink(""))
}

func TestGetDeveloper(t *testing.T) {
	client := NewClient(NewTokenPool([]string{}, WithAllowEmptytoken(true)))

	developer, err := client.GetDeveloper(context.Background(), "liweiyi88")
	if err != nil {
		t.Error(err)
	}

	expect := model.Developer{
		Id:        0,
		GhId:      7248260,
		Username:  "liweiyi88",
		AvatarUrl: "https://avatars.githubusercontent.com/u/7248260?v=4",
	}

	if developer.Username != expect.Username {
		t.Errorf("expect: %v but got :%v", developer, expect)
	}

	if developer.GhId != expect.GhId {
		t.Errorf("expect: %v but got :%v", developer, expect)
	}

	if developer.AvatarUrl != expect.AvatarUrl {
		t.Errorf("expect: %v but got :%v", developer, expect)
	}
}

func TestGetRepository(t *testing.T) {
	client := NewClient(NewTokenPool([]string{}, WithAllowEmptytoken(true)))

	ghRepo, err := client.GetRepository(context.Background(), "liweiyi88/onedump")
	if err != nil {
		t.Error(err)
	}

	expect := model.GhRepository{
		Id:       0,
		GhrId:    540829453,
		FullName: "liweiyi88/onedump",
		Owner: model.Owner{
			Name:      "liweiyi88",
			AvatarUrl: "https://avatars.githubusercontent.com/u/7248260?v=4",
		},
		Language: "Go",
	}

	if ghRepo.FullName != expect.FullName {
		t.Errorf("expect: %v but got :%v", ghRepo, expect)
	}

	if ghRepo.Language != expect.Language {
		t.Errorf("expect: %v but got :%v", ghRepo, expect)
	}

	if ghRepo.Owner.Name != expect.Owner.Name {
		t.Errorf("expect: %v but got :%v", ghRepo, expect)
	}

	if ghRepo.Owner.AvatarUrl != expect.Owner.AvatarUrl {
		t.Errorf("expect: %v but got :%v", ghRepo, expect)
	}
}
