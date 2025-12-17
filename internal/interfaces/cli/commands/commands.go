// Package commands provides all CLI command implementations for Shadowforge
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

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// CLI command handlers - these will be injected with actual service dependencies
type CLIHandlers struct {
	embedHandler           *commands.EmbedHandler
	extractHandler         *commands.ExtractHandler
	analyzeCapacityHandler *commands.AnalyzeCapacityHandler
	logger                 *logrus.Logger
}

// NewCLIHandlers creates CLI command handlers with service dependencies.
// Accepts pre-initialized handlers from the service container.
func NewCLIHandlers(
	embedHandler *commands.EmbedHandler,
	extractHandler *commands.ExtractHandler,
	analyzeCapacityHandler *commands.AnalyzeCapacityHandler,
	logger *logrus.Logger,
) *CLIHandlers {
	return &CLIHandlers{
		embedHandler:           embedHandler,
		extractHandler:         extractHandler,
		analyzeCapacityHandler: analyzeCapacityHandler,
		logger:                 logger,
	}
}

// NewEmbedCommands creates all embed-related commands
func NewEmbedCommands(handlers *CLIHandlers, logger *logrus.Logger) ([]*cobra.Command, error) {

	embedCmd := &cobra.Command{
		Use:   "embed",
		Short: "Embed data using steganography",
		Long: `Embed data into cover media using various steganographic techniques.

Supports multiple techniques:
  • LSB (Least Significant Bit) for PNG/BMP images
  • DCT (Discrete Cosine Transform) for JPEG images
  • Phase encoding for WAV audio files
  • Echo hiding for WAV audio files
  • LSB audio for WAV files
  • Zero-width characters for text files
  • Palette manipulation for GIF/PNG indexed images

Examples:
  # Basic embedding with auto-detected technique
  shadowforge embed -i secret.txt -c cover.png -o stego.png

  # Embedding with specific technique and encryption
  shadowforge embed -i document.pdf -c photo.jpg -o stego.jpg -t dct -p mypassword

  # Embedding with Reed-Solomon error correction
  shadowforge embed -i data.bin -c audio.wav -o stego.wav -r 0.3 -q 90`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.handleEmbedCommand(cmd, args, logger)
		},
	}

	// Add flags for embed command
	embedCmd.Flags().StringP("input", "i", "", "Input file to embed (required)")
	embedCmd.Flags().StringP("cover", "c", "", "Cover media file (required)")
	embedCmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if omitted)")
	embedCmd.Flags().StringP("technique", "t", "", "Steganography technique (auto-detect if empty)")
	embedCmd.Flags().StringP("password", "p", "", "Password for encryption (optional)")
	embedCmd.Flags().Float64P("redundancy", "r", 0.0, "Reed-Solomon redundancy level (0.0-1.0)")
	embedCmd.Flags().IntP("quality", "q", 90, "Quality level (1-100)")
	embedCmd.Flags().BoolP("json", "j", false, "Output result in JSON format")

	// Mark required flags
	embedCmd.MarkFlagRequired("input")
	embedCmd.MarkFlagRequired("cover")
	// output is now optional - will be auto-generated

	return []*cobra.Command{embedCmd}, nil
}

