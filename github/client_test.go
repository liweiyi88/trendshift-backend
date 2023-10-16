package github

import (
	"context"
	"testing"

	"github.com/liweiyi88/gti/model"
)

func TestGetDeveloper(t *testing.T) {
	client := Client{}

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
	client := Client{}

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
