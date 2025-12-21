package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/internal/domain/watermark"
	"github.com/greysquirr3l/shadowforge/internal/interfaces/cli/services"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// NewWatermarkCommand creates the watermark command and subcommands
func NewWatermarkCommand(container *services.ServiceContainer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watermark",
		Short: "Forensic watermarking operations for confidential documents",
		Long: `Embed forensic tracking watermarks in PNG images using distributed
steganography with post-quantum encryption and Reed-Solomon error correction.

Watermarks include:
  - Recipient identifier
  - Preparation timestamp
  - GPG public signing key
  - Optional document ID`,
		Example: `  # Embed watermark across multiple PNGs
  shadowforge watermark embed \
    --recipient "John Doe <john@example.com>" \
    --gpg-key ./signing-key.pub \
    --input-dir ./equations \
    --output-dir ./watermarked

  # Extract watermark from images
  shadowforge watermark extract \
    --input-dir ./watermarked \
    --output ./recovered.json

  # Verify watermark authenticity
  shadowforge watermark verify \
    --input-dir ./watermarked \
    --expected-recipient "John Doe"`,
	}

	// Add subcommands
	cmd.AddCommand(newWatermarkEmbedCommand(container))
	cmd.AddCommand(newWatermarkExtractCommand(container))
	cmd.AddCommand(newWatermarkVerifyCommand(container))

	return cmd
}

