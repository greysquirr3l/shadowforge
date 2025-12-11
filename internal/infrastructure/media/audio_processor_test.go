package media

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAudioProcessor tests audio processor constructor.
func TestNewAudioProcessor(t *testing.T) {
	processor := NewAudioProcessor()
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.detector)
}

// createMinimalWAV creates a minimal valid WAV file for testing.
func createMinimalWAV(sampleRate, bitDepth, channels, numSamples int) []byte {
	bytesPerSample := bitDepth / 8
	dataSize := uint32(numSamples * channels * bytesPerSample)

	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(channels),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(sampleRate * channels * bytesPerSample),
		BlockAlign:    uint16(channels * bytesPerSample),
		BitsPerSample: uint16(bitDepth),
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, &header)

	// Write silent samples
	for i := 0; i < numSamples*channels; i++ {
		switch bitDepth {
		case 8:
			buf.WriteByte(128) // Silence for 8-bit unsigned
		case 16:
			binary.Write(&buf, binary.LittleEndian, int16(0))
		case 24:
			buf.Write([]byte{0, 0, 0})
		case 32:
			binary.Write(&buf, binary.LittleEndian, int32(0))
		}
	}

	return buf.Bytes()
}

// TestLoadWAV_16Bit tests loading a 16-bit WAV file.
func TestLoadWAV_16Bit(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 1000)

	pcm, err := processor.LoadWAV(wavData)

	require.NoError(t, err)
	assert.Equal(t, 44100, pcm.SampleRate)
	assert.Equal(t, 16, pcm.BitsPerSample)
	assert.Equal(t, 2, pcm.Channels)
	assert.Len(t, pcm.Samples, 2)
	assert.Len(t, pcm.Samples[0], 1000)
	assert.Len(t, pcm.Samples[1], 1000)
}

// TestLoadWAV_8Bit tests loading an 8-bit WAV file.
func TestLoadWAV_8Bit(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(8000, 8, 1, 500)

	pcm, err := processor.LoadWAV(wavData)

	require.NoError(t, err)
	assert.Equal(t, 8000, pcm.SampleRate)
	assert.Equal(t, 8, pcm.BitsPerSample)
	assert.Equal(t, 1, pcm.Channels)
	assert.Len(t, pcm.Samples, 1)
	assert.Len(t, pcm.Samples[0], 500)
}

// TestLoadWAV_24Bit tests loading a 24-bit WAV file.
func TestLoadWAV_24Bit(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(48000, 24, 2, 100)

	pcm, err := processor.LoadWAV(wavData)

	require.NoError(t, err)
	assert.Equal(t, 48000, pcm.SampleRate)
	assert.Equal(t, 24, pcm.BitsPerSample)
	assert.Equal(t, 2, pcm.Channels)
}

// TestLoadWAV_32Bit tests loading a 32-bit WAV file.
func TestLoadWAV_32Bit(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(96000, 32, 2, 200)

	pcm, err := processor.LoadWAV(wavData)

	require.NoError(t, err)
	assert.Equal(t, 96000, pcm.SampleRate)
	assert.Equal(t, 32, pcm.BitsPerSample)
	assert.Equal(t, 2, pcm.Channels)
}

// TestLoadWAV_InsufficientData tests loading with too little data.
func TestLoadWAV_InsufficientData(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := []byte{0x52, 0x49, 0x46, 0x46} // Just "RIFF"

	_, err := processor.LoadWAV(wavData)

	assert.ErrorIs(t, err, ErrInsufficientData)
}

// TestLoadWAV_InvalidFormat tests loading non-WAV data.
func TestLoadWAV_InvalidFormat(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature

	_, err := processor.LoadWAV(wavData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format")
}

// TestLoadWAV_CorruptedHeader tests loading corrupted WAV.
func TestLoadWAV_CorruptedHeader(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 100)

	// Corrupt WAVE marker
	copy(wavData[8:12], []byte("XXXX"))

	_, err := processor.LoadWAV(wavData)

	assert.Error(t, err)
}