// handleEmbedCommand processes the embed command with all its flags and options.
func (h *CLIHandlers) handleEmbedCommand(cmd *cobra.Command, args []string, logger *logrus.Logger) error {
	// Get flag values
	inputFile, _ := cmd.Flags().GetString("input")
	coverFile, _ := cmd.Flags().GetString("cover")
	outputFile, _ := cmd.Flags().GetString("output")
	technique, _ := cmd.Flags().GetString("technique")
	password, _ := cmd.Flags().GetString("password")
	redundancy, _ := cmd.Flags().GetFloat64("redundancy")
	quality, _ := cmd.Flags().GetInt("quality")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Auto-generate output filename if not specified
	if outputFile == "" {
		outputFile = generateEmbedOutputFilename(coverFile)
		logger.WithField("auto_generated_output", outputFile).Info("Auto-generated output filename")
	}

	// Validate files exist
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}
	if _, err := os.Stat(coverFile); os.IsNotExist(err) {
		return fmt.Errorf("cover file does not exist: %s", coverFile)
	}

	// Convert technique string to enum if specified, otherwise use intelligent selection
	var stegoTechnique stego.StegoTechnique
	if technique != "" {
		var err error
		stegoTechnique, err = stringToTechnique(technique)
		if err != nil {
			return fmt.Errorf("invalid technique: %w", err)
		}
	} else {
		// Get payload size for intelligent selection
		payloadInfo, err := os.Stat(inputFile)
		if err != nil {
			return fmt.Errorf("failed to stat input file: %w", err)
		}
		payloadSize := payloadInfo.Size()

		// Use intelligent technique selection for maximum obfuscation
		stegoTechnique, err = selectOptimalTechnique(coverFile, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to select technique: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"selected_technique": stegoTechnique,
			"payload_size":       payloadSize,
			"cover_file":         coverFile,
		}).Info("Auto-selected optimal technique for maximum obfuscation")
	}

	// Create embed command
	embedCmd := commands.EmbedCommand{
		InputFile:  inputFile,
		CoverFile:  coverFile,
		OutputFile: outputFile,
		Technique:  stegoTechnique,
		Password:   password,
		Redundancy: redundancy,
		Quality:    quality,
	}

	// Use real handler instead of simulation
	ctx := context.Background()
	result, err := h.embedHandler.Handle(ctx, embedCmd)
	if err != nil {
		return fmt.Errorf("embed operation failed: %w", err)
	}

	// Output result
	if jsonOutput {
		return h.outputJSON(result)
	}

	return h.outputEmbedResult(result)
}

