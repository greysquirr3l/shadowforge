package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/selection"
	"github.com/spf13/cobra"
)

// NewSelectionCommands creates all selection-related commands.
func NewSelectionCommands(handlers *CLIHandlers) ([]*cobra.Command, error) {
	return []*cobra.Command{
		newScanCommand(handlers),
		newSelectCommand(handlers),
		newSuggestCommand(handlers),
	}, nil
}

// outputJSONToStdout writes data as JSON to stdout.
func outputJSONToStdout(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func newScanCommand(handlers *CLIHandlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan <directory>",
		Short: "Scan directory for available cover media",
		Long: `Scan a directory for media files and analyze their steganographic capacity.

This command recursively scans directories to identify suitable cover media files
(images, audio, text) and calculates their embedding capacity for each technique.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recursive, _ := cmd.Flags().GetBool("recursive")
			includeHidden, _ := cmd.Flags().GetBool("hidden")
			maxDepth, _ := cmd.Flags().GetInt("max-depth")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			scanCmd := commands.ScanDirectoryCommand{
				Directory:     args[0],
				Recursive:     recursive,
				IncludeHidden: includeHidden,
				MaxDepth:      maxDepth,
			}

			result, err := handlers.scanHandler.Handle(context.Background(), scanCmd)
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if jsonOutput {
				return outputJSONToStdout(result)
			}

			displayScanResults(result, args[0])
			return nil
		},
	}

	cmd.Flags().BoolP("recursive", "r", true, "Scan directories recursively")
	cmd.Flags().Bool("hidden", false, "Include hidden files")
	cmd.Flags().IntP("max-depth", "d", 0, "Maximum directory depth (0 = unlimited)")
	cmd.Flags().BoolP("json", "j", false, "Output results in JSON format")

	return cmd
}

func newSelectCommand(handlers *CLIHandlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select <directory> <payload-size>",
		Short: "Select optimal cover files for payload",
		Long: `Automatically select the best combination of cover files for a given payload size.

This command analyzes available media files and selects an optimal subset that:
  • Provides sufficient capacity for the payload
  • Minimizes detectability risk
  • Maximizes media type diversity
  • Respects capacity utilization targets`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			directory := args[0]
			payloadSize, err := parseSize(args[1])
			if err != nil {
				return fmt.Errorf("invalid payload size: %w", err)
			}

			maxFiles, _ := cmd.Flags().GetInt("max-files")
			maxPerType, _ := cmd.Flags().GetInt("max-per-type")
			targetUtil, _ := cmd.Flags().GetFloat64("target-utilization")
			diversity, _ := cmd.Flags().GetBool("prefer-diversity")
			minimize, _ := cmd.Flags().GetBool("minimize-detect")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// First, scan the directory
			scanCmd := commands.ScanDirectoryCommand{
				Directory: directory,
				Recursive: true,
			}

			scanResult, err := handlers.scanHandler.Handle(context.Background(), scanCmd)
			if err != nil {
				return fmt.Errorf("directory scan failed: %w", err)
			}

			if scanResult.TotalSafeCapacity < payloadSize {
				// Generate suggestions
				suggestCmd := commands.GenerateSuggestionsCommand{
					RequiredCapacity:  payloadSize,
					AvailableCapacity: scanResult.TotalSafeCapacity,
				}

				suggestions, err := handlers.suggestHandler.Handle(context.Background(), suggestCmd)
				if err != nil {
					return fmt.Errorf("suggestion generation failed: %w", err)
				}

				displayInsufficientCapacity(suggestions)
				return fmt.Errorf("insufficient capacity")
			}

			// Select optimal covers
			selectCmd := commands.SelectCoversCommand{
				AvailableFiles:    scanResult.Files,
				PayloadSize:       payloadSize,
				MaxFiles:          maxFiles,
				MaxPerType:        maxPerType,
				TargetUtilization: targetUtil,
				PreferDiversity:   diversity,
				MinimizeDetect:    minimize,
			}

			plan, err := handlers.selectHandler.Handle(context.Background(), selectCmd)
			if err != nil {
				return fmt.Errorf("selection failed: %w", err)
			}

			if jsonOutput {
				return outputJSONToStdout(plan)
			}

			displaySelectionPlan(plan, dryRun)
			return nil
		},
	}

	cmd.Flags().Int("max-files", 0, "Maximum number of files to use (0 = unlimited)")
	cmd.Flags().Int("max-per-type", 0, "Maximum files per media type (0 = unlimited)")
	cmd.Flags().Float64("target-utilization", 0.7, "Target capacity utilization (0.0-1.0)")
	cmd.Flags().Bool("prefer-diversity", true, "Prefer diverse media types")
	cmd.Flags().Bool("minimize-detect", true, "Minimize detectability score")
	cmd.Flags().Bool("dry-run", false, "Show plan without executing")
	cmd.Flags().BoolP("json", "j", false, "Output results in JSON format")

	return cmd
}

func newSuggestCommand(handlers *CLIHandlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggest <required-size> <available-size>",
		Short: "Generate suggestions for insufficient capacity",
		Long: `Generate intelligent suggestions when available capacity is insufficient.

This command analyzes the capacity gap and provides:
  • Recommendations for additional media files
  • Alternative strategies (compression, redundancy reduction)
  • Size/format suggestions with capacity estimates`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			required, err := parseSize(args[0])
			if err != nil {
				return fmt.Errorf("invalid required size: %w", err)
			}

			available, err := parseSize(args[1])
			if err != nil {
				return fmt.Errorf("invalid available size: %w", err)
			}

			suggestCmd := commands.GenerateSuggestionsCommand{
				RequiredCapacity:  required,
				AvailableCapacity: available,
			}

			result, err := handlers.suggestHandler.Handle(context.Background(), suggestCmd)
			if err != nil {
				return fmt.Errorf("suggestion generation failed: %w", err)
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				return outputJSONToStdout(result)
			}

			displayInsufficientCapacity(result)
			return nil
		},
	}

	cmd.Flags().BoolP("json", "j", false, "Output results in JSON format")

	return cmd
}

// Display functions

func displayScanResults(result *selection.AnalysisResult, directory string) {
	fmt.Printf("\n📊 Directory Scan Results: %s\n", directory)
	fmt.Printf("═══════════════════════════════════════════════════════\n\n")

	fmt.Printf("📁 Files Found: %d\n", len(result.Files))
	fmt.Printf("💾 Total Safe Capacity: %s\n", formatBytes(result.TotalSafeCapacity))
	fmt.Printf("📈 Average Detectability: %.2f (lower is better)\n", result.AverageDetectScore)
	fmt.Println()

	fmt.Println("📷 Media Type Breakdown:")
	for mediaType, count := range result.MediaTypeCounts {
		emoji := getMediaTypeEmoji(mediaType)
		fmt.Printf("  %s %-10s: %d files\n", emoji, mediaType, count)
	}
	fmt.Println()

	if len(result.Files) > 0 {
		fmt.Println("🎯 Top 5 Files by Capacity:")
		fmt.Println("───────────────────────────────────────────────────────")
		for i, file := range result.Files {
			if i >= 5 {
				break
			}
			relPath := file.Path
			if strings.HasPrefix(relPath, directory) {
				relPath = strings.TrimPrefix(relPath, directory)
				relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
			}
			fmt.Printf("  %d. %s\n", i+1, filepath.Base(relPath))
			fmt.Printf("     Type: %s | Capacity: %s | Technique: %s\n",
				file.Format, formatBytes(file.SafeCapacity), file.RecommendedTechnique)
		}
	}
}

func displaySelectionPlan(plan *selection.SelectionPlan, dryRun bool) {
	if dryRun {
		fmt.Println("\n🧪 DRY RUN - Selection Plan Preview")
	} else {
		fmt.Println("\n✅ Optimal Cover Selection Plan")
	}
	fmt.Printf("═══════════════════════════════════════════════════════\n\n")

	fmt.Printf("📊 Plan Summary:\n")
	fmt.Printf("  Selected Files: %d\n", len(plan.SelectedFiles))
	fmt.Printf("  Total Capacity: %s\n", formatBytes(plan.TotalCapacity))
	fmt.Printf("  Used Capacity: %s\n", formatBytes(plan.UsedCapacity))
	fmt.Printf("  Utilization: %.1f%%\n", plan.Utilization*100)
	fmt.Printf("  Avg Detectability: %.2f\n", plan.AvgDetectability)
	fmt.Printf("  Diversity Score: %.2f\n", plan.DiversityScore)
	fmt.Println()

	if len(plan.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
		fmt.Println()
	}

	fmt.Println("🗂️  Selected Files:")
	fmt.Println("───────────────────────────────────────────────────────")
	for i, alloc := range plan.SelectedFiles {
		fmt.Printf("  %d. %s\n", i+1, filepath.Base(alloc.File.Path))
		fmt.Printf("     Technique: %s | Allocated: %s (%.1f%% util)\n",
			alloc.Technique, formatBytes(alloc.AllocatedBytes), alloc.UtilizationRatio*100)
	}
}

func displayInsufficientCapacity(result *commands.SuggestionResult) {
	fmt.Println("\n❌ Insufficient Capacity Detected")
	fmt.Printf("═══════════════════════════════════════════════════════\n\n")

	fmt.Printf("📏 Capacity Gap:\n")
	fmt.Printf("  Required: %s\n", formatBytes(result.RequiredCapacity))
	fmt.Printf("  Available: %s\n", formatBytes(result.AvailableCapacity))
	fmt.Printf("  Shortfall: %s\n", formatBytes(result.CapacityGap))
	fmt.Println()

	fmt.Println("💡 Recommended Media to Add:")
	fmt.Println("───────────────────────────────────────────────────────")
	for _, suggestion := range result.MediaSuggestions {
		emoji := getMediaTypeEmoji(suggestion.MediaType)
		fmt.Printf("  %s %s\n", emoji, suggestion.MediaType)
		fmt.Printf("     %s\n", suggestion.Message)
		fmt.Printf("     Capacity: ~%s\n", formatBytes(suggestion.ExpectedCapacity))
		if len(suggestion.Examples) > 0 {
			fmt.Printf("     Sources: %s\n", suggestion.Examples[0])
		}
		fmt.Println()
	}

	if len(result.Alternatives) > 0 {
		fmt.Println("🔀 Alternative Strategies:")
		fmt.Println("───────────────────────────────────────────────────────")
		for _, alternative := range result.Alternatives {
			fmt.Printf("  • %s\n", alternative)
			fmt.Println()
		}
	}
}

func getMediaTypeEmoji(mediaType media.MediaType) string {
	switch mediaType {
	case media.MediaTypeImage:
		return "🖼️ "
	case media.MediaTypeAudio:
		return "🎵"
	case media.MediaTypeText:
		return "📝"
	default:
		return "📄"
	}
}

func parseSize(sizeStr string) (int64, error) {
	// Simple size parsing (supports K, M, G suffixes)
	sizeStr = strings.TrimSpace(sizeStr)
	multiplier := int64(1)

	if len(sizeStr) > 0 {
		suffix := sizeStr[len(sizeStr)-1]
		switch suffix {
		case 'K', 'k':
			multiplier = 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		case 'M', 'm':
			multiplier = 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		case 'G', 'g':
			multiplier = 1024 * 1024 * 1024
			sizeStr = sizeStr[:len(sizeStr)-1]
		}
	}

	var size int64
	_, err := fmt.Sscanf(sizeStr, "%d", &size)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}
