package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/internal/domain/watermark"
	infra_crypto "github.com/greysquirr3l/shadowforge/internal/infrastructure/crypto"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/distribution"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	stego_impl "github.com/greysquirr3l/shadowforge/internal/infrastructure/stego_impl"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

func TestWatermarkEmbedExtract_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Initialize services
	log := logger.NewLogger()
	cryptoService := infra_crypto.NewCirclCryptoService(log)
	ecService := errorcorrection.NewRSService(log)
	mediaService := media.NewMediaService(log)
	stegoService := stego_impl.NewStegoService(mediaService, log)
	distService := distribution.NewDistributionService(ecService, mediaService, stegoService, nil, log)

	watermarkService := watermark.NewWatermarkService(
		cryptoService,
		ecService,
		stegoService,
		mediaService,
		distService,
	)

	// Setup test directories
	testDir := t.TempDir()
	inputDir := filepath.Join(testDir, "input")
	outputDir := filepath.Join(testDir, "output")

	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Copy test PNGs to input directory
	testPNGsDir := "tests/watermark/equations"
	if _, err := os.Stat(testPNGsDir); os.IsNotExist(err) {
		t.Skip("Test PNGs not found, skipping integration test")
	}

	copyTestPNGs(t, testPNGsDir, inputDir)

	// Create forensic data
	forensicData := &watermark.ForensicData{
		Recipient:    "Dr. Alice Smith <alice@university.edu>",
		PreparedDate: time.Now(),
		GPGPublicKey: "test-gpg-public-key",
		DocumentID:   "DOC-2025-CONFIDENTIAL-001",
		CreatedAt:    time.Now(),
		WatermarkID:  watermark.GenerateWatermarkID(),
	}

	// Create watermark configuration
	config := &watermark.WatermarkConfig{
		Strategy:   watermark.StrategyRedundant,
		Technique:  "lsb",
		Redundancy: 0.3,
		Encrypt:    true,
		Sign:       true,
	}

	// Embed watermark
	receiptPath := filepath.Join(testDir, "watermark_receipt.md")
	embedReq := &watermark.EmbedRequest{
		ForensicData: forensicData,
		Config:       config,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		ReceiptPath:  receiptPath,
	}

	embedResult, err := watermarkService.Embed(ctx, embedReq)
	require.NoError(t, err)
	assert.NotNil(t, embedResult)
	assert.Equal(t, forensicData.WatermarkID, embedResult.WatermarkID)
	assert.Greater(t, embedResult.ImagesProcessed, 0)

	t.Logf("Embedded watermark in %d images", embedResult.ImagesProcessed)

	// Extract watermark
	extractReq := &watermark.ExtractRequest{
		InputDir:    outputDir,
		Config:      config,
		Threshold:   10,
		ReceiptPath: receiptPath,
	}

	extractResult, err := watermarkService.Extract(ctx, extractReq)
	require.NoError(t, err)
	assert.NotNil(t, extractResult)
	assert.True(t, extractResult.Success)
	assert.True(t, extractResult.Verified)

	// Verify extracted data matches original
	assert.Equal(t, forensicData.WatermarkID, extractResult.ForensicData.WatermarkID)
	assert.Equal(t, forensicData.Recipient, extractResult.ForensicData.Recipient)
	assert.Equal(t, forensicData.DocumentID, extractResult.ForensicData.DocumentID)
	assert.Equal(t, forensicData.GPGPublicKey, extractResult.ForensicData.GPGPublicKey)

	t.Logf("Extracted watermark from %d / %d shards", extractResult.ShardsRecovered, extractResult.ShardsNeeded)
}

func TestWatermarkPartialRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Initialize services
	log := logger.NewLogger()
	cryptoService := infra_crypto.NewCirclCryptoService(log)
	ecService := errorcorrection.NewRSService(log)
	mediaService := media.NewMediaService(log)
	stegoService := stego_impl.NewStegoService(mediaService, log)
	distService := distribution.NewDistributionService(ecService, mediaService, stegoService, nil, log)

	watermarkService := watermark.NewWatermarkService(
		cryptoService,
		ecService,
		stegoService,
		mediaService,
		distService,
	)

	// Setup test directories
	testDir := t.TempDir()
	inputDir := filepath.Join(testDir, "input")
	outputDir := filepath.Join(testDir, "output")
	partialDir := filepath.Join(testDir, "partial")

	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))
	require.NoError(t, os.MkdirAll(partialDir, 0755))

	// Copy test PNGs
	testPNGsDir := "tests/watermark/equations"
	if _, err := os.Stat(testPNGsDir); os.IsNotExist(err) {
		t.Skip("Test PNGs not found, skipping integration test")
	}

	copyTestPNGs(t, testPNGsDir, inputDir)

	// Create forensic data
	forensicData := &watermark.ForensicData{
		Recipient:    "Bob Johnson <bob@corp.com>",
		PreparedDate: time.Now(),
		GPGPublicKey: "test-gpg-key",
		DocumentID:   "DOC-PARTIAL-TEST",
		CreatedAt:    time.Now(),
		WatermarkID:  watermark.GenerateWatermarkID(),
	}

	config := &watermark.WatermarkConfig{
		Strategy:   watermark.StrategyRedundant,
		Technique:  "lsb",
		Redundancy: 0.3,
		Encrypt:    true,
		Sign:       false,
	}

	// Embed watermark
	receiptPath := filepath.Join(testDir, "watermark_receipt.md")
	embedReq := &watermark.EmbedRequest{
		ForensicData: forensicData,
		Config:       config,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		ReceiptPath:  receiptPath,
	}

	_, err := watermarkService.Embed(ctx, embedReq)
	require.NoError(t, err)

	// Copy only 70% of images (simulating 30% loss)
	files, err := filepath.Glob(filepath.Join(outputDir, "*.png"))
	require.NoError(t, err)

	keepCount := int(float64(len(files)) * 0.7)
	for i := 0; i < keepCount; i++ {
		src := files[i]
		dst := filepath.Join(partialDir, filepath.Base(src))
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dst, data, 0644))
	}

	t.Logf("Simulated 30%% image loss: keeping %d / %d images", keepCount, len(files))

	// Extract from partial set
	extractReq := &watermark.ExtractRequest{
		InputDir:    partialDir,
		Config:      config,
		Threshold:   int(float64(keepCount) * 0.8), // Need 80% of remaining
		ReceiptPath: receiptPath,
	}

	extractResult, err := watermarkService.Extract(ctx, extractReq)
	require.NoError(t, err)
	assert.True(t, extractResult.Success)

	// Verify data integrity despite loss
	assert.Equal(t, forensicData.WatermarkID, extractResult.ForensicData.WatermarkID)
	assert.Equal(t, forensicData.Recipient, extractResult.ForensicData.Recipient)

	t.Logf("Successfully recovered watermark from %d / %d shards", extractResult.ShardsRecovered, extractResult.ShardsNeeded)
}

