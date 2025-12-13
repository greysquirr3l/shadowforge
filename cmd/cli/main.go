// Package main provides the CLI entry point for Shadowforge
package main

import (
	"context"
	"os"

	"github.com/greysquirr3l/shadowforge/internal/interfaces/cli/root"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// main is the entry point for the Shadowforge CLI application
func main() {
	// Initialize global logger
	globalLogger := logger.NewLogger()
	ctx := context.Background()

	// Create root command with all subcommands
	rootCmd, err := root.NewRootCommand(globalLogger)
	if err != nil {
		globalLogger.WithError(err).Error("Failed to create root command")
		os.Exit(1)
	}

	// Execute the CLI
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		globalLogger.WithError(err).Error("Command execution failed")
		os.Exit(1)
	}
}
