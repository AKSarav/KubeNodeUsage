package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger() {
	// If logger is already initialized, don't initialize again
	if Logger != nil {
		return
	}

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
		// Handle the error if the log level is invalid by using info level instead of panicking
		level = logrus.InfoLevel
	}

	// Set the log level
	logger.SetLevel(level)

	Logger = logger
}
