package trending

import (
	"time"
)

type Owner struct {
	Name      string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

type GhRepository struct {
	Id        int    `json:"repository_id"` // primary key saved in DB.
	GhrId     int    `json:"id"`            // id from github repository api response.
	FullName  string `json:"full_name"`
	Owner     Owner  `json:"owner"`
	Forks     int    `json:"forks"`
	Stars     int    `json:"watchers"`
	Language  string `json:"language"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