// TestLoadWAV_NonPCMFormat tests unsupported audio format.
func TestLoadWAV_NonPCMFormat(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 100)

	// Change audio format to non-PCM (e.g., 3 = IEEE float)
	binary.LittleEndian.PutUint16(wavData[20:22], 3)

	_, err := processor.LoadWAV(wavData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported audio format")
}

// TestSaveWAV_16Bit tests saving a 16-bit WAV file.
func TestSaveWAV_16Bit(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 16,
		Channels:      2,
		Samples: [][]int32{
			{0, 100, -100, 200, -200},
			{0, 150, -150, 250, -250},
		},
	}

	wavData, err := processor.SaveWAV(pcm)

	require.NoError(t, err)
	assert.NotEmpty(t, wavData)
	assert.Equal(t, "RIFF", string(wavData[0:4]))
	assert.Equal(t, "WAVE", string(wavData[8:12]))
}

// TestSaveWAV_8Bit tests saving an 8-bit WAV file.
func TestSaveWAV_8Bit(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    8000,
		BitsPerSample: 8,
		Channels:      1,
		Samples: [][]int32{
			{0, 10, -10, 20, -20},
		},
	}

	wavData, err := processor.SaveWAV(pcm)

	require.NoError(t, err)
	assert.NotEmpty(t, wavData)
}

// TestSaveWAV_24Bit tests saving a 24-bit WAV file.
func TestSaveWAV_24Bit(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    48000,
		BitsPerSample: 24,
		Channels:      2,
		Samples: [][]int32{
			{0, 1000, -1000},
			{0, 1500, -1500},
		},
	}

	wavData, err := processor.SaveWAV(pcm)

	require.NoError(t, err)
	assert.NotEmpty(t, wavData)
}

// TestSaveWAV_32Bit tests saving a 32-bit WAV file.
func TestSaveWAV_32Bit(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    96000,
		BitsPerSample: 32,
		Channels:      2,
		Samples: [][]int32{
			{0, 10000, -10000},
			{0, 15000, -15000},
		},
	}

	wavData, err := processor.SaveWAV(pcm)

	require.NoError(t, err)
	assert.NotEmpty(t, wavData)
}

// TestSaveWAV_NilPCM tests saving with nil PCM data.
func TestSaveWAV_NilPCM(t *testing.T) {
	processor := NewAudioProcessor()

	_, err := processor.SaveWAV(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PCM data is nil")
}

// TestSaveWAV_NoChannels tests saving with no channels.
func TestSaveWAV_NoChannels(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 16,
		Channels:      0,
		Samples:       [][]int32{},
	}

	_, err := processor.SaveWAV(pcm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio channels")
}

// TestSaveWAV_NoSamples tests saving with no samples.
func TestSaveWAV_NoSamples(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 16,
		Channels:      2,
		Samples: [][]int32{
			{},
		},
	}

	_, err := processor.SaveWAV(pcm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio samples")
}

// TestSaveWAV_InvalidBitDepth tests saving with unsupported bit depth.
func TestSaveWAV_InvalidBitDepth(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 12, // Unsupported
		Channels:      2,
		Samples: [][]int32{
			{0, 100},
			{0, 150},
		},
	}

	_, err := processor.SaveWAV(pcm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported bit depth")
}

// TestWAVRoundTrip tests full WAV encode/decode cycle.
func TestWAVRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
		bitDepth   int
		channels   int
	}{
		{"16-bit Stereo 44.1kHz", 44100, 16, 2},
		{"8-bit Mono 8kHz", 8000, 8, 1},
		{"24-bit Stereo 48kHz", 48000, 24, 2},
		{"32-bit Stereo 96kHz", 96000, 32, 2},
	}

	processor := NewAudioProcessor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original PCM data
			original := &PCMData{
				SampleRate:    tt.sampleRate,
				BitsPerSample: tt.bitDepth,
				Channels:      tt.channels,
				Samples:       make([][]int32, tt.channels),
			}

			// Create test samples with pattern
			for ch := 0; ch < tt.channels; ch++ {
				if tt.bitDepth == 8 {
					// 8-bit samples have range -128 to 127
					original.Samples[ch] = []int32{0, 50, -50, 100, -100, 127, -128}
				} else {
					original.Samples[ch] = []int32{0, 100, -100, 200, -200, 300, -300}
				}
			}

			// Encode to WAV
			wavData, err := processor.SaveWAV(original)
			require.NoError(t, err)

			// Decode back
			recovered, err := processor.LoadWAV(wavData)
			require.NoError(t, err)

			// Verify metadata
			assert.Equal(t, original.SampleRate, recovered.SampleRate)
			assert.Equal(t, original.BitsPerSample, recovered.BitsPerSample)
			assert.Equal(t, original.Channels, recovered.Channels)

			// Verify samples (with tolerance for bit depth conversion)
			for ch := 0; ch < tt.channels; ch++ {
				assert.Len(t, recovered.Samples[ch], len(original.Samples[ch]))
				for i := range original.Samples[ch] {
					// Allow small deviation for 8-bit (due to unsigned conversion)
					if tt.bitDepth == 8 {
						assert.InDelta(t, original.Samples[ch][i], recovered.Samples[ch][i], 1)
					} else {
						assert.Equal(t, original.Samples[ch][i], recovered.Samples[ch][i])
					}
				}
			}
		})
	}
}

