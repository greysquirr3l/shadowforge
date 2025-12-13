package commands

import (
	"errors"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

var (
	ErrMissingInputFile  = errors.New("input file is required")
	ErrMissingCoverFile  = errors.New("cover file is required")
	ErrMissingOutputFile = errors.New("output file is required")
	ErrInvalidTechnique  = errors.New("invalid steganography technique")
)

// EmbedCommand represents a request to embed data into cover media.
type EmbedCommand struct {
	InputFile  string               `json:"input_file"`
	CoverFile  string               `json:"cover_file"`
	OutputFile string               `json:"output_file"`
	Technique  stego.StegoTechnique `json:"technique,omitempty"` // Auto-detect if empty
	Password   string               `json:"password,omitempty"`
	Redundancy float64              `json:"redundancy,omitempty"` // Reed-Solomon redundancy (0.0-1.0)
	Quality    int                  `json:"quality,omitempty"`    // Quality level (1-100)
}

// CommandName returns the command identifier.
func (c EmbedCommand) CommandName() string {
	return "embed"
}

// Validate validates the embed command parameters.
func (c EmbedCommand) Validate() error {
	if c.InputFile == "" {
		return ErrMissingInputFile
	}
	if c.CoverFile == "" {
		return ErrMissingCoverFile
	}
	if c.OutputFile == "" {
		return ErrMissingOutputFile
	}
	if c.Redundancy < 0.0 || c.Redundancy > 1.0 {
		return fmt.Errorf("redundancy must be between 0.0 and 1.0, got %f", c.Redundancy)
	}
	if c.Quality < 0 || c.Quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100, got %d", c.Quality)
	}
	return nil
}

// EmbedResult represents the result of an embed operation.
type EmbedResult struct {
	OutputFile      string               `json:"output_file"`
	Technique       stego.StegoTechnique `json:"technique"`
	PayloadSize     int64                `json:"payload_size"`
	CoverSize       int64                `json:"cover_size"`
	OutputSize      int64                `json:"output_size"`
	CapacityUsed    float64              `json:"capacity_used"`   // Percentage of capacity used
	QualityScore    float64              `json:"quality_score"`   // Embedding quality (0.0-1.0)
	ProcessingTime  int64                `json:"processing_time"` // Milliseconds
	EncryptionUsed  bool                 `json:"encryption_used"`
	CompressionUsed bool                 `json:"compression_used"`
}

// ExtractCommand represents a request to extract data from steganographic media.
type ExtractCommand struct {
	InputFile  string               `json:"input_file"`
	OutputFile string               `json:"output_file"`
	Technique  stego.StegoTechnique `json:"technique,omitempty"` // Auto-detect if empty
	Password   string               `json:"password,omitempty"`
}

// CommandName returns the command identifier.
func (c ExtractCommand) CommandName() string {
	return "extract"
}

// Validate validates the extract command parameters.
func (c ExtractCommand) Validate() error {
	if c.InputFile == "" {
		return ErrMissingInputFile
	}
	if c.OutputFile == "" {
		return ErrMissingOutputFile
	}
	return nil
}

// ExtractResult represents the result of an extract operation.
type ExtractResult struct {
	OutputFile        string               `json:"output_file"`
	Technique         stego.StegoTechnique `json:"technique"`
	PayloadSize       int64                `json:"payload_size"`
	InputSize         int64                `json:"input_size"`
	ProcessingTime    int64                `json:"processing_time"` // Milliseconds
	IntegrityPassed   bool                 `json:"integrity_passed"`
	DecryptionUsed    bool                 `json:"decryption_used"`
	DecompressionUsed bool                 `json:"decompression_used"`
}

// AnalyzeCapacityCommand represents a request to analyze embedding capacity.
type AnalyzeCapacityCommand struct {
	CoverFile string               `json:"cover_file"`
	Technique stego.StegoTechnique `json:"technique,omitempty"` // Analyze all if empty
}

// CommandName returns the command identifier.
func (c AnalyzeCapacityCommand) CommandName() string {
	return "analyze_capacity"
}

// Validate validates the analyze capacity command parameters.
func (c AnalyzeCapacityCommand) Validate() error {
	if c.CoverFile == "" {
		return ErrMissingCoverFile
	}
	return nil
}

// CapacityAnalysisResult represents capacity analysis results.
type CapacityAnalysisResult struct {
	CoverFile        string                    `json:"cover_file"`
	CoverSize        int64                     `json:"cover_size"`
	MediaType        string                    `json:"media_type"`
	Format           string                    `json:"format"`
	TechniqueResults []TechniqueCapacityResult `json:"technique_results"`
	Recommendations  []CapacityRecommendation  `json:"recommendations"`
}

// TechniqueCapacityResult represents capacity analysis for a specific technique.
type TechniqueCapacityResult struct {
	Technique         stego.StegoTechnique `json:"technique"`
	MaxCapacity       int64                `json:"max_capacity"`       // Theoretical maximum (bytes)
	SafeCapacity      int64                `json:"safe_capacity"`      // Recommended safe capacity
	QualityScore      float64              `json:"quality_score"`      // Quality score (0.0-1.0)
	DetectabilityRisk float64              `json:"detectability_risk"` // Risk score (0.0-1.0)
	PerformanceScore  float64              `json:"performance_score"`  // Speed score (0.0-1.0)
	Supported         bool                 `json:"supported"`
	ErrorReason       string               `json:"error_reason,omitempty"`
}

// CapacityRecommendation provides guidance for optimal embedding.
type CapacityRecommendation struct {
	Type       string               `json:"type"` // "best", "fastest", "safest", "warning"
	Technique  stego.StegoTechnique `json:"technique"`
	Reason     string               `json:"reason"`
	MaxPayload int64                `json:"max_payload"`
	Confidence float64              `json:"confidence"` // 0.0-1.0
}
