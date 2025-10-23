package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
)

const JWTMaxAge = 60 * time.Minute

var LanguageToScrape = []string{"", "javascript", "python", "go", "java", "php", "c++", "c", "typescript", "ruby", "c#", "rust", "dart", "swift", "kotlin"}

var (
	DatabaseDSN          string
	GitHubTokens         []string
	GinMode              string
	SignIngKey           string
	AlgoliasearchAppId   string
	AlgoliasearchApiKey  string
	MeilisearchMasterKey string
	MeilisearchHost      string
)

var githubTokens string

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
	githubTokens = os.Getenv("GITHUB_TOKENS")

	GitHubTokens = strings.Split(githubTokens, ",")

	GinMode = os.Getenv("GIN_MODE")
	SignIngKey = os.Getenv("SIGNING_KEY")
	MeilisearchMasterKey = os.Getenv("MEILISEARCH_MASTER_KEY")
	MeilisearchHost = os.Getenv("MEILISEARCH_HOST")

	AlgoliasearchAppId = os.Getenv("ALGOLIASEARCH_APPID")
	AlgoliasearchApiKey = os.Getenv("ALGOLIASEARCH_APIKEY")

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		TracesSampleRate: 1.0,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}
