// Package stego provides steganography technique implementations.
//
// This file implements LSB (Least Significant Bit) steganography for audio files,
// specifically targeting WAV audio samples.
package stego

import (
	"context"
	"log/slog"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
)

// LSBAudioEmbeddingConfig configures LSB audio steganography parameters.
type LSBAudioEmbeddingConfig struct {
	BitsPerSample      int     // Number of LSBs to modify per sample (1-4)
	ChannelMask        uint8   // Which channels to use (bit mask: 1=left, 2=right, 3=both)
	SampleInterval     int     // Embed in every Nth sample (default: 1)
	AmplitudeThreshold float64 // Minimum amplitude to consider for embedding (0.0-1.0)
	UseRandomOrder     bool    // Use PRNG-based sample order
	PreserveSilence    bool    // Skip silent regions
	QualityLevel       int     // Quality level 1-5 (affects embedding density)
}

// DefaultLSBAudioConfig returns default LSB audio steganography configuration.
func DefaultLSBAudioConfig() LSBAudioEmbeddingConfig {
	return LSBAudioEmbeddingConfig{
		BitsPerSample:      1,    // Conservative single bit
		ChannelMask:        3,    // Both channels
		SampleInterval:     1,    // Every sample
		AmplitudeThreshold: 0.01, // Skip very quiet samples
		UseRandomOrder:     false,
		PreserveSilence:    true,
		QualityLevel:       3, // Balanced quality
	}
}

// LSBAudioTechnique implements LSB steganography for audio files.
//
// This technique modifies the least significant bits of audio samples
// to embed hidden data. It's designed to be imperceptible to human
// hearing while providing reasonable capacity.
type LSBAudioTechnique struct {
	config LSBAudioEmbeddingConfig
	logger *slog.Logger
}

// NewLSBAudioTechnique creates a new LSB audio steganography technique.
func NewLSBAudioTechnique(config LSBAudioEmbeddingConfig, logger *slog.Logger) *LSBAudioTechnique {
	if logger == nil {
		logger = slog.Default()
	}

	// Validate configuration
	if config.BitsPerSample < 1 || config.BitsPerSample > 4 {
		config.BitsPerSample = 1
	}
	if config.ChannelMask == 0 {
		config.ChannelMask = 1 // Default to left channel
	}
	if config.SampleInterval < 1 {
		config.SampleInterval = 1
	}
	if config.AmplitudeThreshold < 0 || config.AmplitudeThreshold > 1 {
		config.AmplitudeThreshold = 0.01
	}
	if config.QualityLevel < 1 || config.QualityLevel > 5 {
		config.QualityLevel = 3
	}

	return &LSBAudioTechnique{
		config: config,
		logger: logger,
	}
}

// NewLSBAudioWithDefaults creates an LSB audio technique with default configuration.
func NewLSBAudioWithDefaults(logger *slog.Logger) *LSBAudioTechnique {
	return NewLSBAudioTechnique(DefaultLSBAudioConfig(), logger)
}

// Name returns the technique identifier.
func (t *LSBAudioTechnique) Name() string {
	return "lsb_audio"
}

// SupportsFormat checks if this technique can be applied to the given format.
func (t *LSBAudioTechnique) SupportsFormat(format media.MediaFormat) bool {
	switch format {
	case media.FormatWAV:
		return true
	default:
		return false
	}
}

// Embed hides payload data in audio samples using LSB modification.
func (t *LSBAudioTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
	t.logger.Info("LSB audio embedding started",
		slog.Int("carrier_size", len(carrier)),
		slog.Int("payload_size", len(payload)),
		slog.Int("bits_per_sample", t.config.BitsPerSample))

	if len(payload) == 0 {
		return nil, stego.ErrEmptyPayload
	}

	if len(carrier) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	// Parse audio samples from carrier
	samples, sampleRate, bitDepth, channels, err := t.parseAudioData(carrier)
	if err != nil {
		return nil, err
	}

	// Check capacity
	capacity := t.calculateSampleCapacity(samples, channels)
	if capacity < len(payload) {
		return nil, stego.ErrInsufficientCapacity
	}

	// Convert payload to bits
	payloadBits := t.payloadToBits(payload)

	// Embed payload bits into samples
	modifiedSamples, err := t.embedBitsInSamples(samples, payloadBits, channels)
	if err != nil {
		return nil, err
	}

	// Reconstruct audio data
	stegoAudio, err := t.reconstructAudioData(modifiedSamples, sampleRate, bitDepth, channels)
	if err != nil {
		return nil, err
	}

	t.logger.Info("LSB audio embedding completed",
		slog.Int("embedded_bits", len(payloadBits)),
		slog.Int("modified_samples", len(modifiedSamples)))

	return stegoAudio, nil
}