// newWatermarkEmbedCommand creates the embed subcommand
func newWatermarkEmbedCommand(container *services.ServiceContainer) *cobra.Command {
	var (
		recipient  string
		gpgKeyPath string
		inputDir   string
		outputDir  string
		technique  string
		redundancy float64
		documentID string
		strategy   string
		noEncrypt  bool
		noSign     bool
	)

	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Embed forensic watermark across PNG images",
		Long: `Distributes an encrypted forensic watermark across multiple PNG images
using LSB or DCT steganography with Reed-Solomon redundancy.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate required parameters
			if recipient == "" {
				return fmt.Errorf("--recipient is required")
			}
			if gpgKeyPath == "" {
				return fmt.Errorf("--gpg-key is required")
			}
			if inputDir == "" {
				return fmt.Errorf("--input-dir is required")
			}
			if outputDir == "" {
				return fmt.Errorf("--output-dir is required")
			}

			// Load GPG public key
			gpgKeyData, err := os.ReadFile(gpgKeyPath)
			if err != nil {
				return fmt.Errorf("failed to read GPG key: %w", err)
			}

			// Create forensic data
			forensicData := &watermark.ForensicData{
				Recipient:    recipient,
				PreparedDate: time.Now(),
				GPGPublicKey: string(gpgKeyData),
				DocumentID:   documentID,
				CreatedAt:    time.Now(),
				WatermarkID:  watermark.GenerateWatermarkID(),
			}

			// Validate forensic data
			if err := forensicData.Validate(); err != nil {
				return fmt.Errorf("invalid forensic data: %w", err)
			}

			// Create watermark configuration
			config := &watermark.WatermarkConfig{
				Strategy:   watermark.WatermarkStrategy(strategy),
				Technique:  stego.StegoTechnique(technique),
				Redundancy: redundancy,
				Encrypt:    !noEncrypt,
				Sign:       !noSign,
			}

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Create embed request
			req := &watermark.EmbedRequest{
				ForensicData:  forensicData,
				Config:        config,
				InputDir:      inputDir,
				OutputDir:     outputDir,
				PublicKeyPath: gpgKeyPath,
			}

			// Get watermark service from container
			watermarkService := container.WatermarkService

			logger.Log.Info("Embedding forensic watermark...")

			// Embed watermark
			result, err := watermarkService.Embed(ctx, req)
			if err != nil {
				return fmt.Errorf("watermark embedding failed: %w", err)
			}

			// Display results
			fmt.Printf("\n✅ Watermark Embedded Successfully\n\n")
			fmt.Printf("Watermark ID:       %s\n", result.WatermarkID)
			fmt.Printf("Images Processed:   %d\n", result.ImagesProcessed)
			fmt.Printf("Strategy:           %s\n", result.Strategy)
			fmt.Printf("Total Data Size:    %d bytes\n", result.TotalDataSize)
			fmt.Printf("Distributed Bytes:  %d bytes per image\n", result.DistributedBytes)
			fmt.Printf("Output Directory:   %s\n", outputDir)
			fmt.Printf("\nRecipient:          %s\n", recipient)
			fmt.Printf("Prepared Date:      %s\n", forensicData.PreparedDate.Format(time.RFC3339))
			fmt.Printf("GPG Key:            %s\n", gpgKeyPath)
			if documentID != "" {
				fmt.Printf("Document ID:        %s\n", documentID)
			}

			// Save manifest if available
			if result.Manifest != nil {
				manifestPath := filepath.Join(outputDir, "watermark-manifest.json")
				manifestData, err := json.MarshalIndent(result.Manifest, "", "  ")
				if err != nil {
					logger.Log.WithError(err).Warn("Failed to serialize manifest")
				} else {
					if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
						logger.Log.WithError(err).Warn("Failed to save manifest")
					} else {
						fmt.Printf("\n📄 Manifest saved: %s\n", manifestPath)
					}
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&recipient, "recipient", "r", "", "Recipient identifier (e.g., 'John Doe <john@example.com>') [required]")
	cmd.Flags().StringVarP(&gpgKeyPath, "gpg-key", "k", "", "Path to GPG public signing key [required]")
	cmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Directory containing PNG images [required]")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for watermarked images [required]")
	cmd.Flags().StringVarP(&technique, "technique", "t", "lsb", "Steganography technique: lsb or dct")
	cmd.Flags().Float64Var(&redundancy, "redundancy", 0.3, "Reed-Solomon redundancy (0.0-1.0, e.g., 0.3 = 30%)")
	cmd.Flags().StringVar(&documentID, "document-id", "", "Optional document identifier")
	cmd.Flags().StringVar(&strategy, "strategy", "redundant", "Distribution strategy: distributed, single, or redundant")
	cmd.Flags().BoolVar(&noEncrypt, "no-encrypt", false, "Disable Kyber-1024 encryption")
	cmd.Flags().BoolVar(&noSign, "no-sign", false, "Disable Dilithium3 signatures")

	return cmd
}

// newWatermarkExtractCommand creates the extract subcommand
func newWatermarkExtractCommand(container *services.ServiceContainer) *cobra.Command {
	var (
		inputDir   string
		outputPath string
		technique  string
		threshold  int
		noDecrypt  bool
	)

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract forensic watermark from PNG images",
		Long: `Recovers the forensic watermark from watermarked PNG images using
Reed-Solomon error correction. Can recover from partial image sets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate required parameters
			if inputDir == "" {
				return fmt.Errorf("--input-dir is required")
			}
			if outputPath == "" {
				return fmt.Errorf("--output is required")
			}

			// Create watermark configuration
			config := &watermark.WatermarkConfig{
				Technique: stego.StegoTechnique(technique),
				Encrypt:   !noDecrypt,
			}

			// Create extract request
			req := &watermark.ExtractRequest{
				InputDir:  inputDir,
				Config:    config,
				Threshold: threshold,
			}

			// Get watermark service from container
			watermarkService := container.WatermarkService

			logger.Log.Info("Extracting forensic watermark...")

			// Extract watermark
			result, err := watermarkService.Extract(ctx, req)
			if err != nil {
				return fmt.Errorf("watermark extraction failed: %w", err)
			}

			if !result.Success {
				return fmt.Errorf("insufficient shards: have %d, need %d", result.ShardsRecovered, result.ShardsNeeded)
			}

			// Display results
			fmt.Printf("\n✅ Watermark Extracted Successfully\n\n")
			fmt.Printf("Watermark ID:       %s\n", result.ForensicData.WatermarkID)
			fmt.Printf("Recipient:          %s\n", result.ForensicData.Recipient)
			fmt.Printf("Prepared Date:      %s\n", result.ForensicData.PreparedDate.Format(time.RFC3339))
			fmt.Printf("Created At:         %s\n", result.ForensicData.CreatedAt.Format(time.RFC3339))
			if result.ForensicData.DocumentID != "" {
				fmt.Printf("Document ID:        %s\n", result.ForensicData.DocumentID)
			}
			fmt.Printf("\nShards Recovered:   %d / %d\n", result.ShardsRecovered, result.ShardsNeeded)
			fmt.Printf("Signature Verified: %v\n", result.Verified)

			// Save forensic data to file
			forensicJSON, err := json.MarshalIndent(result.ForensicData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize forensic data: %w", err)
			}

			if err := os.WriteFile(outputPath, forensicJSON, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			fmt.Printf("\n📄 Forensic data saved: %s\n", outputPath)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Directory containing watermarked PNG images [required]")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file for recovered forensic data (JSON) [required]")
	cmd.Flags().StringVarP(&technique, "technique", "t", "lsb", "Steganography technique: lsb or dct")
	cmd.Flags().IntVar(&threshold, "threshold", 10, "Minimum shards needed for recovery")
	cmd.Flags().BoolVar(&noDecrypt, "no-decrypt", false, "Skip Kyber-1024 decryption")

	return cmd
}

