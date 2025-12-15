package testutil

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSimpleWAV(t *testing.T) {
	// Act
	wavData := GenerateSimpleWAV()

	// Assert
	require.NotNil(t, wavData)
	assert.Greater(t, len(wavData), 44, "WAV file should be larger than header")

	// Verify RIFF header
	assert.Equal(t, "RIFF", string(wavData[0:4]))
	assert.Equal(t, "WAVE", string(wavData[8:12]))
	assert.Equal(t, "fmt ", string(wavData[12:16]))
	assert.Equal(t, "data", string(wavData[36:40]))
}

func TestGenerateStereo16BitWAV(t *testing.T) {
	tests := []struct {
		name       string
		numSamples int
	}{
		{"100 samples", 100},
		{"1000 samples", 1000},
		{"10000 samples", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			wavData := GenerateStereo16BitWAV(tt.numSamples)

			// Assert - Read header
			reader := bytes.NewReader(wavData)
			var header WAVHeader
			err := binary.Read(reader, binary.LittleEndian, &header)
			require.NoError(t, err)

			// Verify header values
			assert.Equal(t, uint16(2), header.NumChannels, "Should be stereo")
			assert.Equal(t, uint16(16), header.BitsPerSample, "Should be 16-bit")
			assert.Equal(t, uint32(44100), header.SampleRate, "Should be 44.1kHz")

			// Verify data size
			expectedDataSize := uint32(tt.numSamples * 2 * 2) // samples * channels * bytes_per_sample
			assert.Equal(t, expectedDataSize, header.Subchunk2Size)

			// Verify total file size
			expectedFileSize := 44 + int(expectedDataSize)
			assert.Equal(t, expectedFileSize, len(wavData))
		})
	}
}

func TestGenerateMono16BitWAV(t *testing.T) {
	// Arrange
	numSamples := 1000

	// Act
	wavData := GenerateMono16BitWAV(numSamples)

	// Assert
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	assert.Equal(t, uint16(1), header.NumChannels, "Should be mono")
	assert.Equal(t, uint16(16), header.BitsPerSample)

	expectedDataSize := uint32(numSamples * 1 * 2) // mono * 16-bit
	assert.Equal(t, expectedDataSize, header.Subchunk2Size)
}

func TestGenerate8BitWAV(t *testing.T) {
	// Arrange
	numSamples := 500

	// Act
	wavData := Generate8BitWAV(numSamples)

	// Assert
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	assert.Equal(t, uint16(8), header.BitsPerSample, "Should be 8-bit")
	assert.Equal(t, uint16(1), header.NumChannels, "Should be mono")
	assert.Equal(t, uint32(8000), header.SampleRate, "Should be 8kHz")

	// Verify data size (8-bit = 1 byte per sample)
	expectedDataSize := uint32(numSamples * 1 * 1)
	assert.Equal(t, expectedDataSize, header.Subchunk2Size)

	// Verify 8-bit samples are centered at 128 (unsigned)
	samples := wavData[44:] // Skip header
	for i, sample := range samples {
		assert.Equal(t, uint8(128), sample, "8-bit sample %d should be 128 (silence)", i)
	}
}

func TestGenerate24BitWAV(t *testing.T) {
	// Arrange
	numSamples := 1000

	// Act
	wavData := Generate24BitWAV(numSamples)

	// Assert
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	assert.Equal(t, uint16(24), header.BitsPerSample, "Should be 24-bit")
	assert.Equal(t, uint16(2), header.NumChannels, "Should be stereo")
	assert.Equal(t, uint32(48000), header.SampleRate, "Should be 48kHz")

	// Verify data size (24-bit = 3 bytes per sample)
	expectedDataSize := uint32(numSamples * 2 * 3)
	assert.Equal(t, expectedDataSize, header.Subchunk2Size)
}