// simulateEmbedOperation simulates the embed operation until services are wired
func (h *CLIHandlers) simulateEmbedOperation(cmd commands.EmbedCommand, jsonOutput bool, start time.Time, logger *logrus.Logger) error {
	// Read file sizes for realistic simulation
	inputStat, _ := os.Stat(cmd.InputFile)
	coverStat, _ := os.Stat(cmd.CoverFile)

	// Simulate processing time based on file size
	payloadSize := inputStat.Size()
	coverSize := coverStat.Size()

	// Basic capacity estimation (very rough)
	var estimatedCapacity int64
	var detectedTechnique stego.StegoTechnique = cmd.Technique

	ext := strings.ToLower(filepath.Ext(cmd.CoverFile))
	if cmd.Technique == "" {
		// Auto-detect based on file extension
		switch ext {
		case ".png", ".bmp":
			detectedTechnique = stego.LSB
			estimatedCapacity = coverSize / 8 // Rough LSB capacity
		case ".jpg", ".jpeg":
			detectedTechnique = stego.DCT
			estimatedCapacity = coverSize / 16 // Rough DCT capacity
		case ".wav":
			detectedTechnique = stego.PhaseEncoding
			estimatedCapacity = coverSize / 32 // Rough audio capacity
		case ".txt", ".md":
			detectedTechnique = stego.ZeroWidth
			estimatedCapacity = coverSize / 4 // Rough text capacity
		case ".gif":
			detectedTechnique = stego.Palette
			estimatedCapacity = coverSize / 64 // Rough palette capacity
		default:
			return fmt.Errorf("unsupported cover file format: %s", ext)
		}
	} else {
		estimatedCapacity = coverSize / 10 // Generic estimation
	}

	// Check capacity
	if payloadSize > estimatedCapacity {
		return fmt.Errorf("payload too large: %d bytes, estimated capacity: %d bytes",
			payloadSize, estimatedCapacity)
	}

	// Simulate processing delay
	time.Sleep(time.Duration(payloadSize/1024) * time.Millisecond)

	// Copy cover file to output (simulation)
	coverData, err := os.ReadFile(cmd.CoverFile)
	if err != nil {
		return fmt.Errorf("failed to read cover file: %w", err)
	}

	outputDir := filepath.Dir(cmd.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(cmd.OutputFile, coverData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	processingTime := time.Since(start)
	capacityUsed := float64(payloadSize) / float64(estimatedCapacity) * 100

	// Create simulated result
	result := &commands.EmbedResult{
		OutputFile:      cmd.OutputFile,
		Technique:       detectedTechnique,
		PayloadSize:     payloadSize,
		CoverSize:       coverSize,
		OutputSize:      coverSize, // Same size for simulation
		CapacityUsed:    capacityUsed,
		QualityScore:    0.85, // Simulated quality score
		ProcessingTime:  processingTime.Milliseconds(),
		EncryptionUsed:  cmd.Password != "",
		CompressionUsed: cmd.Redundancy > 0,
	}

	logger.WithFields(logrus.Fields{
		"technique":       result.Technique,
		"capacity_used":   fmt.Sprintf("%.1f%%", result.CapacityUsed),
		"processing_time": fmt.Sprintf("%dms", result.ProcessingTime),
	}).Info("Embed operation simulated successfully")

	// Output result
	if jsonOutput {
		return h.outputJSON(result)
	}

	return h.outputEmbedResult(result)
}

// handleExtractCommand implements the extract command with business logic.
func (h *CLIHandlers) handleExtractCommand(cmd *cobra.Command, args []string) error {
	logger := h.logger.WithField("command", "extract")
	logger.Info("Starting extract operation")

	// Parse flags
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output")
	password, _ := cmd.Flags().GetString("password")
	techniqueStr, _ := cmd.Flags().GetString("technique")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Set logging level
	if verbose {
		logger.Info("Verbose mode enabled")
	}

	// Auto-generate output filename if not specified
	if outputFile == "" {
		outputFile = generateExtractOutputFilename(inputFile)
		logger.WithField("auto_generated_output", outputFile).Info("Auto-generated output filename")
	}

	// Validate required flags
	if inputFile == "" {
		return fmt.Errorf("input file is required (use --input)")
	}

	// Validate input file exists and is readable
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Convert technique string to domain type if specified, otherwise auto-detect
	var technique stego.StegoTechnique
	if techniqueStr != "" {
		var err error
		technique, err = stringToTechnique(techniqueStr)
		if err != nil {
			return fmt.Errorf("invalid technique: %w", err)
		}
	} else {
		// Auto-detect technique based on file type
		var err error
		technique, err = autoDetectEmbeddedTechnique(inputFile)
		if err != nil {
			return fmt.Errorf("failed to auto-detect technique: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"detected_technique": technique,
			"input_file":         inputFile,
		}).Info("Auto-detected embedded technique")
	}

	// Create extract command
	extractCmd := commands.ExtractCommand{
		InputFile:  inputFile,
		OutputFile: outputFile,
		Password:   password,
		Technique:  technique,
	}

	// Validate command
	if err := extractCmd.Validate(); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}

	logger.Info("Validated extract command",
		"input_file", inputFile,
		"output_file", outputFile,
		"password_set", password != "",
		"technique", techniqueStr)

	// Use real handler instead of simulation
	ctx := context.Background()
	result, err := h.extractHandler.Handle(ctx, extractCmd)
	if err != nil {
		return fmt.Errorf("extract operation failed: %w", err)
	}

	// Output results
	if jsonOutput {
		return h.outputJSON(result)
	}
	return h.outputExtractResult(result)
}

// handleAnalyzeCapacityCommand implements the analyze capacity command with business logic.
func (h *CLIHandlers) handleAnalyzeCapacityCommand(cmd *cobra.Command, args []string) error {
	logger := h.logger.WithField("command", "analyze-capacity")
	logger.Info("Starting capacity analysis")

	// Parse arguments and flags
	inputFile := args[0] // guaranteed by cobra.ExactArgs(1)
	technique, _ := cmd.Flags().GetString("technique")
	analyzeAll, _ := cmd.Flags().GetBool("all")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Set logging level
	if verbose {
		logger.Info("Verbose mode enabled")
	}

	// Validate input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Get file info for analysis
	fileInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()
	fileExt := strings.ToLower(filepath.Ext(inputFile))

	logger.Info("Analyzing file",
		"file", inputFile,
		"size", fileSize,
		"extension", fileExt)

	// Determine which technique to analyze
	var analyzeTechnique stego.StegoTechnique
	if technique != "" {
		// Use specified technique
		parsedTechnique, err := stringToTechnique(technique)
		if err != nil {
			return fmt.Errorf("invalid technique: %w", err)
		}
		analyzeTechnique = parsedTechnique
	} else if !analyzeAll {
		// Auto-detect based on file extension
		switch fileExt {
		case ".png", ".bmp":
			analyzeTechnique = stego.LSB
		case ".jpg", ".jpeg":
			analyzeTechnique = stego.DCT
		case ".wav":
			analyzeTechnique = stego.PhaseEncoding // Default to Phase for WAV
		case ".txt", ".md":
			analyzeTechnique = stego.ZeroWidth
		case ".gif":
			analyzeTechnique = stego.Palette
		default:
			return fmt.Errorf("unsupported file format for capacity analysis: %s", fileExt)
		}
	}
	// If analyzeAll is true, we'll run analysis for all supported techniques for the file type

	// Create analyze capacity command
	analyzeCmd := commands.AnalyzeCapacityCommand{
		CoverFile: inputFile,
		Technique: analyzeTechnique, // Empty if analyzeAll
	}

	// Validate command
	if err := analyzeCmd.Validate(); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}

	// Use real handler instead of simulation
	ctx := context.Background()
	result, err := h.analyzeCapacityHandler.Handle(ctx, analyzeCmd)
	if err != nil {
		return fmt.Errorf("capacity analysis failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"techniques_analyzed": len(result.TechniqueResults),
		"file_size":           fileSize,
	}).Info("Capacity analysis completed")

	// Output results
	if jsonOutput {
		return h.outputJSON(result)
	}

	return h.outputCapacityAnalysisResult(result)
}