// newWatermarkVerifyCommand creates the verify subcommand
func newWatermarkVerifyCommand(container *services.ServiceContainer) *cobra.Command {
	var (
		inputDir          string
		expectedRecipient string
		expectedDate      string
		gpgKeyPath        string
		technique         string
		threshold         int
	)

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify forensic watermark authenticity",
		Long: `Extracts and verifies the forensic watermark matches expected values.
Checks recipient, date, and GPG key authenticity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate required parameters
			if inputDir == "" {
				return fmt.Errorf("--input-dir is required")
			}

			// Create watermark configuration
			config := &watermark.WatermarkConfig{
				Technique: stego.StegoTechnique(technique),
				Encrypt:   true,
				Sign:      true,
			}

			// Create extract request
			req := &watermark.ExtractRequest{
				InputDir:  inputDir,
				Config:    config,
				Threshold: threshold,
			}

			// Get watermark service from container
			watermarkService := container.WatermarkService

			logger.Log.Info("Verifying forensic watermark...")

			// Extract watermark
			result, err := watermarkService.Extract(ctx, req)
			if err != nil {
				return fmt.Errorf("watermark verification failed: %w", err)
			}

			if !result.Success {
				return fmt.Errorf("verification failed: insufficient shards")
			}

			// Verify recipient if provided
			recipientMatch := true
			if expectedRecipient != "" {
				recipientMatch = result.ForensicData.Recipient == expectedRecipient
			}

			// Verify date if provided
			dateMatch := true
			if expectedDate != "" {
				expectedTime, err := time.Parse(time.RFC3339, expectedDate)
				if err != nil {
					return fmt.Errorf("invalid expected date format (use RFC3339): %w", err)
				}
				// Check if dates match within 24 hours
				dateMatch = result.ForensicData.PreparedDate.Sub(expectedTime).Abs() < 24*time.Hour
			}

			// Verify GPG key if provided
			gpgKeyMatch := true
			if gpgKeyPath != "" {
				expectedGPGKey, err := os.ReadFile(gpgKeyPath)
				if err != nil {
					return fmt.Errorf("failed to read expected GPG key: %w", err)
				}
				gpgKeyMatch = result.ForensicData.GPGPublicKey == string(expectedGPGKey)
			}

			// Overall verification result
			verified := recipientMatch && dateMatch && gpgKeyMatch && result.Verified

			// Display results
			if verified {
				fmt.Printf("\n✅ Watermark Verification PASSED\n\n")
			} else {
				fmt.Printf("\n❌ Watermark Verification FAILED\n\n")
			}

			fmt.Printf("Watermark ID:       %s\n", result.ForensicData.WatermarkID)
			fmt.Printf("Recipient:          %s", result.ForensicData.Recipient)
			if expectedRecipient != "" {
				if recipientMatch {
					fmt.Printf(" ✓\n")
				} else {
					fmt.Printf(" ✗ (expected: %s)\n", expectedRecipient)
				}
			} else {
				fmt.Printf("\n")
			}

			fmt.Printf("Prepared Date:      %s", result.ForensicData.PreparedDate.Format(time.RFC3339))
			if expectedDate != "" {
				if dateMatch {
					fmt.Printf(" ✓\n")
				} else {
					fmt.Printf(" ✗ (expected: %s)\n", expectedDate)
				}
			} else {
				fmt.Printf("\n")
			}

			if gpgKeyPath != "" {
				fmt.Printf("GPG Key Match:      ")
				if gpgKeyMatch {
					fmt.Printf("✓\n")
				} else {
					fmt.Printf("✗\n")
				}
			}

			fmt.Printf("Signature Verified: ")
			if result.Verified {
				fmt.Printf("✓\n")
			} else {
				fmt.Printf("✗\n")
			}

			fmt.Printf("\nShards Recovered:   %d / %d\n", result.ShardsRecovered, result.ShardsNeeded)

			if !verified {
				os.Exit(1)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Directory containing watermarked PNG images [required]")
	cmd.Flags().StringVar(&expectedRecipient, "expected-recipient", "", "Expected recipient identifier")
	cmd.Flags().StringVar(&expectedDate, "expected-date", "", "Expected preparation date (RFC3339 format)")
	cmd.Flags().StringVar(&gpgKeyPath, "gpg-key", "", "Path to expected GPG public key")
	cmd.Flags().StringVarP(&technique, "technique", "t", "lsb", "Steganography technique: lsb or dct")
	cmd.Flags().IntVar(&threshold, "threshold", 10, "Minimum shards needed for recovery")

	return cmd
}
