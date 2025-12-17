// Package e2e provides end-to-end tests for Shadowforge CLI with real files
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	binaryName  = "shadowforge"
	testDataDir = "../testdata"
)

var projectRoot string

func init() {
	projectRoot = findProjectRoot()
}

// TestE2E_VersionCommand tests the version command with compiled binary
func TestE2E_VersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	output, err := execCommand(binaryPath, "version")
	require.NoError(t, err, "version command should succeed")
	assert.Contains(t, output, "shadowforge version", "output should contain version info")
}

// TestE2E_HelpCommand tests the help command
func TestE2E_HelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	output, err := execCommand(binaryPath, "help")
	require.NoError(t, err, "help command should succeed")
	assert.Contains(t, output, "shadowforge", "help should contain command name")
	assert.Contains(t, output, "embed", "help should list embed command")
	assert.Contains(t, output, "extract", "help should list extract command")
}

// TestE2E_FormatsCommand tests the formats listing command
func TestE2E_FormatsCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	output, err := execCommand(binaryPath, "formats")
	require.NoError(t, err, "formats command should succeed")
	assert.Contains(t, output, "LSB", "should list LSB technique")
	assert.Contains(t, output, "DCT", "should list DCT technique")
}

// TestE2E_EmbedExtractWorkflow_LSB tests complete embed/extract workflow with LSB technique
func TestE2E_EmbedExtractWorkflow_LSB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	tmpDir := t.TempDir()

	// Create test files
	secretFile := filepath.Join(tmpDir, "secret.txt")
	coverFile := filepath.Join(tmpDir, "cover.png")
	stegoFile := filepath.Join(tmpDir, "stego.png")
	extractedFile := filepath.Join(tmpDir, "extracted.txt")

	secretData := "This is a secret message for E2E testing"
	require.NoError(t, os.WriteFile(secretFile, []byte(secretData), 0644))

	// Copy real PNG from mixed-media directory
	sourcePNG := filepath.Join(projectRoot, "mixed-media", "images", "f88dv18h9o1f1.png")
	require.NoError(t, copyFile(sourcePNG, coverFile), "should copy test PNG")

	// Step 1: Embed
	output, err := execCommand(binaryPath, "embed", "-i", secretFile, "-c", coverFile, "-o", stegoFile, "-t", "lsb")
	require.NoError(t, err, "embed command should succeed. Output: %s", output)
	assert.FileExists(t, stegoFile, "stego file should be created")

	// Step 2: Extract (must specify technique - auto-detection not implemented)
	output, err = execCommand(binaryPath, "extract", "-i", stegoFile, "-o", extractedFile, "-t", "lsb")
	require.NoError(t, err, "extract command should succeed. Output: %s", output)
	assert.FileExists(t, extractedFile, "extracted file should be created")

	// Step 3: Verify
	extractedData, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	assert.Equal(t, secretData, string(extractedData), "extracted data should match original")
}

// TestE2E_AnalyzeCapacity tests the capacity analysis command
func TestE2E_AnalyzeCapacity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	tmpDir := t.TempDir()
	coverFile := filepath.Join(tmpDir, "cover.png")

	// Copy real PNG from mixed-media directory
	sourcePNG := filepath.Join(projectRoot, "mixed-media", "images", "f88dv18h9o1f1.png")
	require.NoError(t, copyFile(sourcePNG, coverFile))

	// analyze capacity takes file as positional argument
	output, err := execCommand(binaryPath, "analyze", "capacity", coverFile, "-t", "lsb")
	require.NoError(t, err, "analyze command should succeed")
	assert.Contains(t, output, "bytes", "output should contain capacity information")
}

// TestE2E_JSONOutput tests JSON output mode
func TestE2E_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	output, err := execCommand(binaryPath, "formats")
	require.NoError(t, err, "formats command should succeed")
	// Just verify it outputs technique names
	assert.Contains(t, output, "LSB", "output should list techniques")
	assert.Contains(t, output, "DCT", "output should list techniques")
}

// TestE2E_ErrorScenarios tests various error conditions
func TestE2E_ErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := findOrBuildBinary(t)
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "missing_input_file",
			args:        []string{"embed", "-i", "nonexistent.txt", "-c", filepath.Join(tmpDir, "cover.png"), "-o", filepath.Join(tmpDir, "out.png")},
			shouldError: true,
		},
		{
			name:        "missing_cover_file",
			args:        []string{"embed", "-i", filepath.Join(tmpDir, "secret.txt"), "-c", "nonexistent.png", "-o", filepath.Join(tmpDir, "out.png")},
			shouldError: true,
		},
		{
			name:        "invalid_technique",
			args:        []string{"embed", "-t", "invalid-technique"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := execCommand(binaryPath, tt.args...)
			if tt.shouldError {
				assert.Error(t, err, "command should fail for %s", tt.name)
			} else {
				assert.NoError(t, err, "command should succeed for %s", tt.name)
			}
		})
	}
}

// Helper Functions

// findOrBuildBinary finds existing binary or builds it
func findOrBuildBinary(t *testing.T) string {
	binaryPath := filepath.Join(projectRoot, "bin", binaryName)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Build binary
	t.Logf("Building CLI binary...")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cli")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build binary: %s", string(output))

	return binaryPath
}

// execCommand executes binary with args and returns output
func execCommand(binaryPath string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// findProjectRoot finds the project root directory
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}

	// Walk up until we find go.mod
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find project root (go.mod)")
		}
		dir = parent
	}
}

// fileExists checks if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFile copies a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// Note: createMinimalPNG removed - now using real PNG files from mixed-media/
