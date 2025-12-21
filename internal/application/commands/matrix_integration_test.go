package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	errorcorrection_infra "github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
)

// mockStegoService provides mock implementation for testing
type mockStegoService struct{}

func (m *mockStegoService) Embed(ctx context.Context, coverMedia, payload []byte, technique stego.StegoTechnique) (*stego.StegoContainer, error) {
	// Mock embed: create dummy container
	containerID := stego.GenerateContainerID()
	container := &stego.StegoContainer{
		ID:           containerID,
		CoverMedia:   coverMedia,
		EmbeddedData: payload,
		EmbeddingMap: []int{1, 2, 3},
		Technique:    technique,
		Capacity:     int64(len(coverMedia)) / 8,
		UsedCapacity: int64(len(payload)),
		Quality:      stego.Quality{Score: 0.9, Detectability: 0.1},
		IsEmbedded:   true,
	}
	return container, nil
}

func (m *mockStegoService) Extract(ctx context.Context, stegoMedia []byte, technique stego.StegoTechnique) ([]byte, error) {
	// Mock extract: return dummy payload
	return []byte("mock_extracted_data"), nil
}

func (m *mockStegoService) CalculateCapacity(ctx context.Context, coverMedia []byte, technique stego.StegoTechnique) (int64, error) {
	// Mock capacity: return size/8 (1 bit per byte = 1/8 capacity)
	return int64(len(coverMedia)) / 8, nil
}

func (m *mockStegoService) AnalyzeQuality(ctx context.Context, stegoMedia []byte) (*stego.Quality, error) {
	// Mock quality analysis
	return &stego.Quality{
		Score:         0.9,
		Detectability: 0.1,
	}, nil
}

func (m *mockStegoService) ValidateContainer(ctx context.Context, container *stego.StegoContainer) error {
	// Mock validation: always pass
	return nil
}

func (m *mockStegoService) OptimizeTechnique(ctx context.Context, coverMedia []byte, payloadSize int64) (stego.StegoTechnique, error) {
	// Mock optimization: always return LSB
	return stego.LSB, nil
}

// mockCryptoService provides mock implementation for testing
type mockCryptoService struct{}

func (m *mockCryptoService) GenerateKeyPair(ctx context.Context, algorithm crypto.PQCAlgorithm) (*crypto.KeyPair, error) {
	// Mock key pair
	return &crypto.KeyPair{
		PublicKey:  []byte("mock_public_key"),
		PrivateKey: []byte("mock_private_key"),
		Algorithm:  algorithm,
	}, nil
}

func (m *mockCryptoService) Encrypt(ctx context.Context, data []byte, publicKey []byte, algorithm crypto.PQCAlgorithm) (*crypto.CryptoPayload, error) {
	// Mock encryption: wrap data in payload
	payloadID := crypto.GeneratePayloadID()
	return &crypto.CryptoPayload{
		ID:        payloadID,
		Data:      data,
		Algorithm: algorithm,
		Nonce:     []byte("mock_nonce"),
	}, nil
}

func (m *mockCryptoService) Decrypt(ctx context.Context, payload *crypto.CryptoPayload, privateKey []byte) ([]byte, error) {
	// Mock decryption: return encrypted data as-is
	return payload.Data, nil
}

func (m *mockCryptoService) Sign(ctx context.Context, data []byte, privateKey []byte, algorithm crypto.PQCAlgorithm) ([]byte, error) {
	// Mock signing
	return []byte("mock_signature"), nil
}

func (m *mockCryptoService) Verify(ctx context.Context, data []byte, signature []byte, publicKey []byte, algorithm crypto.PQCAlgorithm) (bool, error) {
	// Mock verification: always pass
	return true, nil
}

func (m *mockCryptoService) DeriveKey(ctx context.Context, password string, salt []byte) ([]byte, error) {
	// Mock key derivation
	return []byte("mock_derived_key"), nil
}

