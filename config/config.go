package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	DatabaseDSN string
	GitHubToken string
	GinMode     string
)

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
	GitHubToken = os.Getenv("GITHUB_TOKEN")
	GinMode = os.Getenv("GIN_MODE")
}