// NewExtractCommands creates all extract-related commands
func NewExtractCommands(handlers *CLIHandlers, logger *logrus.Logger) ([]*cobra.Command, error) {
	extractCmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract data from steganographic media",
		Long: `Extract hidden data from steganographic media using the appropriate technique.

Automatically detects and uses the correct extraction method for:
  • LSB data in PNG/BMP images
  • DCT data in JPEG images
  • Phase-encoded data in WAV audio
  • Echo-hidden data in WAV audio
  • LSB audio data in WAV files
  • Zero-width encoded text data
  • Palette-encoded data in GIF/PNG images`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.handleExtractCommand(cmd, args)
		},
	}

	// Add flags
	extractCmd.Flags().StringP("input", "i", "", "Input stego media file (required)")
	extractCmd.Flags().StringP("output", "o", "", "Output file for extracted data (auto-generated if omitted)")
	extractCmd.Flags().StringP("password", "p", "", "Decryption password (if encrypted)")
	extractCmd.Flags().StringP("technique", "t", "", "Steganography technique (auto-detect if not specified)")
	extractCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	extractCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	// Mark required flags (output is now optional)
	extractCmd.MarkFlagRequired("input")

	return []*cobra.Command{extractCmd}, nil
}

// NewAnalyzeCommands creates all analyze-related commands
func NewAnalyzeCommands(handlers *CLIHandlers, logger *logrus.Logger) ([]*cobra.Command, error) {
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze media for steganographic properties",
		Long: `Analyze media files for steganographic capacity, detectability, and security properties.

Analysis capabilities:
  • Capacity calculation for different techniques
  • Statistical analysis (Chi-square, entropy)
  • Detectability scoring and risk assessment
  • Quality analysis and recommendations`,
	}

	// Add subcommands for different types of analysis
	capacityCmd := &cobra.Command{
		Use:   "capacity [file]",
		Short: "Calculate steganographic capacity",
		Long:  "Calculate the maximum data capacity for different steganographic techniques",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.handleAnalyzeCapacityCommand(cmd, args)
		},
	}

	// Add flags for capacity analysis
	capacityCmd.Flags().StringP("technique", "t", "", "Specific technique to analyze (lsb, dct, phase, echo, zerowidth, palette)")
	capacityCmd.Flags().BoolP("all", "a", false, "Analyze all applicable techniques")
	capacityCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
	capacityCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	detectCmd := &cobra.Command{
		Use:   "detect [file]",
		Short: "Analyze detectability risk",
		Long:  "Analyze the detectability risk and statistical properties of media",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement detectability analysis
			cmd.Printf("Analyzing detectability of: %s\n", args[0])
			return nil
		},
	}

	analyzeCmd.AddCommand(capacityCmd, detectCmd)
	return []*cobra.Command{analyzeCmd}, nil
}