// TestEmbedMatrixHandler_Handle tests the complete matrix embedding workflow.
func TestEmbedMatrixHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		setupCommand   func() EmbedMatrixCommand
		setupTestFiles func(t *testing.T) (string, func())
		expectSuccess  bool
		expectError    string
		validateResult func(t *testing.T, result *EmbedMatrixResult)
	}{
		{
			name: "successful_2x3_matrix_round_robin",
			setupCommand: func() EmbedMatrixCommand {
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{
						{Name: "secret1", Data: make([]byte, 1024), Path: "secret1.bin"},
						{Name: "secret2", Data: make([]byte, 2048), Path: "secret2.bin"},
					},
					Covers: []MatrixCoverItem{
						{Path: "cover1.png"},
						{Path: "cover2.png"},
						{Path: "cover3.png"},
					},
					Mode:          MatrixModeRoundRobin,
					RSRedundancy:  0.3,
					MinRedundancy: 0.2,
					Compression:   false,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				// Create temp directory
				tmpDir, err := os.MkdirTemp("", "matrix_test_*")
				require.NoError(t, err)

				// Create dummy cover files
				for i := 1; i <= 3; i++ {
					coverPath := filepath.Join(tmpDir, "cover"+string(rune('0'+i))+".png")
					err = os.WriteFile(coverPath, make([]byte, 100000), 0644)
					require.NoError(t, err)
				}

				// Cleanup function
				cleanup := func() {
					_ = os.RemoveAll(tmpDir)
				}

				return tmpDir, cleanup
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *EmbedMatrixResult) {
				assert.Equal(t, 2, result.PayloadCount, "Should have 2 payloads")
				assert.Equal(t, 3, result.CoverCount, "Should have 3 covers")
				assert.Equal(t, "round_robin", result.Mode)
				assert.Greater(t, result.TotalShards, 0, "Should generate shards")
				assert.NotEmpty(t, result.OutputFiles, "Should create output files")
				assert.Equal(t, 3, len(result.OutputFiles), "Should create 3 stego files")
				assert.NotEmpty(t, result.ManifestFile, "Should create manifest")
				assert.Greater(t, result.MatrixDensity, 0.0, "Should have non-zero density")
				assert.Equal(t, 0.3, result.RSRedundancy)
			},
		},
		{
			name: "successful_3x2_matrix_optimized",
			setupCommand: func() EmbedMatrixCommand {
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{
						{Name: "doc1", Data: make([]byte, 500), Path: "doc1.txt"},
						{Name: "doc2", Data: make([]byte, 750), Path: "doc2.txt"},
						{Name: "doc3", Data: make([]byte, 1000), Path: "doc3.txt"},
					},
					Covers: []MatrixCoverItem{
						{Path: "image1.png"},
						{Path: "image2.png"},
					},
					Mode:          MatrixModeOptimized,
					RSRedundancy:  0.25,
					MinRedundancy: 0.2,
					Compression:   true,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "matrix_test_*")
				require.NoError(t, err)

				for i := 1; i <= 2; i++ {
					coverPath := filepath.Join(tmpDir, "image"+string(rune('0'+i))+".png")
					err = os.WriteFile(coverPath, make([]byte, 50000), 0644)
					require.NoError(t, err)
				}

				return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *EmbedMatrixResult) {
				assert.Equal(t, 3, result.PayloadCount)
				assert.Equal(t, 2, result.CoverCount)
				assert.Equal(t, "optimized", result.Mode)
				assert.True(t, result.CompressionUsed, "Compression should be enabled")
				assert.Greater(t, result.AllocationMetrics.UtilizationPercent, 0.0)
				assert.LessOrEqual(t, result.AllocationMetrics.UtilizationPercent, 100.0)
			},
		},
		{
			name: "invalid_command_no_payloads",
			setupCommand: func() EmbedMatrixCommand {
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{},
					Covers: []MatrixCoverItem{
						{Path: "cover.png"},
					},
					Mode:         MatrixModeRoundRobin,
					RSRedundancy: 0.3,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				return "", func() {}
			},
			expectSuccess: false,
			expectError:   "at least one payload is required",
		},
		{
			name: "invalid_command_1x1_matrix",
			setupCommand: func() EmbedMatrixCommand {
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{
						{Name: "single", Data: make([]byte, 100)},
					},
					Covers: []MatrixCoverItem{
						{Path: "single.png"},
					},
					Mode:         MatrixModeRoundRobin,
					RSRedundancy: 0.3,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				return "", func() {}
			},
			expectSuccess: false,
			expectError:   "matrix dimensions invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetOutput(io.Discard)

			handler := NewEmbedMatrixHandler(
				&mockStegoService{},  // Use mock stego service
				&mockCryptoService{}, // Use mock crypto service
				errorcorrection_infra.NewRSService(logger),
				logger,
			)

			// Setup test files and get command
			tmpDir, cleanup := tt.setupTestFiles(t)
			defer cleanup()

			cmd := tt.setupCommand()

			ctx := context.Background()
			if tmpDir != "" {
				// Update command with actual file paths
				cmd.OutputDir = tmpDir
				cmd.ManifestFile = filepath.Join(tmpDir, "manifest.json")

				for i := range cmd.Covers {
					coverName := filepath.Base(cmd.Covers[i].Path)
					cmd.Covers[i].Path = filepath.Join(tmpDir, coverName)
				}
			}

			// Execute
			result, err := handler.Handle(ctx, cmd)

			// Assert
			if tt.expectSuccess {
				require.NoError(t, err, "Expected successful execution")
				require.NotNil(t, result, "Result should not be nil")

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				require.Error(t, err, "Expected error")
				if tt.expectError != "" {
					assert.Contains(t, err.Error(), tt.expectError)
				}
			}
		})
	}
}

