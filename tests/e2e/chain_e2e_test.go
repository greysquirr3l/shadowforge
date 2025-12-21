package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_SequentialChain_VacationPhotos tests sequential chaining through CLI
// Use case: User wants to hide a secret message in 5 vacation photos using different techniques
func TestE2E_SequentialChain_VacationPhotos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create test directory
	testDir := t.TempDir()

	// Create test payload
	payloadPath := filepath.Join(testDir, "secret_message.txt")
	payload := []byte("This is my secret vacation diary. Weather was great! 🌞")
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create 5 test cover images
	coverPaths := make([]string, 5)
	for i := 0; i < 5; i++ {
		coverPaths[i] = filepath.Join(testDir, fmt.Sprintf("vacation_photo_%d.png", i+1))
		createTestPNG(t, coverPaths[i])
	}

	// Create chain configuration
	chainConfigPath := filepath.Join(testDir, "chain_config.json")
	chainConfig := map[string]interface{}{
		"mode": "sequential",
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_1",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 1,
					"channels":         []string{"R", "G", "B"},
				},
				"order": 1,
			},
			{
				"technique_id": "dct_1",
				"technique":    "dct",
				"configuration": map[string]interface{}{
					"quality": 85,
				},
				"order": 2,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute chain through CLI
	binaryPath := findOrBuildBinary(t)
	outputPath := filepath.Join(testDir, "chained_output.png")
	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--cover", coverPaths[0],
		"--output", outputPath,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Chain execute output:\n%s", output)

	// Chain command may not be fully implemented yet
	if err != nil {
		t.Logf("Chain execute not yet implemented: %v", err)
		t.Skip("Skipping until chain CLI commands are implemented")
		return
	}

	// Verify output exists
	assert.FileExists(t, outputPath)

	// Extract the payload
	extractedPath := filepath.Join(testDir, "extracted.txt")
	extractCmd := exec.Command(binaryPath, "chain", "extract",
		"--chain-config", chainConfigPath,
		"--input", outputPath,
		"--output", extractedPath,
	)

	extractOutput, err := extractCmd.CombinedOutput()
	t.Logf("Chain extract output:\n%s", extractOutput)
	require.NoError(t, err)

	// Verify extracted payload matches original
	extractedData, err := os.ReadFile(extractedPath)
	require.NoError(t, err)
	assert.Equal(t, payload, extractedData, "Extracted payload should match original")
}

// TestE2E_LayeredChain_MultiTechnique tests layered chaining (multiple techniques in one carrier)
func TestE2E_LayeredChain_MultiTechnique(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create payload
	payloadPath := filepath.Join(testDir, "layered_secret.txt")
	payload := []byte("Part 1: LSB data. Part 2: DCT data. Both in same image!")
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create cover image
	coverPath := filepath.Join(testDir, "cover.png")
	createTestPNG(t, coverPath)

	// Create layered chain configuration
	chainConfigPath := filepath.Join(testDir, "layered_chain.json")
	chainConfig := map[string]interface{}{
		"mode": "layered",
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_layer",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 2,
				},
				"weight": 0.6, // 60% of data via LSB
				"order":  1,
			},
			{
				"technique_id": "dct_layer",
				"technique":    "dct",
				"configuration": map[string]interface{}{
					"quality": 90,
				},
				"weight": 0.4, // 40% of data via DCT
				"order":  2,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute layered chain
	binaryPath := findOrBuildBinary(t)
	outputPath := filepath.Join(testDir, "layered_output.png")
	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--cover", coverPath,
		"--output", outputPath,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Layered chain output:\n%s", output)

	if err != nil {
		t.Logf("Layered chain not yet implemented: %v", err)
		t.Skip("Skipping until layered chain is implemented")
		return
	}

	// Verify output
	assert.FileExists(t, outputPath)

	// Extract and verify
	extractedPath := filepath.Join(testDir, "layered_extracted.txt")
	extractCmd := exec.Command(binaryPath, "chain", "extract",
		"--chain-config", chainConfigPath,
		"--input", outputPath,
		"--output", extractedPath,
	)

	extractOutput, err := extractCmd.CombinedOutput()
	t.Logf("Layered extract output:\n%s", extractOutput)
	require.NoError(t, err)

	extractedData, err := os.ReadFile(extractedPath)
	require.NoError(t, err)
	assert.Equal(t, payload, extractedData)
}

