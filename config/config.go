package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	DatabaseDSN string
)

func Init() {
	godotenv.Load(".env.local")
	godotenv.Load(".env")

	DatabaseDSN = os.Getenv("DATABASE_DSN")
}
