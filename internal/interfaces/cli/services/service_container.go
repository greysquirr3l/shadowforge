// Package services provides service initialization and dependency injection for the CLI
package services

import (
	"github.com/greysquirr3l/shadowforge/pkg/logger"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/crypto"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
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

	return &ServiceContainer{
		CryptoService:          cryptoService,
		ErrorCorrectionService: errorCorrectionService,
		MediaService:           mediaService,
		StegoService:           stegoService,
		EmbedHandler:           embedHandler,
		ExtractHandler:         extractHandler,
		AnalyzeCapacityHandler: analyzeCapacityHandler,
	}, nil
}
