package config

import (
	"os"
	"strconv"
)

type Config struct {
	DB DBConfig
}

type DBConfig struct {
	ConnectionUrl string
	Host          string
	Port          int
	DBName        string
	User          string
	Password      string
	SSLMode       string
}

func LoadConfig() *Config {
	// Load database configuration from environment variables

	// Prioritaskan penggunaan DATABASE_URL jika tersedia
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return &Config{
			DB: DBConfig{
				ConnectionUrl: dbURL,
			},
		}
	}
	// Jika DATABASE_URL tidak tersedia, gunakan variabel individu
	DBconfig := &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			DBName:   getEnv("DB_NAME", "postgres"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "require"),
		},
	}
	return DBconfig
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultVal int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}