// Extract retrieves hidden data from audio samples by reading LSBs.
func (t *LSBAudioTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	t.logger.Info("LSB audio extraction started",
		slog.Int("carrier_size", len(carrier)),
		slog.Int("bits_per_sample", t.config.BitsPerSample))

	if len(carrier) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	// Parse audio samples from carrier
	samples, _, _, channels, err := t.parseAudioData(carrier)
	if err != nil {
		return nil, err
	}

	// Extract bits from samples
	extractedBits, err := t.extractBitsFromSamples(samples, channels)
	if err != nil {
		return nil, err
	}

	if len(extractedBits) == 0 {
		return nil, stego.ErrNoEmbeddedData
	}

	// Convert bits back to payload
	payload := t.bitsToPayload(extractedBits)

	t.logger.Info("LSB audio extraction completed",
		slog.Int("extracted_bits", len(extractedBits)),
		slog.Int("payload_size", len(payload)))

	return payload, nil
}

// CalculateCapacity determines the embedding capacity for LSB audio steganography.
func (t *LSBAudioTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	if len(carrier) == 0 {
		return 0, stego.ErrInsufficientCapacity
	}

	// Parse audio metadata
	samples, _, _, channels, err := t.parseAudioData(carrier)
	if err != nil {
		return 0, err
	}

	capacity := t.calculateSampleCapacity(samples, channels)
	return capacity, nil
}

// parseAudioData extracts audio parameters from raw audio data.
// This is a simplified implementation - in reality would need proper WAV parsing.
func (t *LSBAudioTechnique) parseAudioData(data []byte) ([][]int16, int, int, int, error) {
	// Simplified audio parsing for demonstration
	// In production, would use proper WAV file parsing

	// Assume 16-bit PCM, 44.1kHz, stereo for now
	sampleRate := 44100
	bitDepth := 16
	channels := 2

	// Skip WAV header (assume it exists and is 44 bytes)
	headerSize := 44
	if len(data) < headerSize {
		return nil, 0, 0, 0, stego.ErrEmptyCoverMedia
	}

	audioData := data[headerSize:]
	numSamples := len(audioData) / (bitDepth / 8) / channels

	if numSamples == 0 {
		return nil, 0, 0, 0, stego.ErrEmptyCoverMedia
	}

	// Parse samples
	samples := make([][]int16, channels)
	for ch := 0; ch < channels; ch++ {
		samples[ch] = make([]int16, numSamples)
	}

	// Read interleaved 16-bit samples
	for i := 0; i < numSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			offset := (i*channels + ch) * 2
			if offset+1 < len(audioData) {
				// Little-endian 16-bit
				sample := int16(audioData[offset]) | (int16(audioData[offset+1]) << 8)
				samples[ch][i] = sample
			}
		}
	}

	return samples, sampleRate, bitDepth, channels, nil
}

// reconstructAudioData rebuilds audio data from modified samples.
func (t *LSBAudioTechnique) reconstructAudioData(samples [][]int16, sampleRate, bitDepth, channels int) ([]byte, error) {
	if len(samples) == 0 || len(samples[0]) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	numSamples := len(samples[0])

	// Calculate total size: header + audio data
	headerSize := 44
	audioDataSize := numSamples * channels * (bitDepth / 8)
	totalSize := headerSize + audioDataSize

	result := make([]byte, totalSize)

	// Simple WAV header (44 bytes)
	copy(result[0:4], "RIFF")
	// File size - 8 bytes
	fileSize := uint32(totalSize - 8)
	result[4] = byte(fileSize)
	result[5] = byte(fileSize >> 8)
	result[6] = byte(fileSize >> 16)
	result[7] = byte(fileSize >> 24)

	copy(result[8:12], "WAVE")
	copy(result[12:16], "fmt ")

	// fmt chunk size (16 for PCM)
	result[16] = 16
	result[17] = 0
	result[18] = 0
	result[19] = 0

	// Audio format (1 = PCM)
	result[20] = 1
	result[21] = 0

	// Number of channels
	result[22] = byte(channels)
	result[23] = 0

	// Sample rate
	result[24] = byte(sampleRate)
	result[25] = byte(sampleRate >> 8)
	result[26] = byte(sampleRate >> 16)
	result[27] = byte(sampleRate >> 24)

	// Byte rate
	byteRate := sampleRate * channels * (bitDepth / 8)
	result[28] = byte(byteRate)
	result[29] = byte(byteRate >> 8)
	result[30] = byte(byteRate >> 16)
	result[31] = byte(byteRate >> 24)

	// Block align
	blockAlign := channels * (bitDepth / 8)
	result[32] = byte(blockAlign)
	result[33] = 0

	// Bits per sample
	result[34] = byte(bitDepth)
	result[35] = 0

	// Data chunk
	copy(result[36:40], "data")

	// Data chunk size
	result[40] = byte(audioDataSize)
	result[41] = byte(audioDataSize >> 8)
	result[42] = byte(audioDataSize >> 16)
	result[43] = byte(audioDataSize >> 24)

	// Write interleaved sample data
	offset := headerSize
	for i := 0; i < numSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			if ch < len(samples) {
				sample := samples[ch][i]
				result[offset] = byte(sample)
				result[offset+1] = byte(sample >> 8)
				offset += 2
			}
		}
	}

	return result, nil
}

