package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// AudioProcessor handles audio file operations.
type AudioProcessor struct {
	detector *FormatDetector
}

// NewAudioProcessor creates a new audio processor.
func NewAudioProcessor() *AudioProcessor {
	return &AudioProcessor{
		detector: NewFormatDetector(),
	}
}

// WAVHeader represents a WAV file header structure.
type WAVHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // File size - 8
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 = mono, 2 = stereo
	SampleRate    uint32  // Sample rate in Hz
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8, 16, 24, 32
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // NumSamples * NumChannels * BitsPerSample/8
}

// PCMData represents decoded PCM audio data.
type PCMData struct {
	SampleRate    int
	BitsPerSample int
	Channels      int
	Samples       [][]int32 // [channel][sample]
}

// LoadWAV reads and decodes a WAV file.
func (p *AudioProcessor) LoadWAV(data []byte) (*PCMData, error) {
	if len(data) < 44 {
		return nil, ErrInsufficientData
	}

	// Verify it's a WAV file
	format, mediaType, err := p.detector.DetectFormat(data)
	if err != nil {
		return nil, fmt.Errorf("format detection failed: %w", err)
	}

	if mediaType != media.MediaTypeAudio || format != media.FormatWAV {
		return nil, ErrUnsupportedFormat
	}

	// Parse WAV header
	reader := bytes.NewReader(data)
	var header WAVHeader

	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to read WAV header: %w", err)
	}

	// Validate header
	if string(header.ChunkID[:]) != "RIFF" {
		return nil, ErrCorruptedHeader
	}
	if string(header.Format[:]) != "WAVE" {
		return nil, ErrCorruptedHeader
	}
	if header.AudioFormat != 1 {
		return nil, fmt.Errorf("unsupported audio format: %d (only PCM supported)", header.AudioFormat)
	}

	// Read PCM data
	pcmData := &PCMData{
		SampleRate:    int(header.SampleRate),
		BitsPerSample: int(header.BitsPerSample),
		Channels:      int(header.NumChannels),
	}

	// Calculate number of samples
	bytesPerSample := int(header.BitsPerSample) / 8
	totalBytes := int(header.Subchunk2Size)
	numSamples := totalBytes / (bytesPerSample * int(header.NumChannels))

	// Initialize sample arrays
	pcmData.Samples = make([][]int32, pcmData.Channels)
	for i := range pcmData.Samples {
		pcmData.Samples[i] = make([]int32, numSamples)
	}

	// Read samples
	for sample := 0; sample < numSamples; sample++ {
		for ch := 0; ch < pcmData.Channels; ch++ {
			var value int32

			switch header.BitsPerSample {
			case 8:
				var v uint8
				if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
					return nil, fmt.Errorf("failed to read 8-bit sample: %w", err)
				}
				// 8-bit samples are unsigned, convert to signed
				value = int32(v) - 128

			case 16:
				var v int16
				if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
					return nil, fmt.Errorf("failed to read 16-bit sample: %w", err)
				}
				value = int32(v)

			case 24:
				// 24-bit samples need special handling
				var bytes [3]byte
				if _, err := io.ReadFull(reader, bytes[:]); err != nil {
					return nil, fmt.Errorf("failed to read 24-bit sample: %w", err)
				}
				// Convert 24-bit to 32-bit with sign extension
				value = int32(bytes[0]) | int32(bytes[1])<<8 | int32(bytes[2])<<16
				if value&0x800000 != 0 {
					value |= ^0x00FFFFFF // Sign extend
				}

			case 32:
				var v int32
				if err := binary.Read(reader, binary.LittleEndian, &v); err != nil {
					return nil, fmt.Errorf("failed to read 32-bit sample: %w", err)
				}
				value = v

			default:
				return nil, fmt.Errorf("unsupported bit depth: %d", header.BitsPerSample)
			}

			pcmData.Samples[ch][sample] = value
		}
	}

	return pcmData, nil
}

