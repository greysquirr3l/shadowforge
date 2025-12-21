// Package commands provides CLI command implementations for Shadowforge
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/internal/domain/watermark"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// NewWatermarkCommands creates the watermark command group
func NewWatermarkCommands(watermarkService *watermark.WatermarkService, log *logrus.Logger) ([]*cobra.Command, error) {
	watermarkCmd := &cobra.Command{
		Use:   "watermark",
		Short: "Forensic watermarking for confidential documents",
		Long: `Forensic watermarking embeds recipient tracking data into KaTeX formula PNGs
using post-quantum cryptography and distributed steganography.

Watermark data includes:
  • Recipient name and email
  • Document preparation date
  • GPG public signing key
  • Optional document ID
  • Unique watermark ID

The watermark is encrypted with Kyber-1024, signed with Dilithium3,
and distributed across multiple images using Reed-Solomon error correction.`,
		Example: `  # Embed watermark in formula PNGs
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
	watermarkCmd.AddCommand(newWatermarkEmbedCommand(watermarkService, log))
	watermarkCmd.AddCommand(newWatermarkExtractCommand(watermarkService, log))
	watermarkCmd.AddCommand(newWatermarkVerifyCommand(watermarkService, log))

	return []*cobra.Command{watermarkCmd}, nil
}

func newWatermarkEmbedCommand(watermarkService *watermark.WatermarkService, log *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Embed forensic watermark in images",
		Long: `Embed forensic watermark data across multiple PNG images.

The watermark includes recipient information, preparation date, and GPG signing key.
Data is encrypted with Kyber-1024, signed with Dilithium3, and distributed using
Reed-Solomon error correction for robustness against partial image loss.`,
		Example: `  shadowforge watermark embed \
    --recipient "Jane Smith <jane@corp.com>" \
    --gpg-key ./keys/jane-key.pub \
    --input-dir ./formulas \
    --output-dir ./watermarked \
    --technique lsb \
    --redundancy 0.30 \
    --document-id "CONF-2025-001"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get flags
			recipient, _ := cmd.Flags().GetString("recipient")
			gpgKeyPath, _ := cmd.Flags().GetString("gpg-key")
			inputDir, _ := cmd.Flags().GetString("input-dir")
			outputDir, _ := cmd.Flags().GetString("output-dir")
			techniqueStr, _ := cmd.Flags().GetString("technique")
			redundancy, _ := cmd.Flags().GetFloat64("redundancy")
			documentID, _ := cmd.Flags().GetString("document-id")
			strategyStr, _ := cmd.Flags().GetString("strategy")
			noEncrypt, _ := cmd.Flags().GetBool("no-encrypt")
			noSign, _ := cmd.Flags().GetBool("no-sign")
			receiptPath, _ := cmd.Flags().GetString("receipt")

			// Validate required fields
			if recipient == "" {
				return fmt.Errorf("recipient is required")
			}
			if gpgKeyPath == "" {
				return fmt.Errorf("gpg-key is required")
			}
			if inputDir == "" {
				return fmt.Errorf("input-dir is required")
			}
			if outputDir == "" {
				return fmt.Errorf("output-dir is required")
			}

			// Parse technique
			var technique stego.StegoTechnique
			switch strings.ToLower(techniqueStr) {
			case "lsb":
				technique = stego.LSB
			case "dct":
				technique = stego.DCT
			default:
				return fmt.Errorf("invalid technique: %s (use 'lsb' or 'dct')", techniqueStr)
			}

			// Parse strategy
			var strategy watermark.WatermarkStrategy
			switch strings.ToLower(strategyStr) {
			case "distributed":
				strategy = watermark.StrategyDistributed
			case "single":
				strategy = watermark.StrategySingle
			case "redundant":
				strategy = watermark.StrategyRedundant
			default:
				return fmt.Errorf("invalid strategy: %s", strategyStr)
			}

			// Load GPG public key (read file content)
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
			}

			// Validate forensic data
			if err := forensicData.Validate(); err != nil {
				return fmt.Errorf("invalid forensic data: %w", err)
			}

			// Create watermark config
			config := &watermark.WatermarkConfig{
				Strategy:   strategy,
				Technique:  technique,
				Redundancy: redundancy,
				Encrypt:    !noEncrypt,
				Sign:       !noSign,
			}

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			logger.Log.WithFields(logrus.Fields{
				"recipient":   recipient,
				"input_dir":   inputDir,
				"output_dir":  outputDir,
				"technique":   techniqueStr,
				"redundancy":  redundancy,
				"strategy":    strategyStr,
				"document_id": documentID,
			}).Info("Embedding watermark")

			// Create embed request
			request := &watermark.EmbedRequest{
				ForensicData:  forensicData,
				Config:        config,
				InputDir:      inputDir,
				OutputDir:     outputDir,
				PublicKeyPath: gpgKeyPath,
				ReceiptPath:   receiptPath,
			}

			// Execute embedding
			result, err := watermarkService.Embed(ctx, request)
			if err != nil {
				return fmt.Errorf("watermark embedding failed: %w", err)
			}

			// Display results
			fmt.Printf("\n✅ Watermark embedded successfully!\n\n")
			fmt.Printf("📋 Watermark Details:\n")
			fmt.Printf("   Watermark ID: %s\n", result.WatermarkID)
			fmt.Printf("   Images Processed: %d\n", result.ImagesProcessed)
			fmt.Printf("   Strategy: %s\n", result.Strategy)
			fmt.Printf("   Total Data Size: %d bytes\n", result.TotalDataSize)
			fmt.Printf("   Distributed Bytes: %d\n", result.DistributedBytes)
			fmt.Printf("\n👤 Forensic Data:\n")
			fmt.Printf("   Recipient: %s\n", forensicData.Recipient)
			fmt.Printf("   Prepared Date: %s\n", forensicData.PreparedDate.Format(time.RFC3339))
			fmt.Printf("   GPG Key: %s\n", gpgKeyPath)
			if forensicData.DocumentID != "" {
				fmt.Printf("   Document ID: %s\n", forensicData.DocumentID)
			}

			// Save manifest
			manifestPath := filepath.Join(outputDir, "watermark-manifest.json")
			manifestData, err := json.MarshalIndent(result.Manifest, "", "  ")
			if err != nil {
				logger.Log.WithError(err).Warn("Failed to marshal manifest")
			} else {
				if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
					logger.Log.WithError(err).Warn("Failed to save manifest")
				} else {
					fmt.Printf("\n📄 Manifest saved to: %s\n", manifestPath)
				}
			}

			return nil
		},
	}

	cmd.Flags().String("recipient", "", "Recipient name and email (required)")
	cmd.Flags().String("gpg-key", "", "Path to GPG public signing key (required)")
	cmd.Flags().String("input-dir", "", "Directory containing PNG images (required)")
	cmd.Flags().String("output-dir", "", "Output directory for watermarked images (required)")
	cmd.Flags().String("technique", "lsb", "Steganography technique (lsb, dct)")
	cmd.Flags().Float64("redundancy", 0.30, "Reed-Solomon redundancy (0.0-1.0)")
	cmd.Flags().String("document-id", "", "Optional document identifier")
	cmd.Flags().String("strategy", "distributed", "Distribution strategy (distributed, single, redundant)")
	cmd.Flags().Bool("no-encrypt", false, "Skip encryption (not recommended)")
	cmd.Flags().Bool("no-sign", false, "Skip digital signature (not recommended)")
	cmd.Flags().String("receipt", "", "Write encryption key receipt (Markdown). Required to later decrypt encrypted watermarks")

	cobra.CheckErr(cmd.MarkFlagRequired("recipient"))
	cobra.CheckErr(cmd.MarkFlagRequired("gpg-key"))
	cobra.CheckErr(cmd.MarkFlagRequired("input-dir"))
	cobra.CheckErr(cmd.MarkFlagRequired("output-dir"))

	return cmd
}

