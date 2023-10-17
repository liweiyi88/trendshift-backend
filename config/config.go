package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const JWTMaxAge = 60 * time.Minute

var LanguageToScrape = []string{"", "javascript", "python", "go", "java", "php", "c++", "c", "typescript", "ruby", "c#", "rust"}

var (
	DatabaseDSN          string
	GitHubToken          string
	GinMode              string
	SignIngKey           string
	AlgoliasearchAppId   string
	AlgoliasearchApiKey  string
	MeilisearchMasterKey string
	MeilisearchHost      string
)

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
	GitHubToken = os.Getenv("GITHUB_TOKEN")
	GinMode = os.Getenv("GIN_MODE")
	SignIngKey = os.Getenv("SIGNING_KEY")
	MeilisearchMasterKey = os.Getenv("MEILISEARCH_MASTER_KEY")
	MeilisearchHost = os.Getenv("MEILISEARCH_HOST")

	AlgoliasearchAppId = os.Getenv("ALGOLIASEARCH_APPID")
	AlgoliasearchApiKey = os.Getenv("ALGOLIASEARCH_APIKEY")
}
