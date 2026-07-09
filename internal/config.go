package internal

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                   string
	AppPort                  string
	MongoURI                 string
	MongoDatabase            string
	MongoScrapeLogCollection string
	RodHeadless              bool
	RodNavigationTimeout     int
	RodActionTimeout         int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                   getEnv("APP_ENV", "development"),
		AppPort:                  getEnv("APP_PORT", "8080"),
		MongoURI:                 getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:            getEnv("MONGO_DATABASE", "scraping_service"),
		MongoScrapeLogCollection: getEnv("MONGO_SCRAPE_LOG_COLLECTION", "scrape_logs"),
		RodHeadless:              getEnvBool("ROD_HEADLESS", false),
		RodNavigationTimeout:     getEnvInt("ROD_NAVIGATION_TIMEOUT_SECONDS", 30),
		RodActionTimeout:         getEnvInt("ROD_ACTION_TIMEOUT_SECONDS", 15),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fallback
		}
		return b
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
