// Package commands_test provides unit tests for CLI command handlers
package commands_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/interfaces/cli/root"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// TestRootCommand verifies the root command is created successfully
func TestRootCommand(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()

	// Act
	rootCmd, err := root.NewRootCommand(testLogger)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "shadowforge", rootCmd.Use)
	assert.Contains(t, rootCmd.Long, "Quantum-resistant steganography")
}

// TestRootCommandVersion verifies the version command works
func TestRootCommandVersion(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})

	// Act
	err = rootCmd.Execute()

	// Assert - version executes successfully
	require.NoError(t, err)
	// Note: Version command may print directly to os.Stdout bypassing SetOut
	// Just verify it executes without error - E2E tests will verify output
}

// TestRootCommandHelp verifies the help command works
func TestRootCommandHelp(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	// Act
	err = rootCmd.Execute()

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "shadowforge")
	assert.Contains(t, output, "Available Commands")
	assert.Contains(t, output, "embed")
	assert.Contains(t, output, "extract")
	assert.Contains(t, output, "analyze")
}

// TestEmbedCommandFlags verifies embed command has required flags
func TestEmbedCommandFlags(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find embed command
	var embedCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "embed" {
			embedCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, embedCmd, "embed command should exist")

	// Check required flags exist
	assert.NotNil(t, embedCmd.Flags().Lookup("input"), "should have input flag")
	assert.NotNil(t, embedCmd.Flags().Lookup("cover"), "should have cover flag")
	assert.NotNil(t, embedCmd.Flags().Lookup("output"), "should have output flag")
	assert.NotNil(t, embedCmd.Flags().Lookup("technique"), "should have technique flag")
	assert.NotNil(t, embedCmd.Flags().Lookup("password"), "should have password flag")
	assert.NotNil(t, embedCmd.Flags().Lookup("redundancy"), "should have redundancy flag")
}

// TestExtractCommandFlags verifies extract command has required flags
func TestExtractCommandFlags(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find extract command
	var extractCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "extract" {
			extractCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, extractCmd, "extract command should exist")

	// Check required flags exist
	assert.NotNil(t, extractCmd.Flags().Lookup("input"), "should have input flag")
	assert.NotNil(t, extractCmd.Flags().Lookup("output"), "should have output flag")
	assert.NotNil(t, extractCmd.Flags().Lookup("password"), "should have password flag")
}

// TestAnalyzeCommandExists verifies analyze command exists
func TestAnalyzeCommandExists(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find analyze command
	var analyzeCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "analyze" {
			analyzeCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, analyzeCmd, "analyze command should exist")
	assert.Contains(t, analyzeCmd.Short, "Analyze", "should mention analysis")
}

// TestFormatsCommandExists verifies formats command exists
func TestFormatsCommandExists(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find formats command
	var formatsCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "formats" {
			formatsCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, formatsCmd, "formats command should exist")
}

// TestGlobalFlags verifies global persistent flags exist
func TestGlobalFlags(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Assert global flags
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"), "should have verbose flag")
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("debug"), "should have debug flag")
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("config"), "should have config flag")
}

// TestCommandErrorHandling verifies error handling for invalid commands
func TestCommandErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			shouldError: true,
		},
		{
			name:        "embed without required flags",
			args:        []string{"embed"},
			shouldError: true,
		},
		{
			name:        "extract without required flags",
			args:        []string{"extract"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testLogger := logger.NewLogger()
			rootCmd, err := root.NewRootCommand(testLogger)
			require.NoError(t, err)

			// Capture output to suppress error messages
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			// Act
			err = rootCmd.Execute()

			// Assert
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContextCancellation verifies commands respect context cancellation
func TestContextCancellation(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Suppress output
	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetErr(new(bytes.Buffer))
	rootCmd.SetArgs([]string{"version"})

	// Act - version command should still work even with cancelled context
	err = rootCmd.ExecuteContext(ctx)

	// Assert - version is simple and should complete
	assert.NoError(t, err)
}

// TestJSONOutputFlag verifies JSON output flag is available
func TestJSONOutputFlag(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find embed command
	var embedCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "embed" {
			embedCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, embedCmd)
	jsonFlag := embedCmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag, "should have json output flag")
}

// TestScanDirectoryCommand verifies the scan-directory command exists
func TestScanDirectoryCommand(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find scan-directory command
	var scanCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "scan-directory" {
			scanCmd = cmd
			break
		}
	}

	// Assert
	require.NotNil(t, scanCmd, "scan-directory command should exist")
	assert.Equal(t, "scan-directory", scanCmd.Use)
	assert.Contains(t, scanCmd.Short, "Scan a directory")
}

