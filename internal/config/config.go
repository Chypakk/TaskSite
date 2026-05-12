package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBDriver   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
}

func Load() (Config, error) {
	envPath := findEnvFile()

	if err := godotenv.Load(envPath); err != nil {
		return Config{}, fmt.Errorf("failed to parse .env: %w", err)
	}

	cfg := Config{
		ServerPort: os.Getenv("SERVER_PORT"),
		DBDriver:   os.Getenv("DB_DRIVER"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPass:     os.Getenv("DB_PASS"),
		DBName:     os.Getenv("DB_NAME"),
	}

	return cfg, nil
}

// поиск env файла
func findEnvFile() string {
	// Рядом с exe
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// В текущей рабочей директории
	if cwd, err := os.Getwd(); err == nil {
		envPath := filepath.Join(cwd, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	return ""
}
