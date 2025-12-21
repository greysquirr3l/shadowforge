// Package watermark provides forensic watermarking domain logic
package watermark

import (
	"fmt"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// ForensicData contains watermark information for document tracking
type ForensicData struct {
	// Recipient name and email address
	Recipient string `json:"recipient"`

	// PreparedDate is when the document was prepared for this recipient
	PreparedDate time.Time `json:"prepared_date"`

	// GPGPublicKey is the ASCII-armored GPG public key for signing
	GPGPublicKey string `json:"gpg_public_key"`

	// DocumentID is an optional unique identifier for the document
	DocumentID string `json:"document_id,omitempty"`

	// WatermarkID is a unique identifier for this watermark instance
	WatermarkID string `json:"watermark_id"`

	// CreatedAt is when the watermark was created
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the forensic data is valid
func (f *ForensicData) Validate() error {
	if f.Recipient == "" {
		return fmt.Errorf("recipient is required")
	}
	if f.GPGPublicKey == "" {
		return fmt.Errorf("GPG public key is required")
	}
	if f.PreparedDate.IsZero() {
		return fmt.Errorf("prepared date is required")
	}
	return nil
}

// WatermarkStrategy defines how the watermark is distributed
type WatermarkStrategy string

const (
	// StrategyDistributed distributes watermark across multiple images
	StrategyDistributed WatermarkStrategy = "distributed"

	// StrategySingle embeds watermark in a single image
	StrategySingle WatermarkStrategy = "single"

	// StrategyRedundant embeds complete watermark in each image for redundancy
	StrategyRedundant WatermarkStrategy = "redundant"
)

// WatermarkConfig configures watermark embedding parameters
type WatermarkConfig struct {
	// Strategy determines distribution method
	Strategy WatermarkStrategy

	// Technique is the steganography technique to use
	Technique stego.StegoTechnique

	// Redundancy is the Reed-Solomon redundancy ratio (0.0-1.0)
	// e.g., 0.30 = 30% redundancy (10 data shards + 3 parity shards)
	Redundancy float64

	// Encrypt determines if payload should be encrypted
	Encrypt bool

	// Sign determines if payload should be digitally signed
	Sign bool
}

// DefaultConfig returns default watermark configuration
func DefaultConfig() *WatermarkConfig {
	return &WatermarkConfig{
		Strategy:   StrategyDistributed,
		Technique:  stego.LSB,
		Redundancy: 0.30, // 30% redundancy
		Encrypt:    true,
		Sign:       true,
	}
}