// SaveWAV encodes and writes a WAV file.
func (p *AudioProcessor) SaveWAV(pcm *PCMData) ([]byte, error) {
	if pcm == nil {
		return nil, fmt.Errorf("PCM data is nil")
	}

	if len(pcm.Samples) == 0 {
		return nil, fmt.Errorf("no audio channels")
	}

	if len(pcm.Samples[0]) == 0 {
		return nil, fmt.Errorf("no audio samples")
	}

	// Validate bit depth
	if pcm.BitsPerSample != 8 && pcm.BitsPerSample != 16 && pcm.BitsPerSample != 24 && pcm.BitsPerSample != 32 {
		return nil, fmt.Errorf("unsupported bit depth: %d", pcm.BitsPerSample)
	}

	numSamples := len(pcm.Samples[0])
	bytesPerSample := pcm.BitsPerSample / 8
	dataSize := uint32(numSamples * pcm.Channels * bytesPerSample)

	// Build WAV header
	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16, // PCM
		AudioFormat:   1,  // PCM
		NumChannels:   uint16(pcm.Channels),
		SampleRate:    uint32(pcm.SampleRate),
		ByteRate:      uint32(pcm.SampleRate * pcm.Channels * bytesPerSample),
		BlockAlign:    uint16(pcm.Channels * bytesPerSample),
		BitsPerSample: uint16(pcm.BitsPerSample),
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	var buf bytes.Buffer

	// Write header
	if err := binary.Write(&buf, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Write samples
	for sample := 0; sample < numSamples; sample++ {
		for ch := 0; ch < pcm.Channels; ch++ {
			value := pcm.Samples[ch][sample]

			switch pcm.BitsPerSample {
			case 8:
				// Convert signed to unsigned for 8-bit
				v := uint8(value + 128)
				if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
					return nil, fmt.Errorf("failed to write 8-bit sample: %w", err)
				}

			case 16:
				v := int16(value)
				if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
					return nil, fmt.Errorf("failed to write 16-bit sample: %w", err)
				}

			case 24:
				// Write 24-bit sample as 3 bytes
				bytes := [3]byte{
					byte(value),
					byte(value >> 8),
					byte(value >> 16),
				}
				if _, err := buf.Write(bytes[:]); err != nil {
					return nil, fmt.Errorf("failed to write 24-bit sample: %w", err)
				}

			case 32:
				if err := binary.Write(&buf, binary.LittleEndian, value); err != nil {
					return nil, fmt.Errorf("failed to write 32-bit sample: %w", err)
				}
			}
		}
	}

	return buf.Bytes(), nil
}

// GetAudioInfo extracts audio metadata without fully decoding.
func (p *AudioProcessor) GetAudioInfo(data []byte) (*AudioInfo, error) {
	return p.detector.ParseAudioInfo(data)
}

// CalculateCapacity estimates the steganographic capacity for audio.
func (p *AudioProcessor) CalculateCapacity(data []byte, technique string) (*media.CapacityInfo, error) {
	// Get audio info
	info, err := p.detector.ParseAudioInfo(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audio info: %w", err)
	}

	// Only WAV is fully supported for capacity calculation
	if info.Format != media.FormatWAV {
		return nil, fmt.Errorf("capacity calculation only supported for WAV format")
	}

	// Load PCM data to get exact sample count
	pcm, err := p.LoadWAV(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load WAV: %w", err)
	}

	totalSamples := int64(len(pcm.Samples[0]))
	var rawCapacity, safeCapacity int64

	switch technique {
	case "lsb":
		// LSB audio: 1 bit per sample
		rawCapacity = totalSamples / 8 // Convert to bytes

		// Safe capacity: Use 80% of samples to avoid detectability
		safeCapacity = int64(float64(rawCapacity) * 0.80)

	case "lsb-2":
		// 2-bit LSB: 2 bits per sample
		rawCapacity = (totalSamples * 2) / 8
		safeCapacity = int64(float64(rawCapacity) * 0.70)

	case "phase":
		// Phase encoding: Embed in phase of frequency components
		// Estimate: 1 bit per 1024-sample segment
		segments := totalSamples / 1024
		rawCapacity = segments / 8

		// Safe capacity: Use fewer segments
		safeCapacity = int64(float64(rawCapacity) * 0.60)

	case "echo":
		// Echo hiding: Embed by introducing imperceptible echoes
		// Estimate: 1 bit per 8192 samples (conservative)
		rawCapacity = totalSamples / (8192 * 8)

		// Safe capacity: Even more conservative
		safeCapacity = int64(float64(rawCapacity) * 0.50)

	default:
		return nil, fmt.Errorf("unknown technique: %s", technique)
	}

	// Calculate quality score
	qualityScore := p.calculateQualityScore(info, totalSamples)

	return &media.CapacityInfo{
		AssetID:        media.AssetID{}, // Empty, to be set by caller
		MediaType:      media.MediaTypeAudio,
		TotalCapacity:  rawCapacity,
		UsableCapacity: safeCapacity,
		RecommendedMax: safeCapacity,
		QualityImpact:  1.0 - qualityScore,
	}, nil
}

