package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database          DatabaseConfig
	Redis             RedisConfig
	AppEnv            string
	AppPort           string
	ProductServiceURL string
	UserServiceURL    string
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "store_db"),
			Port:     dbPort,
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "3006"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
