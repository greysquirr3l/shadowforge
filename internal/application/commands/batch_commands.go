package commands

import (
	"errors"
	"os"
)

var (
	// Batch-specific errors
	ErrNoPayloads       = errors.New("at least one payload is required")
	ErrPayloadTooLarge  = errors.New("combined payload exceeds cover capacity")
	ErrInvalidBatchMode = errors.New("invalid batch mode")
	ErrMissingStegoFile = errors.New("stego file not found")
)

// BatchPayloadItem represents a single payload in a batch operation.
type BatchPayloadItem struct {
	Name string `json:"name"` // Identifier for this payload
	Data []byte `json:"data"` // Payload data
	Path string `json:"path"` // Original file path (optional, for metadata)
}

// EmbedBatchCommand represents a request to embed multiple payloads in a single cover (N:1 pattern).
type EmbedBatchCommand struct {
	Payloads     []BatchPayloadItem `json:"payloads"`    // Multiple payloads to embed
	CoverFile    string             `json:"cover_file"`  // Single cover media file
	OutputFile   string             `json:"output_file"` // Output stego media file
	Technique    string             `json:"technique,omitempty"`
	Password     string             `json:"password,omitempty"`      // Optional encryption
	IndexFile    string             `json:"index_file,omitempty"`    // Path to save payload index
	Compression  bool               `json:"compression,omitempty"`   // Compress before embedding
	RSRedundancy float64            `json:"rs_redundancy,omitempty"` // Optional error correction
}

// CommandName returns the command identifier.
func (c EmbedBatchCommand) CommandName() string {
	return "embed_batch"
}

// Validate validates the batch embed command parameters.
func (c EmbedBatchCommand) Validate() error {
	if len(c.Payloads) == 0 {
		return ErrNoPayloads
	}

	// Validate all payload items
	for i, payload := range c.Payloads {
		if payload.Name == "" {
			return errors.New("payload name is required for indexing")
		}
		if len(payload.Data) == 0 {
			return errors.New("payload data cannot be empty")
		}
		if i > 0 {
			// Check for duplicate names
			for j := 0; j < i; j++ {
				if c.Payloads[j].Name == payload.Name {
					return errors.New("duplicate payload name: " + payload.Name)
				}
			}
		}
	}

	if c.CoverFile == "" {
		return ErrMissingCoverFile
	}

	// Check cover file exists
	if _, err := os.Stat(c.CoverFile); os.IsNotExist(err) {
		return errors.New("cover file does not exist: " + c.CoverFile)
	}

	if c.OutputFile == "" {
		return ErrMissingOutputFile
	}

	if c.RSRedundancy < 0 || c.RSRedundancy > 1.0 {
		return errors.New("RS redundancy must be between 0.0 and 1.0")
	}

	return nil
}

// EmbedBatchResult represents the result of a batch embed operation.
type EmbedBatchResult struct {
	OutputFile       string            `json:"output_file"`
	IndexFile        string            `json:"index_file,omitempty"`
	PayloadCount     int               `json:"payload_count"`
	TotalPayloadSize int64             `json:"total_payload_size"`
	CombinedSize     int64             `json:"combined_size"` // After aggregation
	CompressedSize   int64             `json:"compressed_size,omitempty"`
	CapacityUsed     float64           `json:"capacity_used"`   // Percentage
	ProcessingTime   int64             `json:"processing_time"` // Milliseconds
	Technique        string            `json:"technique"`
	EncryptionUsed   bool              `json:"encryption_used"`
	CompressionUsed  bool              `json:"compression_used"`
	RSRedundancy     float64           `json:"rs_redundancy,omitempty"`
	PayloadMetadata  []PayloadMetadata `json:"payload_metadata"`
}

// PayloadMetadata holds metadata about an embedded payload in batch mode.
type PayloadMetadata struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"` // Byte offset in combined payload
	Hash   string `json:"hash"`   // SHA-256 hash for verification
}

// ExtractBatchCommand represents a request to extract multiple payloads from a single cover (N:1 pattern).
type ExtractBatchCommand struct {
	StegoFile    string   `json:"stego_file"` // Stego media containing multiple payloads
	OutputDir    string   `json:"output_dir"` // Directory to extract payloads to
	IndexFile    string   `json:"index_file"` // Payload index file
	Password     string   `json:"password,omitempty"`
	PayloadNames []string `json:"payload_names,omitempty"` // Optional: extract specific payloads only
}

// CommandName returns the command identifier.
func (c ExtractBatchCommand) CommandName() string {
	return "extract_batch"
}

// Validate validates the batch extract command parameters.
func (c ExtractBatchCommand) Validate() error {
	if c.StegoFile == "" {
		return ErrMissingStegoFile
	}

	// Check stego file exists
	if _, err := os.Stat(c.StegoFile); os.IsNotExist(err) {
		return errors.New("stego file does not exist: " + c.StegoFile)
	}

	if c.IndexFile == "" {
		return errors.New("index file is required for batch extraction")
	}

	// Check index file exists
	if _, err := os.Stat(c.IndexFile); os.IsNotExist(err) {
		return errors.New("index file does not exist: " + c.IndexFile)
	}

	if c.OutputDir == "" {
		return errors.New("output directory is required")
	}

	return nil
}

// ExtractBatchResult represents the result of a batch extraction operation.
type ExtractBatchResult struct {
	OutputFiles        []string `json:"output_files"`
	PayloadCount       int      `json:"payload_count"`
	ExtractedPayloads  int      `json:"extracted_payloads"`
	FailedPayloads     int      `json:"failed_payloads"`
	TotalSize          int64    `json:"total_size"`
	ProcessingTime     int64    `json:"processing_time"` // Milliseconds
	DecryptionUsed     bool     `json:"decryption_used"`
	DecompressionUsed  bool     `json:"decompression_used"`
	VerificationPassed bool     `json:"verification_passed"` // Hash verification
}

// PayloadIndex represents the index file structure for batch operations.
// This is serialized to JSON and stored alongside the stego media.
type PayloadIndex struct {
	Version         string            `json:"version"` // Index format version
	PayloadCount    int               `json:"payload_count"`
	TotalSize       int64             `json:"total_size"`
	CompressionUsed bool              `json:"compression_used"`
	EncryptionUsed  bool              `json:"encryption_used"`
	RSRedundancy    float64           `json:"rs_redundancy,omitempty"`
	Payloads        []PayloadMetadata `json:"payloads"`
	CreatedAt       string            `json:"created_at"`
}