func TestGenerate32BitWAV(t *testing.T) {
	// Arrange
	numSamples := 1000

	// Act
	wavData := Generate32BitWAV(numSamples)

	// Assert
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	assert.Equal(t, uint16(32), header.BitsPerSample, "Should be 32-bit")
	assert.Equal(t, uint16(2), header.NumChannels, "Should be stereo")
	assert.Equal(t, uint32(48000), header.SampleRate, "Should be 48kHz")

	// Verify data size (32-bit = 4 bytes per sample)
	expectedDataSize := uint32(numSamples * 2 * 4)
	assert.Equal(t, expectedDataSize, header.Subchunk2Size)
}

func TestGenerateWAV_CustomConfig(t *testing.T) {
	// Arrange
	config := WAVConfig{
		SampleRate:    22050,
		BitDepth:      16,
		Channels:      1,
		NumSamples:    500,
		GenerateTone:  false,
		ToneFrequency: 0,
	}

	// Act
	wavData := GenerateWAV(config)

	// Assert
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	assert.Equal(t, uint32(22050), header.SampleRate)
	assert.Equal(t, uint16(16), header.BitsPerSample)
	assert.Equal(t, uint16(1), header.NumChannels)

	expectedDataSize := uint32(500 * 1 * 2)
	assert.Equal(t, expectedDataSize, header.Subchunk2Size)
}

func TestDefaultWAVConfig(t *testing.T) {
	// Act
	config := DefaultWAVConfig()

	// Assert
	assert.Equal(t, 44100, config.SampleRate)
	assert.Equal(t, 16, config.BitDepth)
	assert.Equal(t, 2, config.Channels)
	assert.Equal(t, 1000, config.NumSamples)
	assert.False(t, config.GenerateTone)
	assert.Equal(t, 0, config.ToneFrequency)
}

func TestGetWAVCapacity(t *testing.T) {
	tests := []struct {
		name             string
		config           WAVConfig
		expectedCapacity int
	}{
		{
			name: "1000 samples stereo 16-bit",
			config: WAVConfig{
				SampleRate: 44100,
				BitDepth:   16,
				Channels:   2,
				NumSamples: 1000,
			},
			expectedCapacity: 250, // 1000 * 2 / 8 = 250 bytes
		},
		{
			name: "1000 samples mono 16-bit",
			config: WAVConfig{
				SampleRate: 44100,
				BitDepth:   16,
				Channels:   1,
				NumSamples: 1000,
			},
			expectedCapacity: 125, // 1000 * 1 / 8 = 125 bytes
		},
		{
			name: "10000 samples stereo 16-bit",
			config: WAVConfig{
				SampleRate: 44100,
				BitDepth:   16,
				Channels:   2,
				NumSamples: 10000,
			},
			expectedCapacity: 2500, // 10000 * 2 / 8 = 2500 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			capacity := GetWAVCapacity(tt.config)

			// Assert
			assert.Equal(t, tt.expectedCapacity, capacity)
		})
	}
}

func TestGetWAVCapacityForFile(t *testing.T) {
	tests := []struct {
		name             string
		generateFunc     func(int) []byte
		numSamples       int
		expectedCapacity int
	}{
		{
			name:             "Stereo 16-bit 1000 samples",
			generateFunc:     GenerateStereo16BitWAV,
			numSamples:       1000,
			expectedCapacity: 250, // 1000 * 2 / 8
		},
		{
			name:             "Mono 16-bit 1000 samples",
			generateFunc:     GenerateMono16BitWAV,
			numSamples:       1000,
			expectedCapacity: 125, // 1000 * 1 / 8
		},
		{
			name:             "8-bit 500 samples",
			generateFunc:     Generate8BitWAV,
			numSamples:       500,
			expectedCapacity: 62, // 500 * 1 / 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			wavData := tt.generateFunc(tt.numSamples)

			// Act
			capacity := GetWAVCapacityForFile(wavData)

			// Assert
			assert.Equal(t, tt.expectedCapacity, capacity)
		})
	}
}

func TestGetWAVCapacityForFile_InvalidData(t *testing.T) {
	tests := []struct {
		name    string
		wavData []byte
	}{
		{"empty data", []byte{}},
		{"too short", []byte{1, 2, 3, 4}},
		{"incomplete header", make([]byte, 30)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			capacity := GetWAVCapacityForFile(tt.wavData)

			// Assert
			assert.Equal(t, 0, capacity, "Should return 0 for invalid data")
		})
	}
}

