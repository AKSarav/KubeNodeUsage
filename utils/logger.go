package utils

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger *logrus.Logger

func InitLogger() {
	// Create and configure the logger
	logger := logrus.New()

	// Check for an environment variable to set the log level
	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel == "" {
		logLevel = "info" // Default log level
	}

	// Parse the log level from the environment variable
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		// Handle the error if the log level is invalid
		panic("Invalid LOG_LEVEL environment variable value")
	}

	// Set the log level
	logger.SetLevel(level)

	Logger = logger
}
