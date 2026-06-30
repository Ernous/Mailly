package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr string
	ServerURL  string
	DBPath     string

	GoogleClientID     string
	GoogleClientSecret string

	MicrosoftClientID     string
	MicrosoftClientSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		ServerAddr:             getEnv("SERVER_ADDR", "127.0.0.1:3000"),
		ServerURL:              getEnv("SERVER_URL", "http://localhost:3000"),
		DBPath:                 getEnv("DB_PATH", "./data/mailly"),
		GoogleClientID:         getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", ""),
		MicrosoftClientID:      getEnv("MICROSOFT_CLIENT_ID", ""),
		MicrosoftClientSecret:  getEnv("MICROSOFT_CLIENT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
