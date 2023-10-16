package model

import (
	"time"

	"github.com/liweiyi88/trendshift-backend/utils/dbutils"
)

type Developer struct {
	Id              int                `json:"developer_id"` // primary key saved in DB.
	GhId            int                `json:"id"`           // id from github repository api response.
	Username        string             `json:"login"`        // github api use login as username.
	AvatarUrl       string             `json:"avatar_url"`
	Name            dbutils.NullString `json:"name"`
	Company         dbutils.NullString `json:"company"`
	Blog            dbutils.NullString `json:"blog"`
	Location        dbutils.NullString `json:"location"`
	Email           dbutils.NullString `json:"email"`
	Bio             dbutils.NullString `json:"bio"`
	TwitterUsername dbutils.NullString `json:"twitter_username"`
	PublicRepos     int                `json:"public_repos"`
	PublicGists     int                `json:"public_gists"`
	Followers       int                `json:"followers"`
	Following       int                `json:"following"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}