// calculateQualityScore assesses audio quality for steganography (0.0 - 1.0).
func (p *AudioProcessor) calculateQualityScore(info *AudioInfo, totalSamples int64) float64 {
	score := 0.5 // Base score

	// Higher sample rates are better
	if info.SampleRate >= 48000 {
		score += 0.2
	} else if info.SampleRate >= 44100 {
		score += 0.1
	} else if info.SampleRate < 16000 {
		score -= 0.1
	}

	// Higher bit depth is better
	if info.BitDepth >= 24 {
		score += 0.2
	} else if info.BitDepth >= 16 {
		score += 0.1
	} else if info.BitDepth <= 8 {
		score -= 0.2
	}

	// Stereo is slightly better than mono (more capacity)
	if info.Channels >= 2 {
		score += 0.05
	}

	// Longer audio is better
	durationSeconds := float64(totalSamples) / float64(info.SampleRate)
	if durationSeconds > 60 {
		score += 0.1
	} else if durationSeconds < 5 {
		score -= 0.1
	}

	// Clamp to [0.0, 1.0]
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// ValidateAudio checks if audio data is valid and can be processed.
func (p *AudioProcessor) ValidateAudio(data []byte) error {
	if len(data) == 0 {
		return ErrInsufficientData
	}

	// Verify format is supported
	format, mediaType, err := p.detector.DetectFormat(data)
	if err != nil {
		return fmt.Errorf("format detection failed: %w", err)
	}

	if mediaType != media.MediaTypeAudio {
		return fmt.Errorf("not an audio file")
	}

	// Format-specific validation
	switch format {
	case media.FormatWAV:
		return p.validateWAV(data)
	case media.FormatFLAC:
		return p.validateFLAC(data)
	case media.FormatMP3:
		return p.validateMP3(data)
	default:
		return ErrUnsupportedFormat
	}
}

// validateWAV performs WAV-specific validation.
func (p *AudioProcessor) validateWAV(data []byte) error {
	if len(data) < 44 {
		return ErrInsufficientData
	}

	// Check RIFF header
	if string(data[0:4]) != "RIFF" {
		return ErrCorruptedHeader
	}

	// Check WAVE format
	if string(data[8:12]) != "WAVE" {
		return ErrCorruptedHeader
	}

	// Check fmt chunk
	if string(data[12:16]) != "fmt " {
		return fmt.Errorf("missing fmt chunk")
	}

	// Verify audio format (should be 1 for PCM)
	audioFormat := binary.LittleEndian.Uint16(data[20:22])
	if audioFormat != 1 {
		return fmt.Errorf("unsupported audio format: %d", audioFormat)
	}

	// Verify we can load it
	_, err := p.LoadWAV(data)
	return err
}

// validateFLAC performs FLAC-specific validation.
func (p *AudioProcessor) validateFLAC(data []byte) error {
	if len(data) < 4 {
		return ErrInsufficientData
	}

	// Check fLaC signature
	if string(data[0:4]) != "fLaC" {
		return ErrCorruptedHeader
	}

	// FLAC is read-only, just verify header
	return nil
}

// validateMP3 performs MP3-specific validation.
func (p *AudioProcessor) validateMP3(data []byte) error {
	if len(data) < 3 {
		return ErrInsufficientData
	}

	// Check for ID3 tag or MP3 sync word
	hasID3 := string(data[0:3]) == "ID3"
	hasMP3Sync := (data[0] == 0xFF && (data[1]&0xE0) == 0xE0)

	if !hasID3 && !hasMP3Sync {
		return ErrCorruptedHeader
	}

	// MP3 is read-only, just verify header
	return nil
}

// CreateSilence creates silent PCM data of specified duration.
func (p *AudioProcessor) CreateSilence(sampleRate, bitDepth, channels int, durationSeconds float64) (*PCMData, error) {
	if sampleRate <= 0 || bitDepth <= 0 || channels <= 0 || durationSeconds <= 0 {
		return nil, fmt.Errorf("invalid audio parameters")
	}

	if bitDepth != 8 && bitDepth != 16 && bitDepth != 24 && bitDepth != 32 {
		return nil, fmt.Errorf("unsupported bit depth: %d", bitDepth)
	}

	numSamples := int(float64(sampleRate) * durationSeconds)

	pcm := &PCMData{
		SampleRate:    sampleRate,
		BitsPerSample: bitDepth,
		Channels:      channels,
		Samples:       make([][]int32, channels),
	}

	// Create silent samples (all zeros)
	for i := 0; i < channels; i++ {
		pcm.Samples[i] = make([]int32, numSamples)
		// All zeros = silence
	}

	return pcm, nil
}

// ModifySamples allows modification of PCM samples (for steganography).
func (p *AudioProcessor) ModifySamples(pcm *PCMData, modifier func(channel, sample int, value int32) int32) error {
	if pcm == nil {
		return fmt.Errorf("PCM data is nil")
	}

	for ch := 0; ch < len(pcm.Samples); ch++ {
		for s := 0; s < len(pcm.Samples[ch]); s++ {
			pcm.Samples[ch][s] = modifier(ch, s, pcm.Samples[ch][s])
		}
	}

	return nil
}