func TestWAVHeader_RIFFFormat(t *testing.T) {
	// Arrange
	wavData := GenerateSimpleWAV()

	// Act - Parse header manually
	chunkID := string(wavData[0:4])
	format := string(wavData[8:12])
	subchunk1ID := string(wavData[12:16])
	subchunk2ID := string(wavData[36:40])

	// Assert - Verify RIFF format compliance
	assert.Equal(t, "RIFF", chunkID, "ChunkID must be 'RIFF'")
	assert.Equal(t, "WAVE", format, "Format must be 'WAVE'")
	assert.Equal(t, "fmt ", subchunk1ID, "Subchunk1ID must be 'fmt '")
	assert.Equal(t, "data", subchunk2ID, "Subchunk2ID must be 'data'")
}

func TestWAVHeader_ChunkSizeCalculation(t *testing.T) {
	// Arrange
	numSamples := 1000
	wavData := GenerateStereo16BitWAV(numSamples)

	// Act - Read chunk size from header
	chunkSizeBytes := wavData[4:8]
	chunkSize := binary.LittleEndian.Uint32(chunkSizeBytes)

	// Assert - ChunkSize should be file_size - 8
	expectedChunkSize := uint32(len(wavData) - 8)
	assert.Equal(t, expectedChunkSize, chunkSize, "ChunkSize = file_size - 8")
}

func TestWAVHeader_ByteRateCalculation(t *testing.T) {
	// Arrange
	wavData := GenerateStereo16BitWAV(1000)

	// Act
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	// Assert - ByteRate = SampleRate * NumChannels * BitsPerSample/8
	expectedByteRate := header.SampleRate * uint32(header.NumChannels) * uint32(header.BitsPerSample/8)
	assert.Equal(t, expectedByteRate, header.ByteRate)
}

func TestWAVHeader_BlockAlignCalculation(t *testing.T) {
	// Arrange
	wavData := GenerateStereo16BitWAV(1000)

	// Act
	reader := bytes.NewReader(wavData)
	var header WAVHeader
	require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))

	// Assert - BlockAlign = NumChannels * BitsPerSample/8
	expectedBlockAlign := header.NumChannels * (header.BitsPerSample / 8)
	assert.Equal(t, expectedBlockAlign, header.BlockAlign)
}

func TestAllBitDepths_ValidWAVFiles(t *testing.T) {
	bitDepths := []struct {
		name      string
		generator func(int) []byte
		bitDepth  uint16
	}{
		{"8-bit", Generate8BitWAV, 8},
		{"16-bit mono", GenerateMono16BitWAV, 16},
		{"16-bit stereo", GenerateStereo16BitWAV, 16},
		{"24-bit", Generate24BitWAV, 24},
		{"32-bit", Generate32BitWAV, 32},
	}

	for _, bd := range bitDepths {
		t.Run(bd.name, func(t *testing.T) {
			// Act
			wavData := bd.generator(1000)

			// Assert - Valid RIFF/WAVE structure
			assert.Equal(t, "RIFF", string(wavData[0:4]))
			assert.Equal(t, "WAVE", string(wavData[8:12]))

			// Assert - Correct bit depth in header
			reader := bytes.NewReader(wavData)
			var header WAVHeader
			require.NoError(t, binary.Read(reader, binary.LittleEndian, &header))
			assert.Equal(t, bd.bitDepth, header.BitsPerSample)

			// Assert - PCM format
			assert.Equal(t, uint16(1), header.AudioFormat, "Should be PCM")
		})
	}
}

// Benchmark tests
func BenchmarkGenerateSimpleWAV(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSimpleWAV()
	}
}

func BenchmarkGenerateStereo16BitWAV_1000Samples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateStereo16BitWAV(1000)
	}
}

func BenchmarkGenerateStereo16BitWAV_10000Samples(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateStereo16BitWAV(10000)
	}
}

func BenchmarkGetWAVCapacityForFile(b *testing.B) {
	wavData := GenerateStereo16BitWAV(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetWAVCapacityForFile(wavData)
	}
}
