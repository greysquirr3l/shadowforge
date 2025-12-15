// Package stego provides steganography technique implementations.
//
// This file implements phase encoding steganography for audio files,
// which modifies the phase spectrum while preserving the magnitude.
package stego

import (
	"context"
	"fmt"
	"math"
	"math/cmplx"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	mediaInfra "github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/dsp/fourier"
)

// PhaseEmbeddingConfig configures phase encoding parameters.
type PhaseEmbeddingConfig struct {
	SegmentSize    int     // Size of FFT segments (default: 1024)
	PhaseThreshold float64 // Phase modification threshold (default: π/8)
	MinFrequency   int     // Minimum frequency bin to modify (default: 10)
	MaxFrequency   int     // Maximum frequency bin to modify (default: 512)
	OverlapFactor  float64 // Overlap between segments (default: 0.5)
}

// DefaultPhaseConfig returns default phase encoding configuration.
func DefaultPhaseConfig() PhaseEmbeddingConfig {
	return PhaseEmbeddingConfig{
		SegmentSize:    1024,
		PhaseThreshold: math.Pi / 8, // 22.5 degrees
		MinFrequency:   10,
		MaxFrequency:   512,
		OverlapFactor:  0.5,
	}
}

// PhaseTechnique implements audio phase encoding steganography.
//
// This technique embeds data by modifying the phase spectrum of audio segments
// while preserving the magnitude spectrum. The human ear is less sensitive to
// phase changes than magnitude changes, making this technique relatively imperceptible.
type PhaseTechnique struct {
	config PhaseEmbeddingConfig
	logger *logrus.Logger
}

// NewPhaseTechnique creates a new phase encoding technique.
func NewPhaseTechnique(config PhaseEmbeddingConfig, logger *logrus.Logger) *PhaseTechnique {
	if logger == nil {
		logger = logrus.New()
	}

	return &PhaseTechnique{
		config: config,
		logger: logger,
	}
}

// NewPhaseWithDefaults creates a phase technique with default configuration.
func NewPhaseWithDefaults(logger *logrus.Logger) *PhaseTechnique {
	return NewPhaseTechnique(DefaultPhaseConfig(), logger)
}

// Name returns the technique identifier.
func (p *PhaseTechnique) Name() string {
	return "phase"
}

// SupportsFormat checks if this technique can be applied to the given format.
func (p *PhaseTechnique) SupportsFormat(format media.MediaFormat) bool {
	switch format {
	case media.FormatWAV:
		return true
	default:
		return false
	}
}

// Embed hides payload data within audio using phase encoding.
func (p *PhaseTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
	p.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
		"payload_size": len(payload),
	}).Info("Phase embedding started")

	if len(payload) == 0 {
		return nil, stego.ErrEmptyPayload
	}

	// Load WAV file
	audioProc := mediaInfra.NewAudioProcessor()
	pcmData, err := audioProc.LoadWAV(carrier)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	// Convert payload to bits
	payloadBits := make([]bool, len(payload)*8)
	for i, b := range payload {
		for bit := 0; bit < 8; bit++ {
			payloadBits[i*8+bit] = (b>>(7-bit))&1 == 1
		}
	}

	// Add length header (32 bits for payload length)
	lengthBits := make([]bool, 32)
	payloadLen := uint32(len(payload))
	for i := 0; i < 32; i++ {
		lengthBits[i] = (payloadLen>>(31-i))&1 == 1
	}
	allBits := append(lengthBits, payloadBits...)

	// Work with first channel only (mono embedding)
	samples := pcmData.Samples[0]

	// Convert int32 samples to float64 for segmentation
	floatSamples := make([]float64, len(samples))
	for i, sample := range samples {
		floatSamples[i] = float64(sample) / 32768.0
	}

	// Segment audio with overlap
	segments := p.segmentAudio(floatSamples)

	// Embed bits in phase spectrum
	bitIndex := 0
	fft := fourier.NewFFT(p.config.SegmentSize)

	for i := 0; i < len(segments) && bitIndex < len(allBits); i++ {
		segment := segments[i]

		// Apply FFT (input is already float64)
		freqDomain := fft.Coefficients(nil, segment)

		// Modify phase in usable frequency range
		for freq := p.config.MinFrequency; freq < p.config.MaxFrequency && bitIndex < len(allBits); freq++ {
			if freq >= len(freqDomain)/2 {
				break
			}

			magnitude := cmplx.Abs(freqDomain[freq])
			phase := cmplx.Phase(freqDomain[freq])

			// Embed bit by modifying phase
			if allBits[bitIndex] {
				phase += p.config.PhaseThreshold
			} else {
				phase -= p.config.PhaseThreshold
			}

			// Reconstruct complex number with new phase
			freqDomain[freq] = cmplx.Rect(magnitude, phase)

			// Mirror for symmetry (maintain real signal)
			if freq > 0 && freq < len(freqDomain)/2 {
				freqDomain[len(freqDomain)-freq] = cmplx.Conj(freqDomain[freq])
			}

			bitIndex++
		}

		// Apply IFFT
		modifiedSegment := fft.Sequence(nil, freqDomain)

		// Convert back to int32 and update samples
		for j := 0; j < len(modifiedSegment) && i*p.config.SegmentSize+j < len(samples); j++ {
			// Denormalize and clamp (modifiedSegment is already float64)
			value := modifiedSegment[j] * 32768.0
			if value > 32767 {
				value = 32767
			} else if value < -32768 {
				value = -32768
			}
			samples[i*p.config.SegmentSize+j] = int32(value)
		}
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

	p.logger.WithFields(logrus.Fields{
		"result_size":   len(result),
		"bits_embedded": bitIndex,
	}).Info("Phase embedding completed")

	return result, nil
}