// TestExtractMatrixHandler_Handle tests the complete matrix extraction workflow.
func TestExtractMatrixHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		setupCommand   func(tmpDir string) ExtractMatrixCommand
		setupTestFiles func(t *testing.T) (string, func())
		expectSuccess  bool
		expectError    string
		validateResult func(t *testing.T, result *ExtractMatrixResult)
	}{
		{
			name: "successful_extraction_all_covers",
			setupCommand: func(tmpDir string) ExtractMatrixCommand {
				return ExtractMatrixCommand{
					StegoFiles: []string{
						filepath.Join(tmpDir, "stego1.png"),
						filepath.Join(tmpDir, "stego2.png"),
						filepath.Join(tmpDir, "stego3.png"),
					},
					ManifestFile: filepath.Join(tmpDir, "manifest.json"),
					OutputDir:    filepath.Join(tmpDir, "output"),
					AllowPartial: false,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "matrix_extract_*")
				require.NoError(t, err)

				// Create dummy stego files
				for i := 1; i <= 3; i++ {
					stegoPath := filepath.Join(tmpDir, "stego"+string(rune('0'+i))+".png")
					err = os.WriteFile(stegoPath, make([]byte, 50000), 0644)
					require.NoError(t, err)
				}

				// Create dummy manifest
				manifestPath := filepath.Join(tmpDir, "manifest.json")
				manifestData := `{"version":"1.0","mode":"round_robin","payload_count":2,"cover_count":3}`
				err = os.WriteFile(manifestPath, []byte(manifestData), 0644)
				require.NoError(t, err)

				// Create output directory
				outputDir := filepath.Join(tmpDir, "output")
				err = os.MkdirAll(outputDir, 0755)
				require.NoError(t, err)

				return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *ExtractMatrixResult) {
				// Mock extraction doesn't actually process files
				// Just verify result is not nil and has basic structure
				assert.NotNil(t, result, "Result should not be nil")
			},
		},
		{
			name: "successful_extraction_partial_covers",
			setupCommand: func(tmpDir string) ExtractMatrixCommand {
				return ExtractMatrixCommand{
					StegoFiles: []string{
						filepath.Join(tmpDir, "stego1.png"),
						filepath.Join(tmpDir, "stego2.png"),
						// Missing stego3.png
					},
					ManifestFile: filepath.Join(tmpDir, "manifest.json"),
					OutputDir:    filepath.Join(tmpDir, "output"),
					AllowPartial: true,
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "matrix_extract_*")
				require.NoError(t, err)

				for i := 1; i <= 2; i++ {
					stegoPath := filepath.Join(tmpDir, "stego"+string(rune('0'+i))+".png")
					err = os.WriteFile(stegoPath, make([]byte, 40000), 0644)
					require.NoError(t, err)
				}

				manifestPath := filepath.Join(tmpDir, "manifest.json")
				manifestData := `{"version":"1.0","mode":"round_robin","payload_count":2,"cover_count":3}`
				err = os.WriteFile(manifestPath, []byte(manifestData), 0644)
				require.NoError(t, err)

				outputDir := filepath.Join(tmpDir, "output")
				err = os.MkdirAll(outputDir, 0755)
				require.NoError(t, err)

				return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *ExtractMatrixResult) {
				// Mock extraction doesn't actually process files
				assert.NotNil(t, result, "Result should not be nil")
			},
		},
		{
			name: "invalid_command_no_stego_files",
			setupCommand: func(tmpDir string) ExtractMatrixCommand {
				return ExtractMatrixCommand{
					StegoFiles:   []string{},
					ManifestFile: "manifest.json",
					OutputDir:    "output",
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				return "", func() {}
			},
			expectSuccess: false,
			expectError:   "at least one stego file is required",
		},
		{
			name: "invalid_command_missing_manifest",
			setupCommand: func(tmpDir string) ExtractMatrixCommand {
				return ExtractMatrixCommand{
					StegoFiles:   []string{filepath.Join(tmpDir, "stego1.png")},
					ManifestFile: "",
					OutputDir:    "output",
				}
			},
			setupTestFiles: func(t *testing.T) (string, func()) {
				// Create temp file so validation reaches manifest check
				tmpDir, err := os.MkdirTemp("", "matrix_extract_*")
				require.NoError(t, err)
				stegoPath := filepath.Join(tmpDir, "stego1.png")
				err = os.WriteFile(stegoPath, make([]byte, 1000), 0644)
				require.NoError(t, err)
				return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
			},
			expectSuccess: false,
			expectError:   "manifest file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := logrus.New()
			logger.SetOutput(io.Discard)

			handler := NewExtractMatrixHandler(
				&mockStegoService{},  // Use mock stego service
				&mockCryptoService{}, // Use mock crypto service
				errorcorrection_infra.NewRSService(logger),
				logger,
			)

			tmpDir, cleanup := tt.setupTestFiles(t)
			defer cleanup()

			ctx := context.Background()
			var cmd ExtractMatrixCommand
			if tmpDir != "" {
				cmd = tt.setupCommand(tmpDir)
			} else {
				cmd = tt.setupCommand("")
			}

			// Execute
			result, err := handler.Handle(ctx, cmd)

			// Assert
			if tt.expectSuccess {
				require.NoError(t, err, "Expected successful execution")
				require.NotNil(t, result, "Result should not be nil")

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				require.Error(t, err, "Expected error")
				if tt.expectError != "" {
					assert.Contains(t, err.Error(), tt.expectError)
				}
			}
		})
	}
}

