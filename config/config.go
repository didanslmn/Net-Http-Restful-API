package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	JWT   JWTConfig
	Redis RedisConfig
}

type AppConfig struct {
	Port string
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

type JWTConfig struct {
	PrivateKeyPath  string
	PublicKeyPath   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	Audience        string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig() (*Config, error) {
	return &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		// Database configuration
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			DBName:   getEnv("DB_NAME", "postgres"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "require"),
		},
		// JWT configuration
		JWT: JWTConfig{
			PrivateKeyPath:  getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
			PublicKeyPath:   getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
			AccessTokenTTL:  getEnvAsDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvAsDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			Issuer:          getEnv("ISSUER", "myapp"),
			Audience:        getEnv("AUDIENCE", "user-myapp"),
		},
		// Reis configuration
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}, nil
}

// Helper Function

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

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}
