package commands

import (
	"errors"
	"os"
)

var (
	// Matrix-specific errors
	ErrNoMatrix           = errors.New("matrix configuration is required")
	ErrInvalidMatrixMode  = errors.New("invalid matrix allocation mode")
	ErrMatrixCapacity     = errors.New("insufficient total capacity for matrix distribution")
	ErrMatrixDimensions   = errors.New("matrix dimensions invalid (need N payloads and M covers)")
	ErrInsufficientCovers = errors.New("insufficient cover media for matrix distribution")
)

// MatrixMode defines the allocation strategy for many-to-many distribution.
type MatrixMode string

const (
	// MatrixModeRoundRobin distributes shards in round-robin fashion across covers
	MatrixModeRoundRobin MatrixMode = "round_robin"

	// MatrixModeRandom distributes shards randomly for unpredictable patterns
	MatrixModeRandom MatrixMode = "random"

	// MatrixModeOptimized uses capacity-aware allocation for optimal space usage
	MatrixModeOptimized MatrixMode = "optimized"

	// MatrixModeBalanced ensures even distribution across all covers
	MatrixModeBalanced MatrixMode = "balanced"
)

// MatrixPayloadItem represents a single payload in matrix distribution.
type MatrixPayloadItem struct {
	Name string `json:"name"` // Identifier for this payload
	Data []byte `json:"data"` // Payload data
	Path string `json:"path"` // Original file path (optional)
}

// MatrixCoverItem represents a single cover medium in the matrix.
type MatrixCoverItem struct {
	Path      string `json:"path"`                // Cover file path
	Technique string `json:"technique,omitempty"` // Override technique for this cover
}

// EmbedMatrixCommand represents a request to embed multiple payloads across multiple covers (N:M pattern).
type EmbedMatrixCommand struct {
	Payloads      []MatrixPayloadItem `json:"payloads"`      // N payloads
	Covers        []MatrixCoverItem   `json:"covers"`        // M cover media
	OutputDir     string              `json:"output_dir"`    // Directory for output stego media
	ManifestFile  string              `json:"manifest_file"` // Matrix manifest file
	Mode          MatrixMode          `json:"mode"`          // Allocation strategy
	Password      string              `json:"password,omitempty"`
	RSRedundancy  float64             `json:"rs_redundancy"`         // Redundancy level (0.2-0.5 recommended)
	MinRedundancy float64             `json:"min_redundancy"`        // Minimum cross-redundancy ratio
	Compression   bool                `json:"compression,omitempty"` // Compress payloads
}

// CommandName returns the command identifier.
func (c EmbedMatrixCommand) CommandName() string {
	return "embed_matrix"
}

// Validate validates the matrix embed command parameters.
func (c EmbedMatrixCommand) Validate() error {
	if len(c.Payloads) == 0 {
		return errors.New("at least one payload is required")
	}

	if len(c.Covers) == 0 {
		return errors.New("at least one cover is required")
	}

	// Matrix requires multiple payloads OR multiple covers (ideally both)
	if len(c.Payloads) == 1 && len(c.Covers) == 1 {
		return ErrMatrixDimensions
	}

	// Validate all payloads
	for i, payload := range c.Payloads {
		if payload.Name == "" {
			return errors.New("payload name is required")
		}
		if len(payload.Data) == 0 {
			return errors.New("payload data cannot be empty")
		}
		// Check for duplicates
		for j := 0; j < i; j++ {
			if c.Payloads[j].Name == payload.Name {
				return errors.New("duplicate payload name: " + payload.Name)
			}
		}
	}

	// Validate all covers exist
	for _, cover := range c.Covers {
		if cover.Path == "" {
			return errors.New("cover path is required")
		}
		if _, err := os.Stat(cover.Path); os.IsNotExist(err) {
			return errors.New("cover file does not exist: " + cover.Path)
		}
	}

	if c.OutputDir == "" {
		return errors.New("output directory is required")
	}

	if c.ManifestFile == "" {
		return errors.New("manifest file is required for matrix operations")
	}

	// Validate mode
	switch c.Mode {
	case MatrixModeRoundRobin, MatrixModeRandom, MatrixModeOptimized, MatrixModeBalanced:
		// Valid
	case "":
		return errors.New("matrix mode is required")
	default:
		return ErrInvalidMatrixMode
	}

	// Validate redundancy levels
	if c.RSRedundancy < 0 || c.RSRedundancy > 1.0 {
		return errors.New("RS redundancy must be between 0.0 and 1.0")
	}

	if c.MinRedundancy < 0 || c.MinRedundancy > 1.0 {
		return errors.New("minimum redundancy must be between 0.0 and 1.0")
	}

	// Recommended: At least 20% redundancy for matrix operations
	if c.RSRedundancy < 0.2 {
		// Warning: low redundancy, but still valid
	}

	return nil
}