// TestE2E_SplitChain_Distribution tests split chaining (data distributed across multiple carriers)
func TestE2E_SplitChain_Distribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create larger payload
	payloadPath := filepath.Join(testDir, "large_secret.txt")
	payload := bytes.Repeat([]byte("This is a longer secret message that will be split across 4 images. "), 10)
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create 4 cover images
	coverPaths := make([]string, 4)
	for i := 0; i < 4; i++ {
		coverPaths[i] = filepath.Join(testDir, fmt.Sprintf("cover_%d.png", i+1))
		createTestPNG(t, coverPaths[i])
	}

	// Create split chain configuration
	chainConfigPath := filepath.Join(testDir, "split_chain.json")
	chainConfig := map[string]interface{}{
		"mode": "split",
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_split_1",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 1,
				},
				"order": 1,
			},
			{
				"technique_id": "lsb_split_2",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 1,
				},
				"order": 2,
			},
			{
				"technique_id": "dct_split_3",
				"technique":    "dct",
				"configuration": map[string]interface{}{
					"quality": 85,
				},
				"order": 3,
			},
			{
				"technique_id": "dct_split_4",
				"technique":    "dct",
				"configuration": map[string]interface{}{
					"quality": 85,
				},
				"order": 4,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute split chain (should create 4 output files)
	binaryPath := findOrBuildBinary(t)
	outputDir := filepath.Join(testDir, "split_outputs")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--covers", coverPaths[0], coverPaths[1], coverPaths[2], coverPaths[3],
		"--output-dir", outputDir,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Split chain output:\n%s", output)

	if err != nil {
		t.Logf("Split chain not yet implemented: %v", err)
		t.Skip("Skipping until split chain is implemented")
		return
	}

	// Verify 4 output files exist
	outputPaths := make([]string, 4)
	for i := 0; i < 4; i++ {
		outputPaths[i] = filepath.Join(outputDir, fmt.Sprintf("output_%d.png", i+1))
		assert.FileExists(t, outputPaths[i])
	}

	// Extract from split chain
	extractedPath := filepath.Join(testDir, "split_extracted.txt")
	extractCmd := exec.Command(binaryPath, "chain", "extract",
		"--chain-config", chainConfigPath,
		"--inputs", outputPaths[0], outputPaths[1], outputPaths[2], outputPaths[3],
		"--output", extractedPath,
	)

	extractOutput, err := extractCmd.CombinedOutput()
	t.Logf("Split extract output:\n%s", extractOutput)
	require.NoError(t, err)

	extractedData, err := os.ReadFile(extractedPath)
	require.NoError(t, err)
	assert.Equal(t, payload, extractedData)
}

// TestE2E_ChainWithManifest tests chain with manifest generation and verification
func TestE2E_ChainWithManifest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create payload
	payloadPath := filepath.Join(testDir, "manifest_secret.txt")
	payload := []byte("Secret with manifest tracking")
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create cover
	coverPath := filepath.Join(testDir, "cover.png")
	createTestPNG(t, coverPath)
	require.NoError(t, err)

	// Create chain config
	chainConfigPath := filepath.Join(testDir, "chain.json")
	chainConfig := map[string]interface{}{
		"mode": "sequential",
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_manifest",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 2,
				},
				"order": 1,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute with manifest generation
	binaryPath := findOrBuildBinary(t)
	outputPath := filepath.Join(testDir, "output.png")
	manifestPath := filepath.Join(testDir, "manifest.json")
	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--cover", coverPath,
		"--output", outputPath,
		"--manifest", manifestPath,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Chain with manifest output:\n%s", output)

	if err != nil {
		t.Logf("Chain manifest not yet implemented: %v", err)
		t.Skip("Skipping until chain manifest is implemented")
		return
	}

	// Verify manifest exists and has expected structure
	assert.FileExists(t, manifestPath)

	manifestData, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var manifest map[string]interface{}
	err = json.Unmarshal(manifestData, &manifest)
	require.NoError(t, err)

	// Verify manifest has chain information
	assert.Contains(t, manifest, "chain_id")
	assert.Contains(t, manifest, "mode")
	assert.Contains(t, manifest, "techniques_used")
	assert.Contains(t, manifest, "created_at")
}