func TestWatermarkTechniques(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	techniques := []string{"lsb", "dct"}

	for _, tech := range techniques {
		t.Run(tech, func(t *testing.T) {
			ctx := context.Background()

			// Initialize services
			log := logger.NewLogger()
			cryptoService := infra_crypto.NewCirclCryptoService(log)
			ecService := errorcorrection.NewRSService(log)
			mediaService := media.NewMediaService(log)
			stegoService := stego_impl.NewStegoService(mediaService, log)
			distService := distribution.NewDistributionService(ecService, mediaService, stegoService, nil, log)

			watermarkService := watermark.NewWatermarkService(
				cryptoService,
				ecService,
				stegoService,
				mediaService,
				distService,
			)

			// Setup directories
			testDir := t.TempDir()
			inputDir := filepath.Join(testDir, "input")
			outputDir := filepath.Join(testDir, "output")

			require.NoError(t, os.MkdirAll(inputDir, 0755))
			require.NoError(t, os.MkdirAll(outputDir, 0755))

			// Copy test PNGs
			testPNGsDir := "tests/watermark/equations"
			if _, err := os.Stat(testPNGsDir); os.IsNotExist(err) {
				t.Skip("Test PNGs not found")
			}

			copyTestPNGs(t, testPNGsDir, inputDir)

			// Forensic data
			forensicData := &watermark.ForensicData{
				Recipient:    "Test User",
				PreparedDate: time.Now(),
				GPGPublicKey: "test-key",
				WatermarkID:  watermark.GenerateWatermarkID(),
				CreatedAt:    time.Now(),
			}

			config := &watermark.WatermarkConfig{
				Strategy:   watermark.StrategyRedundant,
				Technique:  stego.StegoTechnique(tech),
				Redundancy: 0.3,
				Encrypt:    true,
				Sign:       false,
			}

			// Embed
			receiptPath := filepath.Join(testDir, "watermark_receipt.md")
			embedReq := &watermark.EmbedRequest{
				ForensicData: forensicData,
				Config:       config,
				InputDir:     inputDir,
				OutputDir:    outputDir,
				ReceiptPath:  receiptPath,
			}

			embedResult, err := watermarkService.Embed(ctx, embedReq)
			require.NoError(t, err)
			assert.Greater(t, embedResult.ImagesProcessed, 0)

			// Extract
			extractReq := &watermark.ExtractRequest{
				InputDir:    outputDir,
				Config:      config,
				Threshold:   10,
				ReceiptPath: receiptPath,
			}

			extractResult, err := watermarkService.Extract(ctx, extractReq)
			require.NoError(t, err)
			assert.True(t, extractResult.Success)
			assert.Equal(t, forensicData.WatermarkID, extractResult.ForensicData.WatermarkID)

			t.Logf("Technique %s: Embedded in %d images, recovered from %d shards",
				tech, embedResult.ImagesProcessed, extractResult.ShardsRecovered)
		})
	}
}

func TestForensicDataSerialization(t *testing.T) {
	forensicData := &watermark.ForensicData{
		Recipient:    "Charlie Davis <charlie@example.com>",
		PreparedDate: time.Now(),
		GPGPublicKey: "-----BEGIN PGP PUBLIC KEY BLOCK-----\ntest key\n-----END PGP PUBLIC KEY BLOCK-----",
		DocumentID:   "DOC-123",
		CreatedAt:    time.Now(),
		WatermarkID:  watermark.GenerateWatermarkID(),
	}

	// Serialize
	data, err := json.Marshal(forensicData)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	var recovered watermark.ForensicData
	err = json.Unmarshal(data, &recovered)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, forensicData.Recipient, recovered.Recipient)
	assert.Equal(t, forensicData.GPGPublicKey, recovered.GPGPublicKey)
	assert.Equal(t, forensicData.DocumentID, recovered.DocumentID)
	assert.Equal(t, forensicData.WatermarkID, recovered.WatermarkID)
}

// Helper functions

func copyTestPNGs(t *testing.T, srcDir, dstDir string) {
	files, err := filepath.Glob(filepath.Join(srcDir, "*.png"))
	require.NoError(t, err)

	for _, src := range files {
		data, err := os.ReadFile(src)
		require.NoError(t, err)

		dst := filepath.Join(dstDir, filepath.Base(src))
		require.NoError(t, os.WriteFile(dst, data, 0644))
	}

	t.Logf("Copied %d test PNGs from %s to %s", len(files), srcDir, dstDir)
}