// EmbedMatrixResult represents the result of a matrix embed operation.
type EmbedMatrixResult struct {
	OutputFiles       []string            `json:"output_files"`       // Generated stego files
	ManifestFile      string              `json:"manifest_file"`      // Matrix manifest
	PayloadCount      int                 `json:"payload_count"`      // N
	CoverCount        int                 `json:"cover_count"`        // M
	TotalShards       int                 `json:"total_shards"`       // Total shards generated
	ShardsPerPayload  map[string]int      `json:"shards_per_payload"` // Shard count by payload
	ShardsPerCover    map[string]int      `json:"shards_per_cover"`   // Shard count by cover
	CrossRedundancy   float64             `json:"cross_redundancy"`   // Actual cross-redundancy achieved
	MatrixDensity     float64             `json:"matrix_density"`     // Utilization percentage
	Mode              string              `json:"mode"`               // Allocation mode used
	ProcessingTime    int64               `json:"processing_time"`    // Milliseconds
	EncryptionUsed    bool                `json:"encryption_used"`
	CompressionUsed   bool                `json:"compression_used"`
	RSRedundancy      float64             `json:"rs_redundancy"`
	AllocationMetrics AllocationMetrics   `json:"allocation_metrics"`
	MatrixMap         MatrixAllocationMap `json:"matrix_map"` // Full allocation details
}

// AllocationMetrics provides statistics about the matrix allocation.
type AllocationMetrics struct {
	TotalCapacity       int64   `json:"total_capacity"`        // Sum of all cover capacities
	UsedCapacity        int64   `json:"used_capacity"`         // Actually used capacity
	UtilizationPercent  float64 `json:"utilization_percent"`   // Overall utilization
	MinCoverUtilization float64 `json:"min_cover_utilization"` // Least utilized cover
	MaxCoverUtilization float64 `json:"max_cover_utilization"` // Most utilized cover
	AvgCoverUtilization float64 `json:"avg_cover_utilization"` // Average utilization
	LoadBalanceScore    float64 `json:"load_balance_score"`    // 0-1, 1 = perfectly balanced
}

// MatrixAllocationMap describes the complete payload-to-shard-to-cover mapping.
type MatrixAllocationMap struct {
	Payloads map[string]PayloadAllocation `json:"payloads"` // Key: payload name
	Covers   map[string]CoverAllocation   `json:"covers"`   // Key: cover path
}

// PayloadAllocation describes how a single payload was distributed.
type PayloadAllocation struct {
	Name            string           `json:"name"`
	Size            int64            `json:"size"`
	DataShards      int              `json:"data_shards"`
	ParityShards    int              `json:"parity_shards"`
	ShardPlacements []ShardPlacement `json:"shard_placements"` // Where each shard went
}

// ShardPlacement describes where a single shard was embedded.
type ShardPlacement struct {
	ShardIndex int    `json:"shard_index"` // 0-based shard number
	CoverPath  string `json:"cover_path"`  // Which cover contains this shard
	ShardHash  string `json:"shard_hash"`  // SHA-256 of shard data
}

// CoverAllocation describes what was embedded in a single cover.
type CoverAllocation struct {
	Path           string           `json:"path"`
	Capacity       int64            `json:"capacity"`
	UsedCapacity   int64            `json:"used_capacity"`
	Utilization    float64          `json:"utilization"` // Percentage
	ShardCount     int              `json:"shard_count"`
	PayloadSources []string         `json:"payload_sources"` // Which payloads contributed shards
	Shards         []ShardReference `json:"shards"`          // All shards in this cover
}