// NewUtilityCommands creates all utility commands
func NewUtilityCommands(logger *logrus.Logger) ([]*cobra.Command, error) {
	// Key generation command
	keygenCmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate cryptographic key pairs",
		Long: `Generate post-quantum cryptographic key pairs for use with Shadowforge.

Generates both Kyber-1024 (key encapsulation) and Dilithium3 (signatures) key pairs
for quantum-resistant security.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement key generation
			cmd.Println("Key generation not yet implemented")
			return nil
		},
	}

	// Formats command
	formatsCmd := &cobra.Command{
		Use:   "formats",
		Short: "List supported media formats",
		Long:  "Display all supported media formats and their compatible steganographic techniques",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Supported formats:")
			cmd.Println("  Images:")
			cmd.Println("    • PNG (LSB, Palette)")
			cmd.Println("    • BMP (LSB)")
			cmd.Println("    • JPEG (DCT)")
			cmd.Println("    • GIF (Palette)")
			cmd.Println("  Audio:")
			cmd.Println("    • WAV (Phase, Echo, LSB)")
			cmd.Println("  Text:")
			cmd.Println("    • TXT (Zero-width)")
			cmd.Println("    • MD (Zero-width)")
		},
	}

	// Scan directory command
	scanDirCmd := &cobra.Command{
		Use:   "scan-directory",
		Short: "Scan a directory for compatible cover media",
		Long: `Scan a directory for compatible cover media files and calculate their steganographic capacity.

This command analyzes all supported media files in the specified directory and provides:
  • Per-file capacity analysis with recommended technique
  • Combined total capacity if all files are used
  • File compatibility assessment

Supported media types:
  • Images: PNG, BMP, JPEG, GIF
  • Audio: WAV
  • Text: TXT, MD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleScanDirectoryCommand(cmd, args, logger)
		},
	}

	// Add flags for scan-directory
	scanDirCmd.Flags().StringP("dir", "d", "", "Directory to scan (required)")
	scanDirCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
	scanDirCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	scanDirCmd.MarkFlagRequired("dir")

	return []*cobra.Command{keygenCmd, formatsCmd, scanDirCmd}, nil
}

// outputEmbedResult displays embed operation results in human-readable format.
func (h *CLIHandlers) outputEmbedResult(result *commands.EmbedResult) error {
	fmt.Printf("✅ Embed operation completed successfully!\n\n")
	fmt.Printf("📄 Output Details:\n")
	fmt.Printf("   File: %s\n", result.OutputFile)
	fmt.Printf("   Technique: %s\n", result.Technique)
	fmt.Printf("   Size: %s\n", formatBytes(result.OutputSize))

	fmt.Printf("\n📊 Embedding Statistics:\n")
	fmt.Printf("   Payload: %s\n", formatBytes(result.PayloadSize))
	fmt.Printf("   Cover: %s\n", formatBytes(result.CoverSize))
	fmt.Printf("   Capacity Used: %.1f%%\n", result.CapacityUsed)
	fmt.Printf("   Quality Score: %.2f/1.0\n", result.QualityScore)

	fmt.Printf("\n⚙️  Processing Info:\n")
	fmt.Printf("   Time: %dms\n", result.ProcessingTime)
	if result.EncryptionUsed {
		fmt.Printf("   🔒 Encryption: Enabled\n")
	}
	if result.CompressionUsed {
		fmt.Printf("   📦 Error Correction: Enabled\n")
	}

	return nil
}

