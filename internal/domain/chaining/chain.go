package chaining

import (
	"context"
	"fmt"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// ChainMode defines the type of technique chaining
type ChainMode string

const (
	// Sequential applies techniques one after another (output becomes input)
	ChainModeSequential ChainMode = "sequential"

	// Layered embeds different data portions using different techniques in same carrier
	ChainModeLayered ChainMode = "layered"

	// Split distributes data across multiple carriers with different techniques
	ChainModeSplit ChainMode = "split"
)

// StegoTechnique is an alias for stego technique type
type StegoTechnique = stego.StegoTechnique

// ChainLink represents a single technique in the chain
type ChainLink struct {
	TechniqueID   string
	Technique     StegoTechnique
	Configuration map[string]interface{}
	Order         int
	Weight        float64 // For layered mode: portion of data (0.0-1.0)
}

// Chain represents a technique chain configuration
type Chain struct {
	ID          string
	Mode        ChainMode
	Links       []ChainLink
	CreatedAt   time.Time
	Description string
}

// ChainExecutionResult contains the result of chain execution
type ChainExecutionResult struct {
	ChainID       string
	Mode          ChainMode
	StepResults   []StepResult
	FinalOutput   []byte
	TotalDuration time.Duration
	Success       bool
	ErrorMessage  string
}

// StepResult represents the result of a single chain link execution
type StepResult struct {
	LinkOrder       int
	Technique       StegoTechnique
	InputSize       int64
	OutputSize      int64
	Duration        time.Duration
	Success         bool
	ErrorMessage    string
	IntermediateKey string // For debugging/recovery
}

// Validate ensures the chain configuration is valid
func (c *Chain) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("chain ID is required")
	}

	if len(c.Links) < 2 {
		return fmt.Errorf("chain must have at least 2 techniques")
	}

	if c.Mode == "" {
		return fmt.Errorf("chain mode is required")
	}

	// Validate mode-specific constraints
	switch c.Mode {
	case ChainModeSequential:
		return c.validateSequential()
	case ChainModeLayered:
		return c.validateLayered()
	case ChainModeSplit:
		return c.validateSplit()
	default:
		return fmt.Errorf("invalid chain mode: %s", c.Mode)
	}
}

func (c *Chain) validateSequential() error {
	// Check order is sequential
	for i, link := range c.Links {
		if link.Order != i {
			return fmt.Errorf("sequential chain links must have consecutive orders")
		}
		if link.Technique == "" {
			return fmt.Errorf("link %d missing technique", i)
		}
	}
	return nil
}

func (c *Chain) validateLayered() error {
	totalWeight := 0.0
	for i, link := range c.Links {
		if link.Weight <= 0 || link.Weight > 1.0 {
			return fmt.Errorf("link %d has invalid weight: %f (must be 0.0-1.0)", i, link.Weight)
		}
		totalWeight += link.Weight
	}

	// Allow small floating point error
	if totalWeight < 0.99 || totalWeight > 1.01 {
		return fmt.Errorf("layered chain weights must sum to 1.0, got: %f", totalWeight)
	}
	return nil
}

func (c *Chain) validateSplit() error {
	if len(c.Links) < 2 {
		return fmt.Errorf("split chain requires at least 2 techniques")
	}

	for i, link := range c.Links {
		if link.Technique == "" {
			return fmt.Errorf("link %d missing technique", i)
		}
	}
	return nil
}

// GetTotalLinks returns the number of techniques in the chain
func (c *Chain) GetTotalLinks() int {
	return len(c.Links)
}

// GetTechniqueByOrder returns the chain link at the specified order
func (c *Chain) GetTechniqueByOrder(order int) (*ChainLink, error) {
	for i := range c.Links {
		if c.Links[i].Order == order {
			return &c.Links[i], nil
		}
	}
	return nil, fmt.Errorf("no link found with order %d", order)
}

// ChainService defines the interface for chain execution
type ChainService interface {
	// ExecuteChain executes the full technique chain
	ExecuteChain(ctx context.Context, chain *Chain, payload []byte, carriers [][]byte) (*ChainExecutionResult, error)

	// ReverseChain extracts data by reversing the chain
	ReverseChain(ctx context.Context, chain *Chain, stegoMedia [][]byte) ([]byte, error)

	// EstimateChainCapacity calculates total capacity for a chain
	EstimateChainCapacity(ctx context.Context, chain *Chain, carriers [][]byte) (int64, error)

	// ValidateChain checks if a chain configuration is valid
	ValidateChain(ctx context.Context, chain *Chain) error
}
