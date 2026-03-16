package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port            string
	DBPath          string
	AdminUsername   string
	AdminPassword   string
	SlackWebhookURL string
	LogLevel        string
}

// LoadConfig creates and returns application configuration from environment variables
func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./serverscheduler.db"
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin"
	}

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return Config{
		Port:            port,
		DBPath:          dbPath,
		AdminUsername:   adminUsername,
		AdminPassword:   adminPassword,
		SlackWebhookURL: slackWebhookURL,
		LogLevel:        logLevel,
	}
}
