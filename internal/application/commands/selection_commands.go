package commands

import (
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/selection"
)

// ScanDirectoryCommand represents a command to scan a directory for media files.
type ScanDirectoryCommand struct {
	Directory     string
	Recursive     bool
	IncludeHidden bool
	MaxDepth      int
}

// Validate ensures the command is valid.
func (c ScanDirectoryCommand) Validate() error {
	if c.Directory == "" {
		return fmt.Errorf("directory path is required")
	}
	if c.MaxDepth < 0 {
		return fmt.Errorf("max depth cannot be negative")
	}
	return nil
}

// SelectCoversCommand represents a command to select optimal cover files.
type SelectCoversCommand struct {
	AvailableFiles    []*selection.MediaAnalysis
	PayloadSize       int64
	MaxFiles          int
	MaxPerType        int
	TargetUtilization float64
	PreferDiversity   bool
	MinimizeDetect    bool
}

// Validate ensures the command is valid.
func (c SelectCoversCommand) Validate() error {
	if len(c.AvailableFiles) == 0 {
		return fmt.Errorf("no available files provided")
	}
	if c.PayloadSize <= 0 {
		return fmt.Errorf("payload size must be positive")
	}
	if c.TargetUtilization < 0 || c.TargetUtilization > 1 {
		return fmt.Errorf("target utilization must be between 0 and 1")
	}
	return nil
}

// GenerateSuggestionsCommand represents a command to generate capacity suggestions.
type GenerateSuggestionsCommand struct {
	RequiredCapacity  int64
	AvailableCapacity int64
}

// Validate ensures the command is valid.
func (c GenerateSuggestionsCommand) Validate() error {
	if c.RequiredCapacity <= 0 {
		return fmt.Errorf("required capacity must be positive")
	}
	if c.AvailableCapacity < 0 {
		return fmt.Errorf("available capacity cannot be negative")
	}
	return nil
}

// SuggestionResult represents the result of suggestion generation.
type SuggestionResult struct {
	RequiredCapacity  int64
	AvailableCapacity int64
	CapacityGap       int64
	MediaSuggestions  []*selection.MediaSuggestion
	Alternatives      []string
}