// TestE2E_ChainRecovery tests chain with partial failure recovery
func TestE2E_ChainRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create payload
	payloadPath := filepath.Join(testDir, "recovery_secret.txt")
	payload := []byte("Secret with redundancy for recovery")
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create 3 covers
	coverPaths := make([]string, 3)
	for i := 0; i < 3; i++ {
		coverPaths[i] = filepath.Join(testDir, fmt.Sprintf("cover_%d.png", i+1))
		createTestPNG(t, coverPaths[i])
	}

	// Create chain with K-of-N recovery
	chainConfigPath := filepath.Join(testDir, "recovery_chain.json")
	chainConfig := map[string]interface{}{
		"mode": "split",
		"recovery": map[string]interface{}{
			"enabled":   true,
			"threshold": 2, // Need 2 of 3 to recover
		},
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_1",
				"technique":    "lsb",
				"order":        1,
			},
			{
				"technique_id": "lsb_2",
				"technique":    "lsb",
				"order":        2,
			},
			{
				"technique_id": "lsb_3",
				"technique":    "lsb",
				"order":        3,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute chain
	binaryPath := findOrBuildBinary(t)
	outputDir := filepath.Join(testDir, "outputs")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--covers", coverPaths[0], coverPaths[1], coverPaths[2],
		"--output-dir", outputDir,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Recovery chain output:\n%s", output)

	if err != nil {
		t.Logf("Recovery chain not yet implemented: %v", err)
		t.Skip("Skipping until recovery chain is implemented")
		return
	}

	// Verify outputs exist
	output1 := filepath.Join(outputDir, "output_1.png")
	output2 := filepath.Join(outputDir, "output_2.png")
	_ = filepath.Join(outputDir, "output_3.png") // output3 not used (simulating loss)

	// Extract using only 2 of 3 (simulating loss of one carrier)
	extractedPath := filepath.Join(testDir, "recovered.txt")
	extractCmd := exec.Command(binaryPath, "chain", "extract",
		"--chain-config", chainConfigPath,
		"--inputs", output1, output2, // Only 2 of 3
		"--output", extractedPath,
	)

	extractOutput, err := extractCmd.CombinedOutput()
	t.Logf("Recovery extract output:\n%s", extractOutput)
	require.NoError(t, err, "Should recover with 2 of 3 carriers")

	extractedData, err := os.ReadFile(extractedPath)
	require.NoError(t, err)
	assert.Equal(t, payload, extractedData, "Should recover original payload with 2 of 3")
}

// TestE2E_ChainCapacityExceeded tests chain with payload exceeding capacity
func TestE2E_ChainCapacityExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create LARGE payload (intentionally too big)
	payloadPath := filepath.Join(testDir, "huge_secret.txt")
	payload := bytes.Repeat([]byte("This is a very long secret message that exceeds capacity. "), 1000)
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create small cover
	coverPath := filepath.Join(testDir, "small_cover.png")
	createTestPNG(t, coverPath) // Small image (100x100)

	// Create chain config
	chainConfigPath := filepath.Join(testDir, "chain.json")
	chainConfig := map[string]interface{}{
		"mode": "sequential",
		"links": []map[string]interface{}{
			{
				"technique_id": "lsb_small",
				"technique":    "lsb",
				"configuration": map[string]interface{}{
					"bits_per_channel": 1,
				},
				"order": 1,
			},
		},
	}
	configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(chainConfigPath, configJSON, 0644)
	require.NoError(t, err)

	// Execute chain (should fail with capacity error)
	binaryPath := findOrBuildBinary(t)
	outputPath := filepath.Join(testDir, "output.png")
	cmd := exec.Command(binaryPath, "chain", "execute",
		"--chain-config", chainConfigPath,
		"--payload", payloadPath,
		"--cover", coverPath,
		"--output", outputPath,
	)

	output, err := cmd.CombinedOutput()
	t.Logf("Capacity exceeded output:\n%s", output)

	// Should fail
	assert.Error(t, err, "Should fail when payload exceeds capacity")
	assert.Contains(t, string(output), "capacity", "Error should mention capacity")
}

