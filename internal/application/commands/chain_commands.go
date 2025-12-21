package commands

import (
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/chaining"
)

// CreateChainCommand creates a new technique chain
type CreateChainCommand struct {
	Mode        chaining.ChainMode
	Links       []ChainLinkDTO
	Description string
}

// ChainLinkDTO represents a chain link in the command
type ChainLinkDTO struct {
	TechniqueID   string
	Technique     string
	Configuration map[string]interface{}
	Order         int
	Weight        float64
}

// CreateChainResult contains the result of chain creation
type CreateChainResult struct {
	ChainID   string
	Success   bool
	Error     string
	CreatedAt time.Time
}

// ExecuteChainCommand executes a technique chain
type ExecuteChainCommand struct {
	ChainID      string
	PayloadPath  string
	CarrierPaths []string
	OutputPath   string
}

// ExecuteChainResult contains the result of chain execution
type ExecuteChainResult struct {
	ChainID       string
	Success       bool
	OutputPaths   []string
	StepsSummary  []StepSummary
	TotalDuration time.Duration
	Error         string
}

// StepSummary summarizes a single chain step
type StepSummary struct {
	Order     int
	Technique string
	Duration  time.Duration
	Success   bool
}

// ReverseChainCommand reverses a technique chain for extraction
type ReverseChainCommand struct {
	ChainID         string
	StegoMediaPaths []string
	OutputPath      string
}

// ReverseChainResult contains the result of chain reversal
type ReverseChainResult struct {
	Success       bool
	ExtractedPath string
	ExtractedSize int64
	Error         string
}

// ValidateChainCommand validates a chain configuration
type ValidateChainCommand struct {
	ChainID string
}

// ValidateChainResult contains validation results
type ValidateChainResult struct {
	Valid             bool
	Errors            []string
	Warnings          []string
	EstimatedCapacity int64
}
