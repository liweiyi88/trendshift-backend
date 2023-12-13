package model

import (
	"time"
	"unicode/utf8"

	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

const maxDescriptionLength = 900

type Owner struct {
	Name      string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

type Trending struct {
	TrendingLanguage dbutils.NullString `json:"trending_language"`
	TrendDate        string             `json:"trend_date"`
	Rank             int                `json:"rank"`
}

type GhRepository struct {
	Id            int                `json:"repository_id"` // primary key saved in DB.
	GhrId         int                `json:"id"`            // id from github repository api response.
	FullName      string             `json:"full_name"`
	Owner         Owner              `json:"owner"`
	Forks         int                `json:"forks"`
	Stars         int                `json:"watchers"`
	Language      string             `json:"language"`
	Description   dbutils.NullString `json:"description"`
	DefaultBranch dbutils.NullString `json:"default_branch"`
	Homepage      dbutils.NullString `json:"homepage"`
	Tags          []Tag              `json:"tags"`
	Trendings     []Trending         `json:"trendings"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func (gr GhRepository) GetDescription() string {
	var description []rune
	suffix := []rune("...")

	var length int
	for _, r := range gr.Description.String {
		if length > maxDescriptionLength {
			description = append(description, suffix...)
			break
		}

		description = append(description, r)
		length += utf8.RuneLen(r)
	}

	return string(description)
}