// outputExtractResult displays extract operation results in human-readable format.
func (h *CLIHandlers) outputExtractResult(result *commands.ExtractResult) error {
	fmt.Printf("✅ Extract operation completed successfully!\n\n")
	fmt.Printf("📄 Output Details:\n")
	fmt.Printf("   File: %s\n", result.OutputFile)
	fmt.Printf("   Technique: %s\n", result.Technique)
	fmt.Printf("   Size: %s\n", formatBytes(result.PayloadSize))

	fmt.Printf("\n📊 Extraction Statistics:\n")
	if result.IntegrityPassed {
		fmt.Printf("   ✅ Integrity: Valid\n")
	} else {
		fmt.Printf("   ❌ Integrity: Invalid\n")
	}

	fmt.Printf("\n⚙️  Processing Info:\n")
	fmt.Printf("   Time: %dms\n", result.ProcessingTime)
	if result.DecryptionUsed {
		fmt.Printf("   🔓 Decryption: Applied\n")
	}
	if result.DecompressionUsed {
		fmt.Printf("   📦 Decompression: Applied\n")
	}

	return nil
}

// outputJSON outputs any result as formatted JSON.
func (h *CLIHandlers) outputJSON(result interface{}) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputCapacityAnalysisResults displays capacity analysis results in human-readable format.
func (h *CLIHandlers) outputCapacityAnalysisResult(result *commands.CapacityAnalysisResult) error {
	fmt.Printf("📊 Capacity Analysis for: %s\n", result.CoverFile)
	fmt.Printf("📄 File Size: %s\n", formatBytes(result.CoverSize))
	fmt.Printf("🎬 Media Type: %s (%s)\n\n", result.MediaType, result.Format)

	// Output technique results
	for i, techResult := range result.TechniqueResults {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("🔧 Technique: %s\n", techResult.Technique)
		if !techResult.Supported {
			fmt.Printf("   ❌ Not supported: %s\n", techResult.ErrorReason)
			continue
		}

		fmt.Printf("   Max Capacity: %s\n", formatBytes(techResult.MaxCapacity))
		fmt.Printf("   Safe Capacity: %s\n", formatBytes(techResult.SafeCapacity))
		fmt.Printf("   Quality Score: %.2f/1.0\n", techResult.QualityScore)
		fmt.Printf("   Detectability Risk: %.1f%%\n", techResult.DetectabilityRisk*100)
		fmt.Printf("   Performance: %.2f/1.0\n", techResult.PerformanceScore)

		// Visual capacity bar
		percentageUsed := float64(techResult.SafeCapacity) / float64(techResult.MaxCapacity) * 100
		barLength := 20
		filledLength := int(float64(barLength) * percentageUsed / 100)
		bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
		fmt.Printf("   Capacity: [%s] %.1f%%\n", bar, percentageUsed)
	}

	// Output recommendations
	if len(result.Recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations:\n")
		for _, rec := range result.Recommendations {
			var emoji string
			switch rec.Type {
			case "best":
				emoji = "🏆"
			case "fastest":
				emoji = "⚡"
			case "safest":
				emoji = "🛡️"
			case "warning":
				emoji = "⚠️"
			default:
				emoji = "ℹ️"
			}
			fmt.Printf("   %s %s (%s) - %s\n",
				emoji, rec.Technique, rec.Type, rec.Reason)
			fmt.Printf("      Max Payload: %s (confidence: %.1f%%)\n",
				formatBytes(rec.MaxPayload), rec.Confidence*100)
		}
	}

	return nil
}

// formatBytes formats byte counts in human-readable format.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MediaFileInfo holds information about a scanned media file
type MediaFileInfo struct {
	Path                 string               `json:"path"`
	Name                 string               `json:"name"`
	Size                 int64                `json:"size"`
	MediaType            string               `json:"media_type"`
	Extension            string               `json:"extension"`
	RecommendedTechnique stego.StegoTechnique `json:"recommended_technique"`
	MaxCapacity          int64                `json:"max_capacity"`
	SafeCapacity         int64                `json:"safe_capacity"`
	QualityScore         float64              `json:"quality_score"`
	StealthScore         float64              `json:"stealth_score"`
}

// ScanDirectoryResult holds the complete directory scan results
type ScanDirectoryResult struct {
	Directory         string          `json:"directory"`
	TotalFiles        int             `json:"total_files"`
	CompatibleFiles   []MediaFileInfo `json:"compatible_files"`
	SkippedFiles      []string        `json:"skipped_files"`
	TotalCapacity     int64           `json:"total_capacity"`
	SafeTotalCapacity int64           `json:"safe_total_capacity"`
	ScanTime          time.Duration   `json:"scan_time_ms"`
}