// TestGetAudioInfo tests audio metadata extraction.
func TestGetAudioInfo(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 1000)

	info, err := processor.GetAudioInfo(wavData)

	require.NoError(t, err)
	assert.Equal(t, media.FormatWAV, info.Format)
	assert.Equal(t, 44100, info.SampleRate)
	assert.Equal(t, 16, info.BitDepth)
	assert.Equal(t, 2, info.Channels)
}

// TestCalculateCapacity_LSB tests capacity calculation for LSB technique.
func TestCalculateCapacity_LSB(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 44100) // 1 second

	capacity, err := processor.CalculateCapacity(wavData, "lsb")

	require.NoError(t, err)
	assert.Equal(t, media.MediaTypeAudio, capacity.MediaType)
	assert.Greater(t, capacity.TotalCapacity, int64(0))
	assert.Greater(t, capacity.UsableCapacity, int64(0))
	assert.LessOrEqual(t, capacity.UsableCapacity, capacity.TotalCapacity)

	// LSB: 1 bit per sample = 44100 bits = 5512 bytes raw
	expectedRaw := int64(44100 / 8)
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)

	// Safe capacity should be 80% of raw
	expectedSafe := int64(float64(expectedRaw) * 0.80)
	assert.Equal(t, expectedSafe, capacity.UsableCapacity)
}

// TestCalculateCapacity_LSB2 tests capacity for 2-bit LSB.
func TestCalculateCapacity_LSB2(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 44100)

	capacity, err := processor.CalculateCapacity(wavData, "lsb-2")

	require.NoError(t, err)

	// LSB-2: 2 bits per sample = 88200 bits = 11025 bytes raw
	expectedRaw := int64((44100 * 2) / 8)
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)

	// Safe capacity should be 70% of raw
	expectedSafe := int64(float64(expectedRaw) * 0.70)
	assert.Equal(t, expectedSafe, capacity.UsableCapacity)
}

// TestCalculateCapacity_Phase tests capacity for phase encoding.
func TestCalculateCapacity_Phase(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 44100)

	capacity, err := processor.CalculateCapacity(wavData, "phase")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))

	// Phase: 1 bit per 1024 samples
	segments := int64(44100 / 1024)
	expectedRaw := segments / 8
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)
}

// TestCalculateCapacity_Echo tests capacity for echo hiding.
func TestCalculateCapacity_Echo(t *testing.T) {
	processor := NewAudioProcessor()
	// Use 5 seconds of audio to ensure we have enough samples
	wavData := createMinimalWAV(44100, 16, 2, 44100*5)

	capacity, err := processor.CalculateCapacity(wavData, "echo")

	require.NoError(t, err)
	assert.Greater(t, capacity.TotalCapacity, int64(0))

	// Echo: 1 bit per 8192 samples (very conservative)
	expectedRaw := int64((44100 * 5) / (8192 * 8))
	assert.Equal(t, expectedRaw, capacity.TotalCapacity)
}