// embedBitsInSamples embeds payload bits into audio samples using LSB modification.
func (t *LSBAudioTechnique) embedBitsInSamples(samples [][]int16, payloadBits []byte, channels int) ([][]int16, error) {
	if len(samples) == 0 || len(samples[0]) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	numSamples := len(samples[0])
	modifiedSamples := make([][]int16, channels)
	for ch := 0; ch < channels; ch++ {
		modifiedSamples[ch] = make([]int16, numSamples)
		copy(modifiedSamples[ch], samples[ch])
	}

	bitIndex := 0
	bitsEmbedded := 0

	for i := 0; i < numSamples && bitIndex < len(payloadBits); i += t.config.SampleInterval {
		for ch := 0; ch < channels && bitIndex < len(payloadBits); ch++ {
			// Check if this channel is enabled
			if (t.config.ChannelMask & (1 << ch)) == 0 {
				continue
			}

			// Check amplitude threshold if enabled
			if t.config.PreserveSilence {
				amplitude := float64(abs(samples[ch][i])) / 32767.0
				if amplitude < t.config.AmplitudeThreshold {
					continue
				}
			}

			// Embed bits in LSBs
			for bit := 0; bit < t.config.BitsPerSample && bitIndex < len(payloadBits); bit++ {
				payloadBit := payloadBits[bitIndex]

				// Clear LSB and set to payload bit
				mask := uint16(1 << bit)
				modifiedSamples[ch][i] &= ^int16(mask) // Clear bit
				if payloadBit == 1 {
					modifiedSamples[ch][i] |= int16(mask) // Set bit
				}

				bitIndex++
				bitsEmbedded++
			}
		}
	}

	t.logger.Debug("LSB audio embedding progress",
		slog.Int("bits_embedded", bitsEmbedded),
		slog.Int("total_bits", len(payloadBits)))

	return modifiedSamples, nil
}

// extractBitsFromSamples extracts embedded bits from audio samples.
func (t *LSBAudioTechnique) extractBitsFromSamples(samples [][]int16, channels int) ([]byte, error) {
	if len(samples) == 0 || len(samples[0]) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	numSamples := len(samples[0])
	var extractedBits []byte

	for i := 0; i < numSamples; i += t.config.SampleInterval {
		for ch := 0; ch < channels; ch++ {
			// Check if this channel is enabled
			if (t.config.ChannelMask & (1 << ch)) == 0 {
				continue
			}

			// Check amplitude threshold if enabled
			if t.config.PreserveSilence {
				amplitude := float64(abs(samples[ch][i])) / 32767.0
				if amplitude < t.config.AmplitudeThreshold {
					continue
				}
			}

			// Extract bits from LSBs
			for bit := 0; bit < t.config.BitsPerSample; bit++ {
				mask := uint16(1 << bit)
				if (samples[ch][i] & int16(mask)) != 0 {
					extractedBits = append(extractedBits, 1)
				} else {
					extractedBits = append(extractedBits, 0)
				}
			}
		}
	}

	return extractedBits, nil
}

// calculateSampleCapacity calculates embedding capacity based on sample count.
func (t *LSBAudioTechnique) calculateSampleCapacity(samples [][]int16, channels int) int {
	if len(samples) == 0 || len(samples[0]) == 0 {
		return 0
	}

	numSamples := len(samples[0])
	usableChannels := 0

	// Count enabled channels
	for ch := 0; ch < channels; ch++ {
		if (t.config.ChannelMask & (1 << ch)) != 0 {
			usableChannels++
		}
	}

	if usableChannels == 0 {
		return 0
	}

	// Calculate usable samples considering interval
	usableSamples := numSamples / t.config.SampleInterval

	// Apply quality level scaling
	qualityFactor := float64(t.config.QualityLevel) / 5.0

	// Total capacity in bits
	capacityBits := int(float64(usableSamples*usableChannels*t.config.BitsPerSample) * qualityFactor)

	// Convert to bytes
	capacityBytes := capacityBits / 8

	return capacityBytes
}

// Helper functions

// payloadToBits converts byte payload to individual bits.
func (t *LSBAudioTechnique) payloadToBits(payload []byte) []byte {
	bits := make([]byte, len(payload)*8)
	for i, b := range payload {
		for j := 0; j < 8; j++ {
			bit := (b >> (7 - j)) & 1
			bits[i*8+j] = bit
		}
	}
	return bits
}

// bitsToPayload converts individual bits back to byte payload.
func (t *LSBAudioTechnique) bitsToPayload(bits []byte) []byte {
	if len(bits) == 0 {
		return []byte{}
	}

	// Pad to byte boundary if needed
	for len(bits)%8 != 0 {
		bits = append(bits, 0)
	}

	payload := make([]byte, len(bits)/8)
	for i := 0; i < len(payload); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			bit := bits[i*8+j]
			b |= (bit << (7 - j))
		}
		payload[i] = b
	}
	return payload
}

// abs returns the absolute value of an int16.
func abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}

// Compile-time interface compliance check
var _ stego.Technique = (*LSBAudioTechnique)(nil)
