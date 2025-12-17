// Package stego provides steganography technique implementations.
//
// This file implements echo hiding steganography for audio files,
// which embeds data by introducing imperceptible echoes.
package stego

import (
	"context"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	mediaInfra "github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	"github.com/sirupsen/logrus"
)

// EchoEmbeddingConfig configures echo hiding parameters.
type EchoEmbeddingConfig struct {
	Delay0     int     // Echo delay for bit 0 in samples (default: 100)
	Delay1     int     // Echo delay for bit 1 in samples (default: 101)
	Amplitude  float64 // Echo amplitude factor (default: 0.1)
	SegmentLen int     // Segment length in samples (default: 8192)
	MixRatio   float64 // Ratio of echo to original signal (default: 0.8)
}

// DefaultEchoConfig returns default echo hiding configuration.
// Echo delays must be significantly different for reliable autocorrelation detection.
// At 44.1kHz: 150 samples = 3.4ms, 500 samples = 11.3ms (7.9ms difference)
// Amplitude and MixRatio tuned for reliable autocorrelation detection on white noise.
func DefaultEchoConfig() EchoEmbeddingConfig {
	return EchoEmbeddingConfig{
		Delay0:     150, // ~3.4ms at 44.1kHz - short echo for bit 0
		Delay1:     500, // ~11.3ms at 44.1kHz - long echo for bit 1
		Amplitude:  0.5, // Echo strength (0.5 gives ~0.02 autocorrelation peak)
		SegmentLen: 8192,
		MixRatio:   0.5, // 50/50 mix of original and echo
	}
}

// EchoTechnique implements audio echo hiding steganography.
//
// This technique embeds data by adding imperceptible echoes to audio segments.
// Different echo delays represent binary 0 and 1. The human ear typically
// cannot distinguish between such small delay differences.
type EchoTechnique struct {
	config EchoEmbeddingConfig
	logger *logrus.Logger
}

// NewEchoTechnique creates a new echo hiding technique.
func NewEchoTechnique(config EchoEmbeddingConfig, logger *logrus.Logger) *EchoTechnique {
	if logger == nil {
		logger = logrus.New()
	}

	return &EchoTechnique{
		config: config,
		logger: logger,
	}
}

// NewEchoWithDefaults creates an echo technique with default configuration.
func NewEchoWithDefaults(logger *logrus.Logger) *EchoTechnique {
	return NewEchoTechnique(DefaultEchoConfig(), logger)
}

// Name returns the technique identifier.
func (e *EchoTechnique) Name() string {
	return "echo"
}

// SupportsFormat checks if this technique can be applied to the given format.
func (e *EchoTechnique) SupportsFormat(format media.MediaFormat) bool {
	switch format {
	case media.FormatWAV:
		return true
	default:
		return false
	}
}