// Extract retrieves hidden data from audio using phase decoding.
func (p *PhaseTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
	p.logger.WithFields(logrus.Fields{
		"carrier_size": len(carrier),
	}).Info("Phase extraction started")

	// Load WAV file
	audioProc := mediaInfra.NewAudioProcessor()
	pcmData, err := audioProc.LoadWAV(carrier)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	// Work with first channel only
	samples := pcmData.Samples[0]

	// Convert int32 samples to float64 for segmentation
	floatSamples := make([]float64, len(samples))
	for i, sample := range samples {
		floatSamples[i] = float64(sample) / 32768.0
	}

	// Segment audio
	segments := p.segmentAudio(floatSamples)

	// Extract bits from phase spectrum
	var extractedBits []bool
	fft := fourier.NewFFT(p.config.SegmentSize)

	for i := 0; i < len(segments); i++ {
		segment := segments[i]

		// Apply FFT (input is already float64)
		freqDomain := fft.Coefficients(nil, segment)

		// Extract bits from phase in usable frequency range
		for freq := p.config.MinFrequency; freq < p.config.MaxFrequency; freq++ {
			if freq >= len(freqDomain)/2 {
				break
			}

			phase := cmplx.Phase(freqDomain[freq])

			// Determine bit based on phase modification
			// Positive phase shift = 1, negative = 0
			bit := phase > 0
			extractedBits = append(extractedBits, bit)
		}
	}

	// First 32 bits are the length
	if len(extractedBits) < 32 {
		return nil, fmt.Errorf("insufficient data: need at least 32 bits for length header")
	}

	// Extract payload length
	var payloadLen uint32
	for i := 0; i < 32; i++ {
		if extractedBits[i] {
			payloadLen |= 1 << (31 - i)
		}
	}

	if payloadLen == 0 || payloadLen > 1024*1024 { // Sanity check (max 1MB)
		return nil, fmt.Errorf("invalid payload length: %d", payloadLen)
	}

	// Extract payload bits
	payloadBitCount := int(payloadLen) * 8
	if len(extractedBits) < 32+payloadBitCount {
		return nil, fmt.Errorf("insufficient data: need %d bits, have %d", 32+payloadBitCount, len(extractedBits))
	}

	// Convert bits to bytes
	payload := make([]byte, payloadLen)
	for i := 0; i < int(payloadLen); i++ {
		var b byte
		for bit := 0; bit < 8; bit++ {
			if extractedBits[32+i*8+bit] {
				b |= 1 << (7 - bit)
			}
		}
		payload[i] = b
	}

	p.logger.WithFields(logrus.Fields{
		"extracted_size": len(payload),
		"bits_extracted": len(extractedBits),
	}).Info("Phase extraction completed")

	return payload, nil
}