// ShardReference identifies a shard within a cover.
type ShardReference struct {
	PayloadName string `json:"payload_name"`
	ShardIndex  int    `json:"shard_index"`
	Size        int64  `json:"size"`
	Offset      int64  `json:"offset"` // Byte offset within cover's embedded data
}

// ExtractMatrixCommand represents a request to extract payloads from matrix distribution.
type ExtractMatrixCommand struct {
	StegoFiles   []string `json:"stego_files"`   // Stego media files (subset of M is acceptable)
	ManifestFile string   `json:"manifest_file"` // Matrix manifest
	OutputDir    string   `json:"output_dir"`    // Output directory for recovered payloads
	Password     string   `json:"password,omitempty"`
	PayloadNames []string `json:"payload_names,omitempty"` // Optional: extract specific payloads
	AllowPartial bool     `json:"allow_partial"`           // Allow partial recovery with fewer shards
}

// CommandName returns the command identifier.
func (c ExtractMatrixCommand) CommandName() string {
	return "extract_matrix"
}

// Validate validates the matrix extract command parameters.
func (c ExtractMatrixCommand) Validate() error {
	if len(c.StegoFiles) == 0 {
		return errors.New("at least one stego file is required")
	}

	// Validate all stego files exist
	for _, file := range c.StegoFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return errors.New("stego file does not exist: " + file)
		}
	}

	if c.ManifestFile == "" {
		return errors.New("manifest file is required")
	}

	if _, err := os.Stat(c.ManifestFile); os.IsNotExist(err) {
		return errors.New("manifest file does not exist: " + c.ManifestFile)
	}

	if c.OutputDir == "" {
		return errors.New("output directory is required")
	}

	return nil
}

// ExtractMatrixResult represents the result of a matrix extraction operation.
type ExtractMatrixResult struct {
	OutputFiles           []string                  `json:"output_files"`
	PayloadCount          int                       `json:"payload_count"`      // Total payloads in matrix
	RecoveredPayloads     int                       `json:"recovered_payloads"` // Successfully recovered
	PartialPayloads       int                       `json:"partial_payloads"`   // Partially recovered
	FailedPayloads        int                       `json:"failed_payloads"`    // Could not recover
	StegoFilesUsed        int                       `json:"stego_files_used"`   // Number of stego files processed
	ShardsCollected       int                       `json:"shards_collected"`   // Total shards extracted
	ShardsRequired        int                       `json:"shards_required"`    // Total shards needed (all payloads)
	RecoveryRate          float64                   `json:"recovery_rate"`      // Percentage of data recovered
	ProcessingTime        int64                     `json:"processing_time"`    // Milliseconds
	DecryptionUsed        bool                      `json:"decryption_used"`
	PayloadRecoveryStatus map[string]RecoveryStatus `json:"payload_recovery_status"` // Per-payload status
}

// RecoveryStatus describes the recovery outcome for a single payload.
type RecoveryStatus struct {
	PayloadName       string  `json:"payload_name"`
	ShardsCollected   int     `json:"shards_collected"`
	ShardsRequired    int     `json:"shards_required"` // Minimum needed
	ShardsTotal       int     `json:"shards_total"`    // Total available
	RecoverySucceeded bool    `json:"recovery_succeeded"`
	RecoveryQuality   float64 `json:"recovery_quality"` // 0-1, based on shard count
	OutputFile        string  `json:"output_file,omitempty"`
	ErrorMessage      string  `json:"error_message,omitempty"`
}

// MatrixManifest represents the complete matrix distribution manifest.
// This is the master record of the entire N:M distribution.
type MatrixManifest struct {
	Version       string              `json:"version"` // Manifest version
	Mode          string              `json:"mode"`    // Allocation mode used
	PayloadCount  int                 `json:"payload_count"`
	CoverCount    int                 `json:"cover_count"`
	TotalShards   int                 `json:"total_shards"`
	RSRedundancy  float64             `json:"rs_redundancy"`
	Encrypted     bool                `json:"encrypted"`
	Compressed    bool                `json:"compressed"`
	AllocationMap MatrixAllocationMap `json:"allocation_map"`
	CreatedAt     string              `json:"created_at"`
	IntegrityHash string              `json:"integrity_hash"` // HMAC of manifest
}