// TestCalculateCapacity_UnknownTechnique tests with invalid technique.
func TestCalculateCapacity_UnknownTechnique(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 1000)

	_, err := processor.CalculateCapacity(wavData, "invalid-technique")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown technique")
}

// TestCalculateCapacity_NonWAV tests capacity with non-WAV format.
func TestCalculateCapacity_NonWAV(t *testing.T) {
	processor := NewAudioProcessor()

	// Create minimal FLAC data using the helper from format_detector_test.go
	flacData := createMinimalFLAC(44100, 16, 2)

	_, err := processor.CalculateCapacity(flacData, "lsb")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only supported for WAV")
}

// TestCalculateCapacity_InvalidData tests capacity with invalid data.
func TestCalculateCapacity_InvalidData(t *testing.T) {
	processor := NewAudioProcessor()
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}

	_, err := processor.CalculateCapacity(invalidData, "lsb")

	assert.Error(t, err)
}

// TestCalculateQualityScore tests audio quality assessment.
func TestCalculateQualityScore(t *testing.T) {
	processor := NewAudioProcessor()

	tests := []struct {
		name         string
		info         *AudioInfo
		totalSamples int64
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name: "High Quality",
			info: &AudioInfo{
				SampleRate: 48000,
				BitDepth:   24,
				Channels:   2,
			},
			totalSamples: 48000 * 120, // 2 minutes
			expectedMin:  0.8,
			expectedMax:  1.0,
		},
		{
			name: "Medium Quality",
			info: &AudioInfo{
				SampleRate: 44100,
				BitDepth:   16,
				Channels:   2,
			},
			totalSamples: 44100 * 30, // 30 seconds
			expectedMin:  0.5,
			expectedMax:  0.8,
		},
		{
			name: "Low Quality",
			info: &AudioInfo{
				SampleRate: 8000,
				BitDepth:   8,
				Channels:   1,
			},
			totalSamples: 8000 * 2, // 2 seconds
			expectedMin:  0.0,
			expectedMax:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := processor.calculateQualityScore(tt.info, tt.totalSamples)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

// TestValidateAudio_WAV tests WAV validation.
func TestValidateAudio_WAV(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 1000)

	err := processor.ValidateAudio(wavData)

	assert.NoError(t, err)
}

// TestValidateAudio_FLAC tests FLAC validation.
func TestValidateAudio_FLAC(t *testing.T) {
	processor := NewAudioProcessor()
	flacData := []byte{0x66, 0x4C, 0x61, 0x43} // "fLaC"

	err := processor.ValidateAudio(flacData)

	assert.NoError(t, err)
}

// TestValidateAudio_MP3 tests MP3 validation.
func TestValidateAudio_MP3(t *testing.T) {
	processor := NewAudioProcessor()

	tests := []struct {
		name string
		data []byte
	}{
		{"ID3 Tag", []byte{0x49, 0x44, 0x33, 0x00, 0x00}}, // "ID3" + version
		{"MP3 Sync", []byte{0xFF, 0xFB, 0x90, 0x00}},      // MP3 frame sync
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateAudio(tt.data)
			assert.NoError(t, err)
		})
	}
}

// TestValidateAudio_Empty tests validation with empty data.
func TestValidateAudio_Empty(t *testing.T) {
	processor := NewAudioProcessor()

	err := processor.ValidateAudio([]byte{})

	assert.ErrorIs(t, err, ErrInsufficientData)
}

// TestValidateAudio_InvalidFormat tests validation with invalid format.
func TestValidateAudio_InvalidFormat(t *testing.T) {
	processor := NewAudioProcessor()
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}

	err := processor.ValidateAudio(invalidData)

	assert.Error(t, err)
}

// TestValidateAudio_CorruptedWAV tests validation with corrupted WAV.
func TestValidateAudio_CorruptedWAV(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 100)

	// Corrupt the RIFF header
	wavData[0] = 0x00

	err := processor.ValidateAudio(wavData)

	assert.Error(t, err)
}