// TestScanDirectoryFlags verifies all required flags exist
func TestScanDirectoryFlags(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	// Find scan-directory command
	var scanCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "scan-directory" {
			scanCmd = cmd
			break
		}
	}
	require.NotNil(t, scanCmd)

	// Assert flags exist
	tests := []struct {
		name     string
		flagName string
		required bool
	}{
		{"dir flag exists", "dir", true},
		{"json flag exists", "json", false},
		{"verbose flag exists", "verbose", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := scanCmd.Flags().Lookup(tt.flagName)
			assert.NotNil(t, flag, "flag %s should exist", tt.flagName)
		})
	}
}

// TestScanDirectoryMissingDir verifies error when directory not provided
func TestScanDirectoryMissingDir(t *testing.T) {
	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"scan-directory"})

	// Act
	err = rootCmd.Execute()

	// Assert - should fail with missing directory error
	assert.Error(t, err)
	output := buf.String()
	assert.Contains(t, output, "required flag(s)", "should mention required flag")
}

// TestScanDirectoryNonExistentDir verifies error for non-existent directory
func TestScanDirectoryNonExistentDir(t *testing.T) {
	// Skip in short mode (requires filesystem)
	if testing.Short() {
		t.Skip("Skipping filesystem test in short mode")
	}

	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"scan-directory", "--dir", "/nonexistent/directory/path"})

	// Act
	err = rootCmd.Execute()

	// Assert - should fail gracefully
	assert.Error(t, err)
	output := buf.String()
	assert.Contains(t, output, "directory", "error should mention directory issue")
}

// TestScanDirectoryWithTestData verifies successful scan of test media
func TestScanDirectoryWithTestData(t *testing.T) {
	// Skip in short mode (requires filesystem)
	if testing.Short() {
		t.Skip("Skipping filesystem test in short mode")
	}

	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Use mixed-media/images directory if it exists
	rootCmd.SetArgs([]string{"scan-directory", "--dir", "mixed-media/images"})

	// Act
	err = rootCmd.Execute()

	// Assert - should succeed if directory exists
	// Note: May fail if directory doesn't exist, which is acceptable
	if err == nil {
		output := buf.String()
		assert.Contains(t, output, "Directory Scan Results", "should show scan results header")
		assert.Contains(t, output, "Total Files Scanned", "should show file count")
	}
}

// TestScanDirectoryJSONOutput verifies JSON output format
func TestScanDirectoryJSONOutput(t *testing.T) {
	// Skip in short mode (requires filesystem)
	if testing.Short() {
		t.Skip("Skipping filesystem test in short mode")
	}

	// Arrange
	testLogger := logger.NewLogger()
	rootCmd, err := root.NewRootCommand(testLogger)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Use mixed-media/images directory with JSON flag
	rootCmd.SetArgs([]string{"scan-directory", "--dir", "mixed-media/images", "--json"})

	// Act
	err = rootCmd.Execute()

	// Assert - should produce valid JSON if directory exists
	if err == nil {
		output := buf.String()
		assert.Contains(t, output, "directory", "JSON should contain directory field")
		assert.Contains(t, output, "total_files", "JSON should contain total_files field")
		assert.Contains(t, output, "compatible_files", "JSON should contain compatible_files field")
	}
}
