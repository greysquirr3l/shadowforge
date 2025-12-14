package commands

import (
	"errors"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/distribution"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

var (
	ErrMissingPayload       = errors.New("payload data is required")
	ErrMissingCoverMedia    = errors.New("at least one cover media file is required")
	ErrInvalidThreshold     = errors.New("invalid threshold configuration")
	ErrManifestPathRequired = errors.New("manifest path is required")
	ErrInvalidShardCount    = errors.New("invalid shard count")
)

// EmbedDistributedCommand represents a request to embed data across multiple carriers (1:N pattern).
type EmbedDistributedCommand struct {
	InputFile        string                   `json:"input_file"`                // Path to payload file
	CoverFiles       []string                 `json:"cover_files"`               // Paths to cover media files
	OutputDirectory  string                   `json:"output_directory"`          // Directory for output files
	ManifestPath     string                   `json:"manifest_path"`             // Path to save manifest
	Technique        stego.StegoTechnique     `json:"technique,omitempty"`       // Auto-detect if empty
	Pattern          distribution.PatternType `json:"pattern,omitempty"`         // Default: OneToMany
	DataShards       int                      `json:"data_shards,omitempty"`     // Default: N-1 (auto-calculate)
	ParityShards     int                      `json:"parity_shards,omitempty"`   // Default: 1 or auto-calculate
	RequiredShards   int                      `json:"required_shards,omitempty"` // K of N (default: data_shards)
	Password         string                   `json:"password,omitempty"`
	RedundancyLevel  float64                  `json:"redundancy_level,omitempty"`  // 0.0-1.0, overrides parity calc
	ManifestPassword string                   `json:"manifest_password,omitempty"` // Separate password for manifest
	MaxWorkers       int                      `json:"max_workers,omitempty"`       // Goroutine pool size (default: NumCPU)
}

// CommandName returns the command identifier.
func (c EmbedDistributedCommand) CommandName() string {
	return "embed_distributed"
}

// Validate validates the distributed embed command parameters.
func (c EmbedDistributedCommand) Validate() error {
	if c.InputFile == "" {
		return ErrMissingPayload
	}
	if len(c.CoverFiles) == 0 {
		return ErrMissingCoverMedia
	}
	if c.OutputDirectory == "" {
		return ErrMissingOutputFile
	}
	if c.ManifestPath == "" {
		return ErrManifestPathRequired
	}

	// Validate shard configuration
	totalShards := len(c.CoverFiles)
	if c.DataShards > 0 && c.DataShards > totalShards {
		return fmt.Errorf("%w: data_shards (%d) exceeds cover files (%d)",
			ErrInvalidShardCount, c.DataShards, totalShards)
	}
	if c.ParityShards > 0 && c.ParityShards > totalShards {
		return fmt.Errorf("%w: parity_shards (%d) exceeds cover files (%d)",
			ErrInvalidShardCount, c.ParityShards, totalShards)
	}
	if c.DataShards > 0 && c.ParityShards > 0 && c.DataShards+c.ParityShards != totalShards {
		return fmt.Errorf("%w: data_shards (%d) + parity_shards (%d) != total (%d)",
			ErrInvalidShardCount, c.DataShards, c.ParityShards, totalShards)
	}

	// Validate threshold
	if c.RequiredShards > 0 {
		if c.RequiredShards > totalShards {
			return fmt.Errorf("%w: required_shards (%d) exceeds total shards (%d)",
				ErrInvalidThreshold, c.RequiredShards, totalShards)
		}
		if c.DataShards > 0 && c.RequiredShards > c.DataShards {
			return fmt.Errorf("%w: required_shards (%d) exceeds data_shards (%d)",
				ErrInvalidThreshold, c.RequiredShards, c.DataShards)
		}
	}

	// Validate redundancy level
	if c.RedundancyLevel < 0.0 || c.RedundancyLevel > 1.0 {
		return fmt.Errorf("redundancy level must be between 0.0 and 1.0, got %f", c.RedundancyLevel)
	}

	return nil
}

// EmbedDistributedResult represents the result of a distributed embed operation.
type EmbedDistributedResult struct {
	OutputFiles     []string                 `json:"output_files"`  // Paths to generated stego files
	ManifestPath    string                   `json:"manifest_path"` // Path to manifest file
	Technique       stego.StegoTechnique     `json:"technique"`
	Pattern         distribution.PatternType `json:"pattern"`
	PayloadSize     int64                    `json:"payload_size"`     // Original payload size
	TotalShards     int                      `json:"total_shards"`     // Total shards created
	DataShards      int                      `json:"data_shards"`      // Number of data shards
	ParityShards    int                      `json:"parity_shards"`    // Number of parity shards
	RequiredShards  int                      `json:"required_shards"`  // K of N threshold
	AverageCapacity float64                  `json:"average_capacity"` // Average capacity used (%)
	ProcessingTime  int64                    `json:"processing_time"`  // Total milliseconds
	ParallelWorkers int                      `json:"parallel_workers"` // Number of concurrent workers used
	EncryptionUsed  bool                     `json:"encryption_used"`
	ManifestSigned  bool                     `json:"manifest_signed"`         // HMAC signature applied
	FailedShards    []int                    `json:"failed_shards,omitempty"` // Indices of failed embeddings
}

// ExtractDistributedCommand represents a request to extract and reconstruct from distributed carriers.
type ExtractDistributedCommand struct {
	ManifestPath     string   `json:"manifest_path"` // Path to manifest file
	StegoFiles       []string `json:"stego_files"`   // Paths to available stego files
	OutputFile       string   `json:"output_file"`   // Path to save reconstructed payload
	Password         string   `json:"password,omitempty"`
	ManifestPassword string   `json:"manifest_password,omitempty"` // Separate password for manifest
	AllowPartial     bool     `json:"allow_partial,omitempty"`     // Allow recovery with exactly K shards
}

// CommandName returns the command identifier.
func (c ExtractDistributedCommand) CommandName() string {
	return "extract_distributed"
}

// Validate validates the distributed extract command parameters.
func (c ExtractDistributedCommand) Validate() error {
	if c.ManifestPath == "" {
		return ErrManifestPathRequired
	}
	if len(c.StegoFiles) == 0 {
		return errors.New("at least one stego file is required")
	}
	if c.OutputFile == "" {
		return ErrMissingOutputFile
	}
	return nil
}

// ExtractDistributedResult represents the result of a distributed extraction operation.
type ExtractDistributedResult struct {
	OutputFile      string `json:"output_file"`
	PayloadSize     int64  `json:"payload_size"`
	ShardsUsed      int    `json:"shards_used"`      // Number of shards successfully extracted
	ShardsRequired  int    `json:"shards_required"`  // K of N threshold
	ShardsTotal     int    `json:"shards_total"`     // N total shards
	ProcessingTime  int64  `json:"processing_time"`  // Milliseconds
	IntegrityPassed bool   `json:"integrity_passed"` // Manifest HMAC verified
	DecryptionUsed  bool   `json:"decryption_used"`
	RecoveryMode    string `json:"recovery_mode"`              // "full", "partial", "minimal"
	CorruptedShards []int  `json:"corrupted_shards,omitempty"` // Indices of corrupted shards
	MissingShards   []int  `json:"missing_shards,omitempty"`   // Indices of missing shards
}
