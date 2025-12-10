// Package main demonstrates usage of the Shadowforge logger.
package main

import (
	"errors"

	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set log level (optional, defaults to Info)
	logger.SetLevel(logrus.DebugLevel)

	// Simple logging
	logger.Info("Shadowforge starting...")
	logger.Debug("This is a debug message")
	logger.Warn("This is a warning")

	// Formatted logging
	logger.Infof("Processing %d files", 5)
	logger.Debugf("Current status: %s", "active")

	// Structured logging with fields
	logger.WithFields(logrus.Fields{
		"module":    "crypto",
		"operation": "encrypt",
		"algorithm": "kyber1024",
	}).Info("Encrypting payload")

	// Logging with single field
	logger.WithField("user", "admin").Info("User authenticated")

	// Error logging
	err := errors.New("connection timeout")
	logger.WithError(err).Error("Failed to connect to service")

	// Chaining fields
	logger.WithFields(logrus.Fields{
		"domain":  "distribution",
		"shards":  15,
		"pattern": "one-to-many",
	}).WithField("threshold", 10).Info("Distribution strategy created")

	logger.Info("Shadowforge operation complete")
}