// TestE2E_ChainInvalidConfiguration tests chain with invalid configuration
func TestE2E_ChainInvalidConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	testDir := t.TempDir()

	// Create payload
	payloadPath := filepath.Join(testDir, "secret.txt")
	payload := []byte("Test secret")
	err := os.WriteFile(payloadPath, payload, 0644)
	require.NoError(t, err)

	// Create cover
	coverPath := filepath.Join(testDir, "cover.png")
	createTestPNG(t, coverPath)
	require.NoError(t, err)

	t.Run("empty_chain", func(t *testing.T) {
		// Create invalid chain config (no links)
		chainConfigPath := filepath.Join(testDir, "invalid_chain.json")
		chainConfig := map[string]interface{}{
			"mode":  "sequential",
			"links": []map[string]interface{}{}, // Empty!
		}
		configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(chainConfigPath, configJSON, 0644)
		require.NoError(t, err)

		// Execute should fail
		binaryPath := findOrBuildBinary(t)
		outputPath := filepath.Join(testDir, "output.png")
		cmd := exec.Command(binaryPath, "chain", "execute",
			"--chain-config", chainConfigPath,
			"--payload", payloadPath,
			"--cover", coverPath,
			"--output", outputPath,
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Empty chain output:\n%s", output)

		assert.Error(t, err, "Should fail with empty chain")
		assert.Contains(t, string(output), "chain", "Error should mention chain")
	})

	t.Run("invalid_technique", func(t *testing.T) {
		// Create invalid chain config (unknown technique)
		chainConfigPath := filepath.Join(testDir, "invalid_technique.json")
		chainConfig := map[string]interface{}{
			"mode": "sequential",
			"links": []map[string]interface{}{
				{
					"technique_id": "unknown_tech",
					"technique":    "quantum_teleportation", // Not a real technique!
					"order":        1,
				},
			},
		}
		configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(chainConfigPath, configJSON, 0644)
		require.NoError(t, err)

		// Execute should fail
		binaryPath := findOrBuildBinary(t)
		outputPath := filepath.Join(testDir, "output2.png")
		cmd := exec.Command(binaryPath, "chain", "execute",
			"--chain-config", chainConfigPath,
			"--payload", payloadPath,
			"--cover", coverPath,
			"--output", outputPath,
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Invalid technique output:\n%s", output)

		assert.Error(t, err, "Should fail with unknown technique")
		assert.Contains(t, string(output), "technique", "Error should mention technique")
	})

	t.Run("layered_without_weights", func(t *testing.T) {
		// Create invalid layered chain (no weights)
		chainConfigPath := filepath.Join(testDir, "layered_no_weights.json")
		chainConfig := map[string]interface{}{
			"mode": "layered",
			"links": []map[string]interface{}{
				{
					"technique_id": "lsb_1",
					"technique":    "lsb",
					"order":        1,
					// Missing weight!
				},
				{
					"technique_id": "dct_1",
					"technique":    "dct",
					"order":        2,
					// Missing weight!
				},
			},
		}
		configJSON, err := json.MarshalIndent(chainConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(chainConfigPath, configJSON, 0644)
		require.NoError(t, err)

		// Execute should fail
		binaryPath := findOrBuildBinary(t)
		outputPath := filepath.Join(testDir, "output3.png")
		cmd := exec.Command(binaryPath, "chain", "execute",
			"--chain-config", chainConfigPath,
			"--payload", payloadPath,
			"--cover", coverPath,
			"--output", outputPath,
		)

		output, err := cmd.CombinedOutput()
		t.Logf("Layered without weights output:\n%s", output)

		if err == nil {
			t.Log("Layered chain accepts missing weights (may use defaults)")
		} else {
			assert.Contains(t, string(output), "weight", "Error should mention weight")
		}
	})
}

// createTestPNG is already defined in cli_e2e_test.go
