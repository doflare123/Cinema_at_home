package config

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	//go:embed genres.json
	JSONFiles embed.FS
)

type Config struct {
	AppEnv       string
	ServerPort   string
	DBDsn        string
	JWTSecretKey string
	LogLevel     string
	ApiKey       string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env файл не найден, использую системные переменные")
	}
	fmt.Print(JSONFiles)

	return &Config{
		AppEnv:       os.Getenv("APP_ENV"),
		ServerPort:   os.Getenv("PORT"),
		DBDsn:        os.Getenv("DSN"),
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		ApiKey:       os.Getenv("API_KEY"),
	}
}

func GetJSONFile(filename string) ([]byte, error) {
	return JSONFiles.ReadFile(filename)
}