// handleScanDirectoryCommand scans a directory for compatible media files
func handleScanDirectoryCommand(cmd *cobra.Command, args []string, logger *logrus.Logger) error {
	startTime := time.Now()

	// Parse flags
	directory, _ := cmd.Flags().GetString("dir")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		logger.Info("Verbose mode enabled")
	}

	// Validate directory exists
	dirInfo, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", directory)
	}

	logger.WithField("directory", directory).Info("Scanning directory for compatible media")

	// Read directory contents
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var compatibleFiles []MediaFileInfo
	var skippedFiles []string
	var totalCapacity int64
	var safeTotalCapacity int64

	// Scan each file
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip subdirectories for now (could add --recursive flag later)
			continue
		}

		filePath := filepath.Join(directory, entry.Name())
		fileExt := strings.ToLower(filepath.Ext(entry.Name()))

		// Check if file extension is supported
		technique, mediaType, supported := detectMediaTypeAndTechnique(fileExt)
		if !supported {
			skippedFiles = append(skippedFiles, entry.Name())
			if verbose {
				logger.WithFields(logrus.Fields{
					"file":      entry.Name(),
					"extension": fileExt,
				}).Debug("Skipping unsupported file type")
			}
			continue
		}

		// Get file info
		fileInfo, err := entry.Info()
		if err != nil {
			logger.WithError(err).WithField("file", entry.Name()).Warn("Failed to get file info")
			skippedFiles = append(skippedFiles, entry.Name())
			continue
		}

		// Calculate capacity based on file size and technique
		maxCapacity, safeCapacity := estimateCapacity(fileInfo.Size(), technique)

		// Get technique scores for quality/stealth
		score := getTechniqueScore(technique)

		mediaFile := MediaFileInfo{
			Path:                 filePath,
			Name:                 entry.Name(),
			Size:                 fileInfo.Size(),
			MediaType:            mediaType,
			Extension:            fileExt,
			RecommendedTechnique: technique,
			MaxCapacity:          maxCapacity,
			SafeCapacity:         safeCapacity,
			QualityScore:         float64(score.CapacityScore) / 100.0,
			StealthScore:         float64(score.StealthScore) / 100.0,
		}

		compatibleFiles = append(compatibleFiles, mediaFile)
		totalCapacity += maxCapacity
		safeTotalCapacity += safeCapacity

		if verbose {
			logger.WithFields(logrus.Fields{
				"file":      entry.Name(),
				"type":      mediaType,
				"technique": technique,
				"capacity":  formatBytes(safeCapacity),
			}).Info("Found compatible media file")
		}
	}

	scanTime := time.Since(startTime)

	result := ScanDirectoryResult{
		Directory:         directory,
		TotalFiles:        len(entries),
		CompatibleFiles:   compatibleFiles,
		SkippedFiles:      skippedFiles,
		TotalCapacity:     totalCapacity,
		SafeTotalCapacity: safeTotalCapacity,
		ScanTime:          scanTime,
	}

	// Output results
	if jsonOutput {
		return outputScanDirectoryJSON(result)
	}
	return outputScanDirectoryResult(result)
}

// detectMediaTypeAndTechnique maps file extensions to media types and recommended techniques
func detectMediaTypeAndTechnique(ext string) (stego.StegoTechnique, string, bool) {
	switch ext {
	case ".png", ".bmp":
		return stego.LSB, "Image", true
	case ".jpg", ".jpeg":
		return stego.DCT, "Image", true
	case ".gif":
		return stego.Palette, "Image", true
	case ".wav":
		return stego.PhaseEncoding, "Audio", true
	case ".txt", ".md":
		return stego.ZeroWidth, "Text", true
	default:
		return "", "", false
	}
}

