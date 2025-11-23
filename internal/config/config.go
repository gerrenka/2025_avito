package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort     string
	PostgresConfig PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfig() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		PostgresConfig: PostgresConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "reviewuser"),
			Password: getEnv("DB_PASSWORD", "reviewpass"),
			DBName:   getEnv("DB_NAME", "reviewdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresConfig.Host,
		c.PostgresConfig.Port,
		c.PostgresConfig.User,
		c.PostgresConfig.Password,
		c.PostgresConfig.DBName,
		c.PostgresConfig.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
