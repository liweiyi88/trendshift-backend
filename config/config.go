package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const JWTMaxAge = 60 * time.Minute

var (
	DatabaseDSN string
	GitHubToken string
	GinMode     string
	SignIngKey  string
)

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
	GitHubToken = os.Getenv("GITHUB_TOKEN")
	GinMode = os.Getenv("GIN_MODE")
	SignIngKey = os.Getenv("SIGNING_KEY")
}