// CalculateCapacity determines embedding capacity for phase encoding.
func (p *PhaseTechnique) CalculateCapacity(ctx context.Context, carrier []byte) (int, error) {
	if len(carrier) < 1024 {
		return 0, stego.ErrInsufficientCapacity
	}

	// TODO: Calculate actual capacity based on audio parameters
	// Full implementation would:
	// 1. Parse audio format (sample rate, channels, bit depth)
	// 2. Calculate number of frequency bins available for modification
	// 3. Account for perceptual constraints
	// 4. Return bits available per audio segment

	// For simplified implementation, assume 1 bit per 8 bytes of audio
	capacity := len(carrier) / 8

	p.logger.WithFields(logrus.Fields{
		"carrier_bytes":  len(carrier),
		"capacity_bytes": capacity,
	}).Debug("Calculated phase embedding capacity")

	return capacity, nil
}

// Private helper functions for full implementation

// segmentAudio divides audio into overlapping segments for FFT processing.
func (p *PhaseTechnique) segmentAudio(samples []float64) [][]float64 {
	segmentSize := p.config.SegmentSize
	overlap := int(float64(segmentSize) * p.config.OverlapFactor)
	step := segmentSize - overlap

	var segments [][]float64
	for i := 0; i+segmentSize <= len(samples); i += step {
		segment := make([]float64, segmentSize)
		copy(segment, samples[i:i+segmentSize])
		segments = append(segments, segment)
	}

	return segments
}

// embedBitsInPhase modifies phase spectrum to embed bits.
func (p *PhaseTechnique) embedBitsInPhase(spectrum []complex128, bits []byte, bitIndex *int) error {
	minFreq := p.config.MinFrequency
	maxFreq := p.config.MaxFrequency
	if maxFreq > len(spectrum)/2 {
		maxFreq = len(spectrum) / 2
	}

	for freq := minFreq; freq < maxFreq && *bitIndex < len(bits)*8; freq++ {
		if *bitIndex >= len(bits)*8 {
			break
		}

		// Get bit to embed
		byteIdx := *bitIndex / 8
		bitPos := *bitIndex % 8
		bit := (bits[byteIdx] >> (7 - bitPos)) & 1

		// Modify phase based on bit value
		magnitude := cmplx.Abs(spectrum[freq])
		currentPhase := cmplx.Phase(spectrum[freq])

		var newPhase float64
		if bit == 1 {
			newPhase = currentPhase + p.config.PhaseThreshold
		} else {
			newPhase = currentPhase - p.config.PhaseThreshold
		}

		// Reconstruct complex number with new phase
		spectrum[freq] = cmplx.Rect(magnitude, newPhase)

		// Mirror for conjugate symmetry in real FFT
		if freq < len(spectrum)/2 {
			spectrum[len(spectrum)-freq] = cmplx.Conj(spectrum[freq])
		}

		*bitIndex++
	}

	return nil
}

// extractBitsFromPhase extracts bits from phase spectrum modifications.
func (p *PhaseTechnique) extractBitsFromPhase(spectrum []complex128, expectedBits int) ([]byte, error) {
	minFreq := p.config.MinFrequency
	maxFreq := p.config.MaxFrequency
	if maxFreq > len(spectrum)/2 {
		maxFreq = len(spectrum) / 2
	}

	bits := make([]byte, (expectedBits+7)/8) // Round up to byte boundary
	bitIndex := 0

	for freq := minFreq; freq < maxFreq && bitIndex < expectedBits; freq++ {
		phase := cmplx.Phase(spectrum[freq])

		// Determine bit based on phase value
		// This is a simplified detection algorithm
		var bit byte
		if phase > 0 {
			bit = 1
		} else {
			bit = 0
		}

		// Set bit in result
		byteIdx := bitIndex / 8
		bitPos := bitIndex % 8
		if bit == 1 {
			bits[byteIdx] |= (1 << (7 - bitPos))
		}

		bitIndex++
	}

	return bits[:expectedBits/8], nil
}

// Compile-time interface compliance check
var _ stego.Technique = (*PhaseTechnique)(nil)