// TestValidateAudio_NonPCMWAV tests validation with non-PCM WAV.
func TestValidateAudio_NonPCMWAV(t *testing.T) {
	processor := NewAudioProcessor()
	wavData := createMinimalWAV(44100, 16, 2, 100)

	// Change to IEEE float format
	binary.LittleEndian.PutUint16(wavData[20:22], 3)

	err := processor.ValidateAudio(wavData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported audio format")
}

// TestValidateAudio_NotAudioFile tests validation with non-audio data.
func TestValidateAudio_NotAudioFile(t *testing.T) {
	processor := NewAudioProcessor()
	// PNG signature
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	err := processor.ValidateAudio(pngData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an audio file")
}

// TestCreateSilence tests creating silent audio.
func TestCreateSilence(t *testing.T) {
	processor := NewAudioProcessor()

	pcm, err := processor.CreateSilence(44100, 16, 2, 1.0)

	require.NoError(t, err)
	assert.Equal(t, 44100, pcm.SampleRate)
	assert.Equal(t, 16, pcm.BitsPerSample)
	assert.Equal(t, 2, pcm.Channels)
	assert.Len(t, pcm.Samples, 2)
	assert.Len(t, pcm.Samples[0], 44100)

	// Verify all samples are zero (silence)
	for ch := 0; ch < 2; ch++ {
		for _, sample := range pcm.Samples[ch] {
			assert.Equal(t, int32(0), sample)
		}
	}
}

// TestCreateSilence_InvalidParameters tests creation with invalid params.
func TestCreateSilence_InvalidParameters(t *testing.T) {
	processor := NewAudioProcessor()

	tests := []struct {
		name       string
		sampleRate int
		bitDepth   int
		channels   int
		duration   float64
	}{
		{"Zero Sample Rate", 0, 16, 2, 1.0},
		{"Zero Bit Depth", 44100, 0, 2, 1.0},
		{"Zero Channels", 44100, 16, 0, 1.0},
		{"Zero Duration", 44100, 16, 2, 0.0},
		{"Negative Duration", 44100, 16, 2, -1.0},
		{"Invalid Bit Depth", 44100, 12, 2, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.CreateSilence(tt.sampleRate, tt.bitDepth, tt.channels, tt.duration)
			assert.Error(t, err)
		})
	}
}

// TestModifySamples tests sample modification for steganography.
func TestModifySamples(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 16,
		Channels:      2,
		Samples: [][]int32{
			{100, 200, 300},
			{150, 250, 350},
		},
	}

	// Modifier that adds 10 to each sample
	modifier := func(channel, sample int, value int32) int32 {
		return value + 10
	}

	err := processor.ModifySamples(pcm, modifier)

	require.NoError(t, err)
	assert.Equal(t, int32(110), pcm.Samples[0][0])
	assert.Equal(t, int32(210), pcm.Samples[0][1])
	assert.Equal(t, int32(310), pcm.Samples[0][2])
	assert.Equal(t, int32(160), pcm.Samples[1][0])
	assert.Equal(t, int32(260), pcm.Samples[1][1])
	assert.Equal(t, int32(360), pcm.Samples[1][2])
}

// TestModifySamples_NilPCM tests modification with nil PCM.
func TestModifySamples_NilPCM(t *testing.T) {
	processor := NewAudioProcessor()

	modifier := func(channel, sample int, value int32) int32 {
		return value
	}

	err := processor.ModifySamples(nil, modifier)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PCM data is nil")
}

// TestModifySamples_LSBEmbedding simulates LSB embedding.
func TestModifySamples_LSBEmbedding(t *testing.T) {
	processor := NewAudioProcessor()

	pcm := &PCMData{
		SampleRate:    44100,
		BitsPerSample: 16,
		Channels:      1,
		Samples: [][]int32{
			{1000, 1001, 1002, 1003, 1004},
		},
	}

	// LSB modifier: set LSB to specific pattern (alternating 0, 1)
	modifier := func(channel, sample int, value int32) int32 {
		// Clear LSB
		value &= ^int32(1)
		// Set LSB based on sample index
		if sample%2 == 1 {
			value |= 1
		}
		return value
	}

	err := processor.ModifySamples(pcm, modifier)

	require.NoError(t, err)

	// Verify LSB pattern
	assert.Equal(t, int32(0), pcm.Samples[0][0]&1) // Even index -> 0
	assert.Equal(t, int32(1), pcm.Samples[0][1]&1) // Odd index -> 1
	assert.Equal(t, int32(0), pcm.Samples[0][2]&1)
	assert.Equal(t, int32(1), pcm.Samples[0][3]&1)
	assert.Equal(t, int32(0), pcm.Samples[0][4]&1)
}

