// Package root provides the root command and command tree for the Shadowforge CLI
package root

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/greysquirr3l/shadowforge/internal/interfaces/cli/commands"
	"github.com/greysquirr3l/shadowforge/internal/interfaces/cli/services"
	"github.com/greysquirr3l/shadowforge/pkg/version"
)

const (
	appName = "shadowforge"
	appDesc = "Quantum-resistant steganography tool - Forge secrets in the shadows, shield them from quantum eyes"
)

// NewRootCommand creates and configures the root command with all subcommands
func NewRootCommand(logger *logrus.Logger) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: "Shadowforge - Quantum-resistant steganography",
		Long: fmt.Sprintf(`%s

%s is a production-grade quantum-resistant steganography tool that combines:
• Post-Quantum Cryptography (Kyber-1024, Dilithium3)
• Reed-Solomon Error Correction for data resilience
• Multiple Steganography Techniques (LSB, DCT, Audio Phase, Echo, Text, Palette)
• Flexible Distribution Patterns (1:1, 1:N, N:1, N:M)

Available aliases: shadowforge, sforge`, appDesc, appName),
		Version:      version.GetVersion(),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set up global flags and configuration here
			return nil
		},
	}

	// Add global persistent flags
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().String("config", "", "Config file path (default: $HOME/.shadowforge.yaml)")

	// Initialize service container with all dependencies
	container, err := services.NewServiceContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service container: %w", err)
	}

	// Create CLI handlers with injected services
	handlers := commands.NewCLIHandlers(
		container.EmbedHandler,
		container.ExtractHandler,
		container.AnalyzeCapacityHandler,
		logger,
	)

	// Create command groups with handlers
	embedCommands, err := commands.NewEmbedCommands(handlers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create embed commands: %w", err)
	}

	extractCommands, err := commands.NewExtractCommands(handlers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create extract commands: %w", err)
	}

	analyzeCommands, err := commands.NewAnalyzeCommands(handlers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyze commands: %w", err)
	}

	utilityCommands, err := commands.NewUtilityCommands(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create utility commands: %w", err)
	}

	// Add all command groups
	rootCmd.AddCommand(embedCommands...)
	rootCmd.AddCommand(extractCommands...)
	rootCmd.AddCommand(analyzeCommands...)
	rootCmd.AddCommand(utilityCommands...)

	// Add version command
	rootCmd.AddCommand(newVersionCommand())

	return rootCmd, nil
}

// newVersionCommand creates the version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display detailed version information about Shadowforge",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", appName, version.GetVersion())
			fmt.Printf("Go version: %s\n", version.GetBuildInfo().GoVersion)
			fmt.Printf("Git commit: %s\n", version.GetGitCommit())
			fmt.Printf("Build date: %s\n", version.GetBuildTime())
			fmt.Println()
			fmt.Println("Supported techniques:")
			fmt.Println("  • LSB Image (PNG, BMP)")
			fmt.Println("  • DCT JPEG")
			fmt.Println("  • Audio Phase (WAV)")
			fmt.Println("  • Audio Echo (WAV)")
			fmt.Println("  • Audio LSB (WAV)")
			fmt.Println("  • Text Zero-Width (TXT, MD)")
			fmt.Println("  • Palette (GIF, PNG indexed)")
			fmt.Println()
			fmt.Println("Security:")
			fmt.Println("  • Post-Quantum Cryptography: Kyber-1024, Dilithium3")
			fmt.Println("  • Reed-Solomon Error Correction")
			fmt.Println("  • Secure memory handling")
		},
	}
}