// Embed hides payload data within audio using echo hiding.
func (e *EchoTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
	e.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
		"payload_size": len(payload),
	}).Info("Echo embedding started")

	if len(payload) == 0 {
		return nil, stego.ErrEmptyPayload
	}

	// Load WAV file
	audioProc := mediaInfra.NewAudioProcessor()
	pcmData, err := audioProc.LoadWAV(carrier)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	// Convert payload to bits (with 32-bit length header)
	payloadLen := uint32(len(payload))
	lengthBits := make([]bool, 32)
	for i := 0; i < 32; i++ {
		lengthBits[i] = (payloadLen>>(31-i))&1 == 1
	}

	payloadBits := e.payloadToBits(payload)
	allBits := append(lengthBits, payloadBits...)

	// Work with first channel
	samples := pcmData.Samples[0]

	// Check capacity
	maxSegments := len(samples) / e.config.SegmentLen
	if maxSegments < len(allBits) {
		return nil, stego.ErrInsufficientCapacity
	}

	// Segment audio
	segments := e.segmentAudio(samples, e.config.SegmentLen)

	// Process each segment and embed one bit per segment
	bitIndex := 0
	for i := 0; i < len(segments) && bitIndex < len(allBits); i++ {
		segment := segments[i]

		// Convert int32 to float64 for processing
		floatSegment := make([]float64, len(segment))
		for j, s := range segment {
			floatSegment[j] = float64(s) / 32768.0
		}

		// Add echo based on bit value
		var delay int
		if allBits[bitIndex] {
			delay = e.config.Delay1
		} else {
			delay = e.config.Delay0
		}

		modifiedSegment := e.addEcho(floatSegment, delay, e.config.Amplitude, e.config.MixRatio)

		// Convert back to int32 and write to samples
		for j := 0; j < len(modifiedSegment) && i*e.config.SegmentLen+j < len(samples); j++ {
			value := modifiedSegment[j] * 32768.0
			if value > 32767 {
				value = 32767
			} else if value < -32768 {
				value = -32768
			}
			samples[i*e.config.SegmentLen+j] = int32(value)
		}

		bitIndex++
	}

	if bitIndex < len(allBits) {
		return nil, stego.ErrInsufficientCapacity
	}

	// Reconstruct WAV
	pcmData.Samples[0] = samples
	result, err := audioProc.SaveWAV(pcmData)
	if err != nil {
		return nil, fmt.Errorf("failed to save WAV: %w", err)
	}

	e.logger.WithFields(logrus.Fields{
		"result_size":   len(result),
		"bits_embedded": bitIndex,
		"delay0":        e.config.Delay0,
		"delay1":        e.config.Delay1,
	}).Info("Echo embedding completed")

	return result, nil
}

// Extract retrieves hidden data from audio using echo detection.
func (e *EchoTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	e.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
	}).Info("Echo extraction started")

	if len(carrier) == 0 {
		return nil, stego.ErrEmptyCoverMedia
	}

	// Load WAV file
	audioProc := mediaInfra.NewAudioProcessor()
	pcmData, err := audioProc.LoadWAV(carrier)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	// Get first channel
	samples := pcmData.Samples[0]

	// Segment audio
	segments := e.segmentAudio(samples, e.config.SegmentLen)

	// Extract bits from each segment
	var extractedBits []bool
	for _, segment := range segments {
		// Convert int32 to float64
		floatSegment := make([]float64, len(segment))
		for j, s := range segment {
			floatSegment[j] = float64(s) / 32768.0
		}

		// Detect echo delay
		delay, err := e.detectEcho(floatSegment)
		if err != nil {
			// If detection fails, assume bit 0
			extractedBits = append(extractedBits, false)
			continue
		}

		// Classify delay as bit 0 or bit 1
		// Delay closer to Delay0 → bit 0
		// Delay closer to Delay1 → bit 1
		diff0 := absInt(delay - e.config.Delay0)
		diff1 := absInt(delay - e.config.Delay1)

		bit := diff1 < diff0
		extractedBits = append(extractedBits, bit)
	}

	// First 32 bits are payload length
	if len(extractedBits) < 32 {
		return nil, stego.ErrCorruptedContainer
	}

	var payloadLen uint32
	for i := 0; i < 32; i++ {
		if extractedBits[i] {
			payloadLen |= 1 << (31 - i)
		}
	}

	// Extract payload bits
	totalBits := 32 + int(payloadLen)*8
	if len(extractedBits) < totalBits {
		return nil, stego.ErrCorruptedContainer
	}

	payloadBits := extractedBits[32:totalBits]
	payload := e.bitsToPayload(payloadBits)

	e.logger.WithFields(logrus.Fields{
		"extracted_size": len(payload),
		"total_bits":     len(extractedBits),
		"payload_length": payloadLen,
	}).Info("Echo extraction completed")

	return payload, nil
}

// Helper function for absolute value of integers
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// CalculateCapacity determines embedding capacity for echo hiding.
func (e *EchoTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	if len(carrier) < e.config.SegmentLen {
		return 0, stego.ErrInsufficientCapacity
	}

	// TODO: Calculate actual capacity based on audio parameters
	// Full implementation would:
	// 1. Parse audio format (sample rate, channels, bit depth)
	// 2. Calculate number of audio segments available
	// 3. Account for echo delay requirements
	// 4. Return bits available (typically 1 bit per segment)

	// For simplified implementation, assume 1 bit per segment
	numSegments := len(carrier) / e.config.SegmentLen
	capacityBits := numSegments
	capacityBytes := capacityBits / 8

	e.logger.WithFields(logrus.Fields{
		"carrier_bytes":  len(carrier),
		"num_segments":   numSegments,
		"capacity_bits":  capacityBits,
		"capacity_bytes": capacityBytes,
	}).Debug("Calculated echo embedding capacity")

	return capacityBytes, nil
}