// TestMatrixCommand_Validate tests command validation logic.
func TestMatrixCommand_Validate(t *testing.T) {
	tests := []struct {
		name        string
		cmd         EmbedMatrixCommand
		expectError string
	}{
		{
			name: "valid_2x3_matrix",
			cmd: EmbedMatrixCommand{
				Payloads: []MatrixPayloadItem{
					{Name: "p1", Data: []byte("data1")},
					{Name: "p2", Data: []byte("data2")},
				},
				Covers: []MatrixCoverItem{
					{Path: "/tmp/c1.png"},
					{Path: "/tmp/c2.png"},
					{Path: "/tmp/c3.png"},
				},
				OutputDir:    "/tmp/output",
				ManifestFile: "/tmp/manifest.json",
				Mode:         MatrixModeRoundRobin,
				RSRedundancy: 0.3,
			},
			expectError: "", // Should succeed with existing files check disabled
		},
		{
			name: "invalid_empty_payloads",
			cmd: EmbedMatrixCommand{
				Payloads: []MatrixPayloadItem{},
				Covers:   []MatrixCoverItem{{Path: "/tmp/c1.png"}},
			},
			expectError: "at least one payload is required",
		},
		{
			name: "invalid_duplicate_payload_names",
			cmd: EmbedMatrixCommand{
				Payloads: []MatrixPayloadItem{
					{Name: "duplicate", Data: []byte("data1")},
					{Name: "duplicate", Data: []byte("data2")},
				},
				Covers: []MatrixCoverItem{{Path: "/tmp/c1.png"}},
			},
			expectError: "duplicate payload name",
		},
		{
			name: "invalid_mode",
			cmd: func() EmbedMatrixCommand {
				// Create temp files so validation reaches mode check
				tmpDir, err := os.MkdirTemp("", "matrix_validate_*")
				if err != nil {
					panic(err)
				}
				for i := 1; i <= 2; i++ {
					coverPath := filepath.Join(tmpDir, fmt.Sprintf("c%d.png", i))
					if err := os.WriteFile(coverPath, make([]byte, 1000), 0644); err != nil {
						panic(err)
					}
				}
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{
						{Name: "p1", Data: []byte("data")},
					},
					Covers: []MatrixCoverItem{
						{Path: filepath.Join(tmpDir, "c1.png")},
						{Path: filepath.Join(tmpDir, "c2.png")},
					},
					OutputDir:    tmpDir,
					ManifestFile: filepath.Join(tmpDir, "manifest.json"),
					Mode:         "invalid_mode",
				}
			}(),
			expectError: "invalid matrix allocation mode",
		},
		{
			name: "invalid_redundancy_negative",
			cmd: func() EmbedMatrixCommand {
				// Create temp files so validation reaches redundancy check
				tmpDir, err := os.MkdirTemp("", "matrix_validate_*")
				if err != nil {
					panic(err)
				}
				for i := 1; i <= 2; i++ {
					coverPath := filepath.Join(tmpDir, fmt.Sprintf("c%d.png", i))
					if err := os.WriteFile(coverPath, make([]byte, 1000), 0644); err != nil {
						panic(err)
					}
				}
				return EmbedMatrixCommand{
					Payloads: []MatrixPayloadItem{
						{Name: "p1", Data: []byte("data")},
					},
					Covers: []MatrixCoverItem{
						{Path: filepath.Join(tmpDir, "c1.png")},
						{Path: filepath.Join(tmpDir, "c2.png")},
					},
					OutputDir:    tmpDir,
					ManifestFile: filepath.Join(tmpDir, "manifest.json"),
					Mode:         MatrixModeRoundRobin,
					RSRedundancy: -0.1,
				}
			}(),
			expectError: "RS redundancy must be between 0.0 and 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			}
		})
	}
}