// Additional coverage tests for validation functions

func TestValidateWAV(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_WAV_Minimal",
			data: []byte{
				// RIFF header
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x24, 0x00, 0x00, 0x00, // Chunk size
				0x57, 0x41, 0x56, 0x45, // "WAVE"
				// fmt subchunk
				0x66, 0x6D, 0x74, 0x20, // "fmt "
				0x10, 0x00, 0x00, 0x00, // Subchunk1Size (16 for PCM)
				0x01, 0x00, // AudioFormat (1 = PCM)
				0x02, 0x00, // NumChannels (2 = stereo)
				0x44, 0xAC, 0x00, 0x00, // SampleRate (44100)
				0x10, 0xB1, 0x02, 0x00, // ByteRate
				0x04, 0x00, // BlockAlign
				0x10, 0x00, // BitsPerSample (16)
				// data subchunk
				0x64, 0x61, 0x74, 0x61, // "data"
				0x00, 0x00, 0x00, 0x00, // Subchunk2Size
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x52, 0x49, 0x46, 0x46},
			expectError: true,
		},
		{
			name: "InvalidRIFF_Error",
			data: []byte{
				0x00, 0x00, 0x00, 0x00,
				0x24, 0x00, 0x00, 0x00,
				0x57, 0x41, 0x56, 0x45,
			},
			expectError: true,
		},
		{
			name: "MissingFmt_Error",
			data: []byte{
				0x52, 0x49, 0x46, 0x46,
				0x24, 0x00, 0x00, 0x00,
				0x57, 0x41, 0x56, 0x45,
				// No fmt chunk
			},
			expectError: true,
		},
		{
			name: "MissingData_Error",
			data: []byte{
				0x52, 0x49, 0x46, 0x46,
				0x24, 0x00, 0x00, 0x00,
				0x57, 0x41, 0x56, 0x45,
				0x66, 0x6D, 0x74, 0x20,
				0x10, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x02, 0x00,
				0x44, 0xAC, 0x00, 0x00,
				0x10, 0xB1, 0x02, 0x00,
				0x04, 0x00, 0x10, 0x00,
				// No data chunk
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewAudioProcessor()
			err := processor.validateWAV(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFLAC(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_FLAC_Minimal",
			data: []byte{
				// fLaC marker
				0x66, 0x4C, 0x61, 0x43,
				// STREAMINFO block header (last block = 1, type = 0)
				0x80,             // Last-metadata-block flag (1) + block type STREAMINFO (0)
				0x00, 0x00, 0x22, // Block length (34 bytes)
				// STREAMINFO data (minimum 34 bytes)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x66, 0x4C, 0x61},
			expectError: true,
		},
		{
			name: "InvalidSignature_Error",
			data: []byte{
				0x00, 0x00, 0x00, 0x00,
				0x80, 0x00, 0x00, 0x22,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewAudioProcessor()
			err := processor.validateFLAC(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMP3(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_MP3_WithID3",
			data: []byte{
				// ID3v2 header
				0x49, 0x44, 0x33, // "ID3"
				0x03, 0x00, // Version 2.3.0
				0x00,                   // Flags
				0x00, 0x00, 0x00, 0x00, // Size (0 for minimal)
				// MP3 frame header (MPEG-1 Layer III, 128 kbps, 44.1 kHz)
				0xFF, 0xFB, // Frame sync + MPEG-1 Layer III
				0x90, 0x00, // Bitrate index + sample rate + padding + private
			},
			expectError: false,
		},
		{
			name: "Valid_MP3_NoID3",
			data: []byte{
				// MP3 frame header only
				0xFF, 0xFB,
				0x90, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0xFF},
			expectError: true,
		},
		{
			name: "NoFrameSync_Error",
			data: []byte{
				0x00, 0x00, 0x00, 0x00,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewAudioProcessor()
			err := processor.validateMP3(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
