// Package services provides service initialization and dependency injection for the CLI
package services

import (
	"fmt"

	"github.com/greysquirr3l/shadowforge/pkg/logger"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/crypto"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/selection"
	stegoImpl "github.com/greysquirr3l/shadowforge/internal/infrastructure/stego_impl"
)

// ServiceContainer holds all initialized services for the CLI application.
// This provides dependency injection and ensures proper service lifecycle management.
type ServiceContainer struct {
	// Infrastructure Services
	CryptoService          *crypto.CirclCryptoService
	ErrorCorrectionService *errorcorrection.RSService
	MediaService           *media.MediaService
	StegoService           *stegoImpl.StegoService

	// Application Command Handlers
	EmbedHandler           *commands.EmbedHandler
	ExtractHandler         *commands.ExtractHandler
	AnalyzeCapacityHandler *commands.AnalyzeCapacityHandler
	ScanHandler            *commands.ScanDirectoryHandler
	SelectHandler          *commands.SelectCoversHandler
	SuggestHandler         *commands.GenerateSuggestionsHandler
}

// NewServiceContainer initializes all services with proper dependency injection.
// This is the single source of truth for service initialization in the CLI.
// Uses the global logger from pkg/logger.
func NewServiceContainer() (*ServiceContainer, error) {

	// Initialize infrastructure services
	cryptoService := crypto.NewCirclCryptoService(logger.Log)
	errorCorrectionService := errorcorrection.NewRSService(logger.Log)
	mediaService := media.NewMediaService(logger.Log)
	stegoService := stegoImpl.NewStegoService(mediaService, logger.Log)

	// Initialize application command handlers
	embedHandler := commands.NewEmbedHandler(
		stegoService,
		cryptoService,
		errorCorrectionService,
		mediaService,
		logger.Log,
	)

	extractHandler := commands.NewExtractHandler(
		stegoService,
		cryptoService,
		errorCorrectionService,
		mediaService,
		logger.Log,
	)

	analyzeCapacityHandler := commands.NewAnalyzeCapacityHandler(
		stegoService,
		mediaService,
		logger.Log,
	)

	// Initialize selection infrastructure
	directoryScanner := selection.NewDirectoryScanner(mediaService, stegoService)

	// Initialize selection command handlers
	scanHandler := commands.NewScanDirectoryHandler(directoryScanner)
	selectHandler := commands.NewSelectCoversHandler()
	suggestHandler := commands.NewGenerateSuggestionsHandler()

	// Validate all handlers were created successfully
	if embedHandler == nil {
		return nil, fmt.Errorf("failed to create embed handler")
	}
	if extractHandler == nil {
		return nil, fmt.Errorf("failed to create extract handler")
	}
	if analyzeCapacityHandler == nil {
		return nil, fmt.Errorf("failed to create analyze capacity handler")
	}

	return &ServiceContainer{
		CryptoService:          cryptoService,
		ErrorCorrectionService: errorCorrectionService,
		MediaService:           mediaService,
		StegoService:           stegoService,
		EmbedHandler:           embedHandler,
		ExtractHandler:         extractHandler,
		AnalyzeCapacityHandler: analyzeCapacityHandler,
		ScanHandler:            scanHandler,
		SelectHandler:          selectHandler,
		SuggestHandler:         suggestHandler,
	}, nil
}
