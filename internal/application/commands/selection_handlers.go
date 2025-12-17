package commands

import (
	"context"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/selection"
	selectionInfra "github.com/greysquirr3l/shadowforge/internal/infrastructure/selection"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// ScanDirectoryHandler handles directory scanning commands.
type ScanDirectoryHandler struct {
	scanner *selectionInfra.DirectoryScanner
}

// NewScanDirectoryHandler creates a new scan directory handler.
func NewScanDirectoryHandler(scanner *selectionInfra.DirectoryScanner) *ScanDirectoryHandler {
	return &ScanDirectoryHandler{scanner: scanner}
}

// Handle executes the scan directory command.
func (h *ScanDirectoryHandler) Handle(
	ctx context.Context,
	cmd ScanDirectoryCommand,
) (*selection.AnalysisResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"directory": cmd.Directory,
		"recursive": cmd.Recursive,
	}).Info("Handling scan directory command")

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	opts := selectionInfra.ScanOptions{
		Recursive:     cmd.Recursive,
		IncludeHidden: cmd.IncludeHidden,
		MaxDepth:      cmd.MaxDepth,
	}

	result, err := h.scanner.ScanDirectory(ctx, cmd.Directory, opts)
	if err != nil {
		return nil, fmt.Errorf("directory scan failed: %w", err)
	}

	logger.Log.WithFields(logrus.Fields{
		"files_found":    len(result.Files),
		"total_capacity": result.TotalSafeCapacity,
	}).Info("Directory scan completed")

	return result, nil
}

// SelectCoversHandler handles cover selection commands.
type SelectCoversHandler struct{}

// NewSelectCoversHandler creates a new select covers handler.
func NewSelectCoversHandler() *SelectCoversHandler {
	return &SelectCoversHandler{}
}

// Handle executes the select covers command.
func (h *SelectCoversHandler) Handle(
	ctx context.Context,
	cmd SelectCoversCommand,
) (*selection.SelectionPlan, error) {
	logger.Log.WithFields(logrus.Fields{
		"available_files": len(cmd.AvailableFiles),
		"payload_size":    cmd.PayloadSize,
	}).Info("Handling select covers command")

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	constraints := selection.SelectionConstraints{
		MaxFiles:          cmd.MaxFiles,
		MaxPerType:        cmd.MaxPerType,
		TargetUtilization: cmd.TargetUtilization,
		PreferDiversity:   cmd.PreferDiversity,
		MinimizeDetect:    cmd.MinimizeDetect,
	}

	selector := selection.NewOptimalSelector(constraints)
	plan, err := selector.SelectCovers(cmd.AvailableFiles, cmd.PayloadSize)
	if err != nil {
		return nil, fmt.Errorf("cover selection failed: %w", err)
	}

	logger.Log.WithFields(logrus.Fields{
		"selected_files":  len(plan.SelectedFiles),
		"utilization":     fmt.Sprintf("%.1f%%", plan.Utilization*100),
		"diversity_score": fmt.Sprintf("%.2f", plan.DiversityScore),
	}).Info("Cover selection completed")

	return plan, nil
}

// GenerateSuggestionsHandler handles suggestion generation commands.
type GenerateSuggestionsHandler struct{}

// NewGenerateSuggestionsHandler creates a new suggestions handler.
func NewGenerateSuggestionsHandler() *GenerateSuggestionsHandler {
	return &GenerateSuggestionsHandler{}
}

// Handle executes the generate suggestions command.
func (h *GenerateSuggestionsHandler) Handle(
	ctx context.Context,
	cmd GenerateSuggestionsCommand,
) (*SuggestionResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"required":  cmd.RequiredCapacity,
		"available": cmd.AvailableCapacity,
	}).Info("Handling generate suggestions command")

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	engine := selection.NewSuggestionEngine()
	suggestions := engine.GenerateSuggestions(cmd.RequiredCapacity, cmd.AvailableCapacity)

	result := &SuggestionResult{
		RequiredCapacity:  suggestions.RequiredCapacity,
		AvailableCapacity: suggestions.AvailableCapacity,
		CapacityGap:       suggestions.CapacityGap,
		MediaSuggestions:  suggestions.MediaSuggestions,
		Alternatives:      suggestions.Alternatives,
	}

	logger.Log.WithFields(logrus.Fields{
		"shortfall":    result.CapacityGap,
		"suggestions":  len(result.MediaSuggestions),
		"alternatives": len(result.Alternatives),
	}).Info("Suggestions generated")

	return result, nil
}
