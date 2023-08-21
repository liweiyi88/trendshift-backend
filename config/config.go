package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	DatabaseDSN      string
	GitHubToken      string
	GinMode          string
	SignIngKey       string
	JWTCookieName    string
	JWTCookieMaxAge  time.Duration
	CORSAllowOrigins []string
)

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
	GitHubToken = os.Getenv("GITHUB_TOKEN")
	GinMode = os.Getenv("GIN_MODE")
	SignIngKey = os.Getenv("SIGNING_KEY")
	JWTCookieName = "gti_access_token"
	JWTCookieMaxAge = 15 * time.Minute
	CORSAllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
}
