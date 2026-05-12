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

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"SERVER_PORT": c.ServerPort,
		"DB_DRIVER":   c.DBDriver,
		"DB_HOST":     c.DBHost,
		"DB_PORT":     c.DBPort,
		"DB_USER":     c.DBUser,
		"DB_NAME":     c.DBName,
		// DB_PASS не проверяем на пустоту, т.к. иногда пароль действительно пустой
	}

	for k, v := range required {
		if v == "" {
			return fmt.Errorf("missing required env var: %s", k)
		}
	}
	return nil
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