// estimateCapacity calculates max and safe capacity based on file size and technique
func estimateCapacity(fileSize int64, technique stego.StegoTechnique) (maxCapacity int64, safeCapacity int64) {
	switch technique {
	case stego.LSB:
		// LSB: 1 bit per pixel channel (24-bit RGB = 3 bits per pixel)
		// Conservative estimate: fileSize / 4 for max, / 8 for safe
		maxCapacity = fileSize / 4
		safeCapacity = fileSize / 8
	case stego.DCT:
		// DCT (JPEG): ~10-20% of file size typically
		maxCapacity = fileSize / 5
		safeCapacity = fileSize / 10
	case stego.Palette:
		// Palette: Very limited, ~1-5% of file size
		maxCapacity = fileSize / 10
		safeCapacity = fileSize / 20
	case stego.PhaseEncoding:
		// Phase encoding: ~5-15% of audio data
		maxCapacity = fileSize / 7
		safeCapacity = fileSize / 15
	case stego.EchoHiding:
		// Echo hiding: ~3-10% of audio data
		maxCapacity = fileSize / 10
		safeCapacity = fileSize / 20
	// LSB for audio - covered by LSB case above
	case stego.ZeroWidth:
		// Zero-width text: Very limited, depends on word count
		// Rough estimate: 1 byte per 10 bytes of text
		maxCapacity = fileSize / 10
		safeCapacity = fileSize / 20
	default:
		maxCapacity = fileSize / 10
		safeCapacity = fileSize / 20
	}
	return
}

// outputScanDirectoryResult displays scan results in human-readable format
func outputScanDirectoryResult(result ScanDirectoryResult) error {
	fmt.Printf("📁 Directory Scan Results\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")
	fmt.Printf("📂 Directory: %s\n", result.Directory)
	fmt.Printf("📊 Total Files Scanned: %d\n", result.TotalFiles)
	fmt.Printf("✅ Compatible Media Files: %d\n", len(result.CompatibleFiles))
	fmt.Printf("⏭️  Skipped Files: %d\n\n", len(result.SkippedFiles))

	if len(result.CompatibleFiles) > 0 {
		fmt.Printf("📋 Compatible Media Files:\n")
		fmt.Printf("───────────────────────────────────────────────────────────────\n")

		// Table header
		fmt.Printf("%-30s %-10s %-12s %-15s %s\n",
			"File", "Type", "Technique", "Safe Capacity", "Quality")
		fmt.Printf("───────────────────────────────────────────────────────────────\n")

		// Table rows
		for _, file := range result.CompatibleFiles {
			fileName := file.Name
			if len(fileName) > 28 {
				fileName = fileName[:25] + "..."
			}

			fmt.Printf("%-30s %-10s %-12s %-15s %.1f/%.1f\n",
				fileName,
				file.MediaType,
				file.RecommendedTechnique,
				formatBytes(file.SafeCapacity),
				file.QualityScore,
				file.StealthScore,
			)
		}

		fmt.Printf("───────────────────────────────────────────────────────────────\n\n")

		// Summary statistics
		fmt.Printf("💾 Total Available Capacity:\n")
		fmt.Printf("   Maximum: %s\n", formatBytes(result.TotalCapacity))
		fmt.Printf("   Safe:    %s (recommended)\n", formatBytes(result.SafeTotalCapacity))
		fmt.Printf("\n")

		// Usage suggestions
		fmt.Printf("💡 Usage Suggestions:\n")
		fmt.Printf("   • For payloads up to %s: Use any single file\n",
			formatBytes(getLargestSafeCapacity(result.CompatibleFiles)))
		fmt.Printf("   • For larger payloads: Use distributed embedding across multiple files\n")
		fmt.Printf("   • Recommended: Keep capacity usage below 70%% for maximum stealth\n\n")
	}

	if len(result.SkippedFiles) > 0 {
		fmt.Printf("⏭️  Skipped Files (unsupported format):\n")
		for _, fileName := range result.SkippedFiles {
			fmt.Printf("   • %s\n", fileName)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("⏱️  Scan completed in %dms\n", result.ScanTime.Milliseconds())

	return nil
}

// outputScanDirectoryJSON outputs scan results in JSON format
func outputScanDirectoryJSON(result ScanDirectoryResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// getLargestSafeCapacity finds the file with the largest safe capacity
func getLargestSafeCapacity(files []MediaFileInfo) int64 {
	var largest int64
	for _, file := range files {
		if file.SafeCapacity > largest {
			largest = file.SafeCapacity
		}
	}
	return largest
}
