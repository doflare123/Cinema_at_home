package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	ServerPort   string
	DBDsn        string
	JWTSecretKey string
	LogLevel     string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, использую системные переменные")
	}

	return &Config{
		AppEnv:       os.Getenv("APP_ENV"),
		ServerPort:   os.Getenv("PORT"),
		DBDsn:        os.Getenv("DSN"),
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
