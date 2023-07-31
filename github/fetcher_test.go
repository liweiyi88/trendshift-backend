package github

import (
	"context"
	"testing"
)

func TestFetchRepository(t *testing.T) {
	fetcher := Fetcher{}

	ghRepo, err := fetcher.FetchRepository(context.Background(), "liweiyi88/onedump")
	if err != nil {
		t.Error(err)
	}

	expect := GhRepository{
		Id:       0,
		GhrId:    540829453,
		FullName: "liweiyi88/onedump",
		Owner: Owner{
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