func newWatermarkExtractCommand(watermarkService *watermark.WatermarkService, log *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract forensic watermark from images",
		Long: `Extract and decrypt forensic watermark data from watermarked images.

Recovers the embedded recipient information, preparation date, and GPG signing key.
Uses Reed-Solomon error correction to recover data even if some images are lost or corrupted.`,
		Example: `  shadowforge watermark extract \
    --input-dir ./watermarked \
    --output ./recovered-data.json \
    --technique lsb \
    --threshold 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get flags
			inputDir, _ := cmd.Flags().GetString("input-dir")
			outputPath, _ := cmd.Flags().GetString("output")
			techniqueStr, _ := cmd.Flags().GetString("technique")
			threshold, _ := cmd.Flags().GetInt("threshold")
			noDecrypt, _ := cmd.Flags().GetBool("no-decrypt")
			receiptPath, _ := cmd.Flags().GetString("receipt")

			// Validate required fields
			if inputDir == "" {
				return fmt.Errorf("input-dir is required")
			}
			if outputPath == "" {
				return fmt.Errorf("output is required")
			}

			// Parse technique
			var technique stego.StegoTechnique
			switch strings.ToLower(techniqueStr) {
			case "lsb":
				technique = stego.LSB
			case "dct":
				technique = stego.DCT
			default:
				return fmt.Errorf("invalid technique: %s (use 'lsb' or 'dct')", techniqueStr)
			}

			// Create watermark config
			config := &watermark.WatermarkConfig{
				Technique: technique,
				Encrypt:   !noDecrypt,
			}

			logger.Log.WithFields(logrus.Fields{
				"input_dir": inputDir,
				"output":    outputPath,
				"technique": techniqueStr,
				"threshold": threshold,
			}).Info("Extracting watermark")

			// Create extract request
			request := &watermark.ExtractRequest{
				InputDir:    inputDir,
				Config:      config,
				Threshold:   threshold,
				ReceiptPath: receiptPath,
			}

			// Execute extraction
			result, err := watermarkService.Extract(ctx, request)
			if err != nil {
				return fmt.Errorf("watermark extraction failed: %w", err)
			}

			if !result.Success {
				return fmt.Errorf("failed to recover watermark (recovered %d/%d shards)",
					result.ShardsRecovered, result.ShardsNeeded)
			}

			// Display results
			fmt.Printf("\n✅ Watermark extracted successfully!\n\n")
			fmt.Printf("📋 Watermark Details:\n")
			fmt.Printf("   Watermark ID: %s\n", result.ForensicData.WatermarkID)
			fmt.Printf("   Shards Recovered: %d/%d\n", result.ShardsRecovered, result.ShardsNeeded)
			if result.Verified {
				fmt.Printf("   Signature: ✅ Verified\n")
			} else {
				fmt.Printf("   Signature: ⚠️  Not verified\n")
			}
			fmt.Printf("\n👤 Forensic Data:\n")
			fmt.Printf("   Recipient: %s\n", result.ForensicData.Recipient)
			fmt.Printf("   Prepared Date: %s\n", result.ForensicData.PreparedDate.Format(time.RFC3339))
			fmt.Printf("   Created At: %s\n", result.ForensicData.CreatedAt.Format(time.RFC3339))
			if result.ForensicData.DocumentID != "" {
				fmt.Printf("   Document ID: %s\n", result.ForensicData.DocumentID)
			}

			// Save forensic data to JSON
			forensicJSON, err := json.MarshalIndent(result.ForensicData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal forensic data: %w", err)
			}

			if err := os.WriteFile(outputPath, forensicJSON, 0644); err != nil {
				return fmt.Errorf("failed to save forensic data: %w", err)
			}

			fmt.Printf("\n📄 Forensic data saved to: %s\n", outputPath)

			return nil
		},
	}

	cmd.Flags().String("input-dir", "", "Directory containing watermarked images (required)")
	cmd.Flags().String("output", "", "Output path for recovered data (required)")
	cmd.Flags().String("technique", "lsb", "Steganography technique (lsb, dct)")
	cmd.Flags().Int("threshold", 10, "Minimum shards needed for recovery")
	cmd.Flags().Bool("no-decrypt", false, "Skip decryption")
	cmd.Flags().String("receipt", "", "Path to encryption key receipt (Markdown) for decrypting encrypted watermarks")

	cobra.CheckErr(cmd.MarkFlagRequired("input-dir"))
	cobra.CheckErr(cmd.MarkFlagRequired("output"))

	return cmd
}

func newWatermarkVerifyCommand(watermarkService *watermark.WatermarkService, log *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify forensic watermark authenticity",
		Long: `Verify the authenticity and integrity of a forensic watermark.

Extracts the watermark and verifies:
  • Recipient name matches expected value
  • Preparation date is within tolerance
  • GPG signing key matches
  • Digital signature is valid

Exits with code 1 if verification fails.`,
		Example: `  shadowforge watermark verify \
    --input-dir ./watermarked \
    --expected-recipient "John Doe" \
    --expected-date "2025-01-15" \
    --gpg-key ./keys/john-key.pub \
    --technique lsb`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get flags
			inputDir, _ := cmd.Flags().GetString("input-dir")
			expectedRecipient, _ := cmd.Flags().GetString("expected-recipient")
			expectedDateStr, _ := cmd.Flags().GetString("expected-date")
			gpgKeyPath, _ := cmd.Flags().GetString("gpg-key")
			techniqueStr, _ := cmd.Flags().GetString("technique")
			threshold, _ := cmd.Flags().GetInt("threshold")
			receiptPath, _ := cmd.Flags().GetString("receipt")

			// Validate required fields
			if inputDir == "" {
				return fmt.Errorf("input-dir is required")
			}

			// Parse technique
			var technique stego.StegoTechnique
			switch strings.ToLower(techniqueStr) {
			case "lsb":
				technique = stego.LSB
			case "dct":
				technique = stego.DCT
			default:
				return fmt.Errorf("invalid technique: %s (use 'lsb' or 'dct')", techniqueStr)
			}

			// Create watermark config
			config := &watermark.WatermarkConfig{
				Technique: technique,
				Encrypt:   true,
				Sign:      true,
			}

			logger.Log.WithFields(logrus.Fields{
				"input_dir":          inputDir,
				"expected_recipient": expectedRecipient,
				"technique":          techniqueStr,
			}).Info("Verifying watermark")

			// Create extract request
			request := &watermark.ExtractRequest{
				InputDir:    inputDir,
				Config:      config,
				Threshold:   threshold,
				ReceiptPath: receiptPath,
			}

			// Execute extraction
			result, err := watermarkService.Extract(ctx, request)
			if err != nil {
				fmt.Printf("❌ Verification failed: %v\n", err)
				os.Exit(1)
			}

			if !result.Success {
				fmt.Printf("❌ Failed to recover watermark (%d/%d shards)\n",
					result.ShardsRecovered, result.ShardsNeeded)
				os.Exit(1)
			}

			// Verify recipient
			recipientMatch := true
			if expectedRecipient != "" {
				recipientMatch = strings.Contains(strings.ToLower(result.ForensicData.Recipient),
					strings.ToLower(expectedRecipient))
			}

			// Verify date (within 24 hours tolerance)
			dateMatch := true
			if expectedDateStr != "" {
				expectedDate, err := time.Parse("2006-01-02", expectedDateStr)
				if err != nil {
					fmt.Printf("⚠️  Invalid date format: %s (use YYYY-MM-DD)\n", expectedDateStr)
					dateMatch = false
				} else {
					dateMatch = result.ForensicData.PreparedDate.Sub(expectedDate).Abs() < 24*time.Hour
				}
			}

			// Verify GPG key
			gpgKeyMatch := true
			if gpgKeyPath != "" {
				expectedGPGKey, err := os.ReadFile(gpgKeyPath)
				if err != nil {
					fmt.Printf("⚠️  Failed to read GPG key: %v\n", err)
					gpgKeyMatch = false
				} else {
					gpgKeyMatch = result.ForensicData.GPGPublicKey == string(expectedGPGKey)
				}
			}

			// Overall verification
			verified := result.Verified && recipientMatch && dateMatch && gpgKeyMatch

			// Display results
			fmt.Printf("\n🔍 Watermark Verification Results:\n\n")
			fmt.Printf("📋 Watermark ID: %s\n\n", result.ForensicData.WatermarkID)

			if result.Verified {
				fmt.Printf("✅ Signature: Verified\n")
			} else {
				fmt.Printf("❌ Signature: Invalid or missing\n")
			}

			if recipientMatch {
				fmt.Printf("✅ Recipient: %s\n", result.ForensicData.Recipient)
			} else {
				fmt.Printf("❌ Recipient: %s (expected: %s)\n",
					result.ForensicData.Recipient, expectedRecipient)
			}

			if dateMatch {
				fmt.Printf("✅ Prepared Date: %s\n", result.ForensicData.PreparedDate.Format(time.RFC3339))
			} else {
				fmt.Printf("❌ Prepared Date: %s (expected: %s)\n",
					result.ForensicData.PreparedDate.Format("2006-01-02"), expectedDateStr)
			}

			if gpgKeyMatch {
				fmt.Printf("✅ GPG Key: Matches\n")
			} else {
				fmt.Printf("❌ GPG Key: Does not match\n")
			}

			fmt.Printf("\n")
			if verified {
				fmt.Printf("✅ Overall Verification: PASSED\n")
				return nil
			} else {
				fmt.Printf("❌ Overall Verification: FAILED\n")
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().String("input-dir", "", "Directory containing watermarked images (required)")
	cmd.Flags().String("expected-recipient", "", "Expected recipient name/email")
	cmd.Flags().String("expected-date", "", "Expected preparation date (YYYY-MM-DD)")
	cmd.Flags().String("gpg-key", "", "Path to expected GPG public key")
	cmd.Flags().String("technique", "lsb", "Steganography technique (lsb, dct)")
	cmd.Flags().Int("threshold", 10, "Minimum shards needed for recovery")
	cmd.Flags().String("receipt", "", "Path to encryption key receipt (Markdown) for decrypting encrypted watermarks")

	cobra.CheckErr(cmd.MarkFlagRequired("input-dir"))

	return cmd
}