// Private helper functions for full implementation

// addEcho adds echo to audio samples based on delay value.
func (e *EchoTechnique) addEcho(samples []float64, delay int, amplitude float64, mixRatio float64) []float64 {
	if delay >= len(samples) {
		return samples
	}

	result := make([]float64, len(samples))
	copy(result, samples)

	// Add echo
	for i := delay; i < len(result); i++ {
		echoSample := samples[i-delay] * amplitude
		result[i] = samples[i]*mixRatio + echoSample*(1-mixRatio)
	}

	return result
}

// detectEcho detects echo delays using autocorrelation.
func (e *EchoTechnique) detectEcho(samples []float64) (int, error) {
	// Look for autocorrelation peaks at expected delays
	corr0 := e.calculateAutocorrelation(samples, e.config.Delay0)
	corr1 := e.calculateAutocorrelation(samples, e.config.Delay1)

	// Return delay with stronger correlation
	if corr0 > corr1 {
		return e.config.Delay0, nil
	}
	return e.config.Delay1, nil
}

// calculateAutocorrelation calculates autocorrelation at specific lag.
func (e *EchoTechnique) calculateAutocorrelation(samples []float64, lag int) float64 {
	if lag >= len(samples) {
		return 0.0
	}

	var correlation float64
	var count int

	for i := lag; i < len(samples); i++ {
		correlation += samples[i] * samples[i-lag]
		count++
	}

	if count == 0 {
		return 0.0
	}

	return correlation / float64(count)
}

// segmentAudio divides audio into segments for processing.
func (e *EchoTechnique) segmentAudio(samples []int32, segmentLen int) [][]int32 {
	var segments [][]int32

	for i := 0; i < len(samples); i += segmentLen {
		end := i + segmentLen
		if end > len(samples) {
			end = len(samples)
		}

		segment := make([]int32, end-i)
		copy(segment, samples[i:end])
		segments = append(segments, segment)
	}

	return segments
}

// payloadToBits converts byte payload to individual bits.
func (e *EchoTechnique) payloadToBits(payload []byte) []bool {
	bits := make([]bool, len(payload)*8)
	for i, b := range payload {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b>>(7-j))&1 == 1
		}
	}
	return bits
}

// bitsToPayload converts individual bits back to byte payload.
func (e *EchoTechnique) bitsToPayload(bits []bool) []byte {
	numBytes := len(bits) / 8
	payload := make([]byte, numBytes)

	for i := 0; i < numBytes; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << (7 - j)
			}
		}
		payload[i] = b
	}

	return payload
}

// validateConfig validates the echo hiding configuration.
func (e *EchoTechnique) validateConfig() error {
	if e.config.Delay0 <= 0 || e.config.Delay1 <= 0 {
		return fmt.Errorf("invalid echo delays: delay0=%d, delay1=%d",
			e.config.Delay0, e.config.Delay1)
	}

	if e.config.Amplitude <= 0 || e.config.Amplitude > 1 {
		return fmt.Errorf("invalid echo amplitude: %f (must be 0 < amplitude <= 1)",
			e.config.Amplitude)
	}

	if e.config.SegmentLen <= e.config.Delay1 {
		return fmt.Errorf("segment length (%d) must be greater than max delay (%d)",
			e.config.SegmentLen, e.config.Delay1)
	}

	if e.config.MixRatio < 0 || e.config.MixRatio > 1 {
		return fmt.Errorf("invalid mix ratio: %f (must be 0 <= ratio <= 1)",
			e.config.MixRatio)
	}

	return nil
}

// Compile-time interface compliance check
var _ stego.Technique = (*EchoTechnique)(nil)
