// Package media provides infrastructure implementations for media processing.
// This includes image, audio, and text processors with format detection.
package media

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

// FormatDetector errors
var (
	ErrUnknownFormat      = errors.New("unknown media format")
	ErrInsufficientData   = errors.New("insufficient data for format detection")
	ErrCorruptedHeader    = errors.New("corrupted file header")
	ErrUnsupportedVariant = errors.New("unsupported format variant")
	ErrUnsupportedFormat  = errors.New("unsupported media format")
)

// Magic bytes for format detection
var (
	// Image formats
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	gifMagic  = []byte("GIF8")
	bmpMagic  = []byte("BM")

	// Audio formats
	riffMagic = []byte("RIFF")
	waveFmt   = []byte("WAVE")
	flacMagic = []byte("fLaC")
	mp3Magic1 = []byte{0xFF, 0xFB}
	mp3Magic2 = []byte{0xFF, 0xFA}
	mp3Magic3 = []byte{0xFF, 0xF3}
	id3Magic  = []byte("ID3")
)

// FormatDetector provides format detection capabilities for media files.
type FormatDetector struct {
	minImageData int // Minimum bytes needed for image format detection
	minAudioData int // Minimum bytes needed for audio format detection
}

// NewFormatDetector creates a new format detector instance.
func NewFormatDetector() *FormatDetector {
	return &FormatDetector{
		minImageData: 12, // Enough for PNG magic + some header
		minAudioData: 12, // Enough for RIFF/WAVE header
	}
}

// DetectFormat automatically detects the media format from raw data.
// Returns the format and media type, or an error if detection fails.
func (f *FormatDetector) DetectFormat(data []byte) (media.MediaFormat, media.MediaType, error) {
	if len(data) < 4 {
		return "", "", ErrInsufficientData
	}

	// Try image formats first
	if format, ok := f.detectImageFormat(data); ok {
		return format, media.MediaTypeImage, nil
	}

	// Try audio formats
	if format, ok := f.detectAudioFormat(data); ok {
		return format, media.MediaTypeAudio, nil
	}

	// Try text formats (heuristic-based)
	if format, ok := f.detectTextFormat(data); ok {
		return format, media.MediaTypeText, nil
	}

	return "", "", ErrUnknownFormat
}

// DetectImageFormat detects image format from data.
func (f *FormatDetector) DetectImageFormat(data []byte) (media.MediaFormat, error) {
	if len(data) < f.minImageData {
		return "", ErrInsufficientData
	}

	if format, ok := f.detectImageFormat(data); ok {
		return format, nil
	}

	return "", ErrUnknownFormat
}

// DetectAudioFormat detects audio format from data.
func (f *FormatDetector) DetectAudioFormat(data []byte) (media.MediaFormat, error) {
	if len(data) < f.minAudioData {
		return "", ErrInsufficientData
	}

	if format, ok := f.detectAudioFormat(data); ok {
		return format, nil
	}

	return "", ErrUnknownFormat
}

// detectImageFormat checks for image format magic bytes.
func (f *FormatDetector) detectImageFormat(data []byte) (media.MediaFormat, bool) {
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && bytes.HasPrefix(data, pngMagic) {
		return media.FormatPNG, true
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && bytes.HasPrefix(data, jpegMagic) {
		return media.FormatJPEG, true
	}

	// GIF: GIF87a or GIF89a
	if len(data) >= 6 && bytes.HasPrefix(data, gifMagic) {
		variant := string(data[4:6])
		if variant == "7a" || variant == "9a" {
			return media.FormatGIF, true
		}
	}

	// BMP: BM
	if len(data) >= 2 && bytes.HasPrefix(data, bmpMagic) {
		return media.FormatBMP, true
	}

	return "", false
}

// detectAudioFormat checks for audio format magic bytes.
func (f *FormatDetector) detectAudioFormat(data []byte) (media.MediaFormat, bool) {
	// WAV: RIFF....WAVE
	if len(data) >= 12 && bytes.HasPrefix(data, riffMagic) {
		if bytes.Equal(data[8:12], waveFmt) {
			return media.FormatWAV, true
		}
	}

	// FLAC: fLaC
	if len(data) >= 4 && bytes.HasPrefix(data, flacMagic) {
		return media.FormatFLAC, true
	}

	// MP3: FF FB, FF FA, FF F3 (frame sync), or ID3 tag
	if len(data) >= 3 {
		if bytes.HasPrefix(data, id3Magic) {
			return media.FormatMP3, true
		}
		if len(data) >= 2 {
			if bytes.HasPrefix(data, mp3Magic1) ||
				bytes.HasPrefix(data, mp3Magic2) ||
				bytes.HasPrefix(data, mp3Magic3) {
				return media.FormatMP3, true
			}
		}
	}

	return "", false
}

// detectTextFormat uses heuristics to detect text format.
func (f *FormatDetector) detectTextFormat(data []byte) (media.MediaFormat, bool) {
	if len(data) == 0 {
		return "", false
	}

	// Check if data is valid UTF-8 text
	if !isValidUTF8Text(data) {
		return "", false
	}

	// Check for Markdown indicators
	if hasMarkdownIndicators(data) {
		return media.FormatMarkdown, true
	}

	// Default to plain text
	return media.FormatTXT, true
}

// isValidUTF8Text checks if data appears to be valid UTF-8 text.
// Returns false if it contains too many non-printable characters.
func isValidUTF8Text(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Sample first 1KB for efficiency
	checkLen := len(data)
	if checkLen > 1024 {
		checkLen = 1024
	}

	nonPrintable := 0
	for i := 0; i < checkLen; i++ {
		b := data[i]
		// Allow common control characters (tab, newline, carriage return)
		if b < 0x20 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintable++
		}
		// Check for null bytes (strong indicator of binary)
		if b == 0x00 {
			return false
		}
	}

	// If more than 10% non-printable, probably not text
	threshold := checkLen / 10
	return nonPrintable < threshold
}

// hasMarkdownIndicators checks for common Markdown patterns.
func hasMarkdownIndicators(data []byte) bool {
	// Sample first 2KB
	checkLen := len(data)
	if checkLen > 2048 {
		checkLen = 2048
	}

	sample := string(data[:checkLen])

	// Check for common Markdown indicators
	mdIndicators := []string{
		"# ",    // Headings
		"## ",   // Headings
		"### ",  // Headings
		"- [ ]", // Task lists
		"- [x]", // Task lists
		"```",   // Code blocks
		"[",     // Links (followed by ](
		"![",    // Images
		"**",    // Bold
		"__",    // Bold
		"> ",    // Blockquotes
		"---",   // Horizontal rules
		"***",   // Horizontal rules
		"| ",    // Tables
		"1. ",   // Ordered lists
		"- ",    // Unordered lists
		"* ",    // Unordered lists
	}

	matchCount := 0
	for _, indicator := range mdIndicators {
		if bytes.Contains([]byte(sample), []byte(indicator)) {
			matchCount++
		}
	}

	// If at least 2 markdown indicators found, likely Markdown
	return matchCount >= 2
}

// ImageInfo contains parsed image header information.
type ImageInfo struct {
	Format     media.MediaFormat
	Width      int
	Height     int
	ColorSpace media.ColorSpace
	BitDepth   int
}

// ParseImageInfo extracts image metadata without fully decoding.
func (f *FormatDetector) ParseImageInfo(data []byte) (*ImageInfo, error) {
	format, err := f.DetectImageFormat(data)
	if err != nil {
		return nil, err
	}

	switch format {
	case media.FormatPNG:
		return f.parsePNGInfo(data)
	case media.FormatJPEG:
		return f.parseJPEGInfo(data)
	case media.FormatGIF:
		return f.parseGIFInfo(data)
	case media.FormatBMP:
		return f.parseBMPInfo(data)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownFormat, format)
	}
}

// parsePNGInfo extracts info from PNG IHDR chunk.
func (f *FormatDetector) parsePNGInfo(data []byte) (*ImageInfo, error) {
	// PNG: 8 byte signature + IHDR chunk
	// IHDR: 4 byte length + 4 byte type + 13 byte data
	if len(data) < 29 {
		return nil, ErrInsufficientData
	}

	// IHDR starts at byte 8
	// Length: bytes 8-11
	// Type: bytes 12-15 ("IHDR")
	// Data: bytes 16-28
	if string(data[12:16]) != "IHDR" {
		return nil, ErrCorruptedHeader
	}

	width := int(binary.BigEndian.Uint32(data[16:20]))
	height := int(binary.BigEndian.Uint32(data[20:24]))
	bitDepth := int(data[24])
	colorType := int(data[25])

	colorSpace := media.ColorSpaceRGB
	switch colorType {
	case 0:
		colorSpace = media.ColorSpaceGray
	case 2:
		colorSpace = media.ColorSpaceRGB
	case 3:
		colorSpace = media.ColorSpaceRGB // Indexed (palette)
	case 4:
		colorSpace = media.ColorSpaceGray // Gray + Alpha
	case 6:
		colorSpace = media.ColorSpaceRGBA
	}

	return &ImageInfo{
		Format:     media.FormatPNG,
		Width:      width,
		Height:     height,
		ColorSpace: colorSpace,
		BitDepth:   bitDepth,
	}, nil
}

// parseJPEGInfo extracts info from JPEG markers.
func (f *FormatDetector) parseJPEGInfo(data []byte) (*ImageInfo, error) {
	if len(data) < 12 {
		return nil, ErrInsufficientData
	}

	// Scan for SOF0 or SOF2 marker (Start of Frame)
	i := 2 // Skip SOI marker
	for i < len(data)-8 {
		if data[i] != 0xFF {
			i++
			continue
		}

		marker := data[i+1]

		// Check for SOF markers (SOF0=0xC0, SOF2=0xC2)
		if marker == 0xC0 || marker == 0xC2 {
			if i+9 >= len(data) {
				return nil, ErrInsufficientData
			}

			precision := int(data[i+4])
			height := int(binary.BigEndian.Uint16(data[i+5 : i+7]))
			width := int(binary.BigEndian.Uint16(data[i+7 : i+9]))
			components := int(data[i+9])

			colorSpace := media.ColorSpaceRGB
			if components == 1 {
				colorSpace = media.ColorSpaceGray
			} else if components == 3 {
				colorSpace = media.ColorSpaceYCbCr
			} else if components == 4 {
				colorSpace = media.ColorSpaceCMYK
			}

			return &ImageInfo{
				Format:     media.FormatJPEG,
				Width:      width,
				Height:     height,
				ColorSpace: colorSpace,
				BitDepth:   precision,
			}, nil
		}

		// Skip to next marker
		if marker == 0xD8 || marker == 0xD9 || marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			i += 2
		} else {
			if i+4 >= len(data) {
				return nil, ErrInsufficientData
			}
			length := int(binary.BigEndian.Uint16(data[i+2 : i+4]))
			i += 2 + length
		}
	}

	return nil, ErrCorruptedHeader
}

// parseGIFInfo extracts info from GIF header.
func (f *FormatDetector) parseGIFInfo(data []byte) (*ImageInfo, error) {
	// GIF: 6 byte signature + 7 byte logical screen descriptor
	if len(data) < 13 {
		return nil, ErrInsufficientData
	}

	width := int(binary.LittleEndian.Uint16(data[6:8]))
	height := int(binary.LittleEndian.Uint16(data[8:10]))
	packed := data[10]

	// Global Color Table Flag and size
	hasGCT := (packed & 0x80) != 0
	bitsPerPixel := 1
	if hasGCT {
		bitsPerPixel = int((packed&0x07)+1) * 3
	}

	return &ImageInfo{
		Format:     media.FormatGIF,
		Width:      width,
		Height:     height,
		ColorSpace: media.ColorSpaceRGB, // GIF uses palette (indexed RGB)
		BitDepth:   bitsPerPixel,
	}, nil
}

// parseBMPInfo extracts info from BMP header.
func (f *FormatDetector) parseBMPInfo(data []byte) (*ImageInfo, error) {
	// BMP: 14 byte file header + DIB header (at least 12 bytes)
	if len(data) < 26 {
		return nil, ErrInsufficientData
	}

	// DIB header size at offset 14
	dibSize := binary.LittleEndian.Uint32(data[14:18])

	var width, height int
	var bitDepth int
	colorSpace := media.ColorSpaceRGB

	if dibSize == 12 {
		// BITMAPCOREHEADER
		width = int(binary.LittleEndian.Uint16(data[18:20]))
		height = int(binary.LittleEndian.Uint16(data[20:22]))
		bitDepth = int(binary.LittleEndian.Uint16(data[24:26]))
	} else {
		// BITMAPINFOHEADER or later
		if len(data) < 38 {
			return nil, ErrInsufficientData
		}
		width = int(int32(binary.LittleEndian.Uint32(data[18:22])))
		h := int32(binary.LittleEndian.Uint32(data[22:26]))
		if h < 0 {
			height = int(-h) // Top-down DIB
		} else {
			height = int(h)
		}
		bitDepth = int(binary.LittleEndian.Uint16(data[28:30]))
	}

	// Determine color space from bit depth
	switch bitDepth {
	case 1, 4, 8:
		colorSpace = media.ColorSpaceRGB // Indexed
	case 24:
		colorSpace = media.ColorSpaceRGB
	case 32:
		colorSpace = media.ColorSpaceRGBA
	}

	return &ImageInfo{
		Format:     media.FormatBMP,
		Width:      width,
		Height:     height,
		ColorSpace: colorSpace,
		BitDepth:   bitDepth,
	}, nil
}

// AudioInfo contains parsed audio header information.
type AudioInfo struct {
	Format     media.MediaFormat
	SampleRate int
	Channels   int
	BitDepth   int
	Duration   float64 // seconds (estimated)
}

// ParseAudioInfo extracts audio metadata without fully decoding.
func (f *FormatDetector) ParseAudioInfo(data []byte) (*AudioInfo, error) {
	format, err := f.DetectAudioFormat(data)
	if err != nil {
		return nil, err
	}

	switch format {
	case media.FormatWAV:
		return f.parseWAVInfo(data)
	case media.FormatFLAC:
		return f.parseFLACInfo(data)
	case media.FormatMP3:
		return f.parseMP3Info(data)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownFormat, format)
	}
}

// parseWAVInfo extracts info from WAV header.
func (f *FormatDetector) parseWAVInfo(data []byte) (*AudioInfo, error) {
	// WAV: RIFF header (12 bytes) + fmt chunk
	if len(data) < 44 {
		return nil, ErrInsufficientData
	}

	// Find fmt chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))

		if chunkID == "fmt " {
			if offset+24 > len(data) {
				return nil, ErrInsufficientData
			}

			// audioFormat := binary.LittleEndian.Uint16(data[offset+8 : offset+10])
			channels := int(binary.LittleEndian.Uint16(data[offset+10 : offset+12]))
			sampleRate := int(binary.LittleEndian.Uint32(data[offset+12 : offset+16]))
			// byteRate := binary.LittleEndian.Uint32(data[offset+16 : offset+20])
			// blockAlign := binary.LittleEndian.Uint16(data[offset+20 : offset+22])
			bitDepth := int(binary.LittleEndian.Uint16(data[offset+22 : offset+24]))

			// Find data chunk for duration
			dataOffset := offset + 8 + chunkSize
			if chunkSize%2 == 1 {
				dataOffset++ // Padding byte
			}

			duration := 0.0
			for dataOffset < len(data)-8 {
				dataChunkID := string(data[dataOffset : dataOffset+4])
				dataChunkSize := int(binary.LittleEndian.Uint32(data[dataOffset+4 : dataOffset+8]))

				if dataChunkID == "data" {
					bytesPerSample := bitDepth / 8
					if bytesPerSample > 0 && sampleRate > 0 && channels > 0 {
						totalSamples := dataChunkSize / (bytesPerSample * channels)
						duration = float64(totalSamples) / float64(sampleRate)
					}
					break
				}
				dataOffset += 8 + dataChunkSize
				if dataChunkSize%2 == 1 {
					dataOffset++
				}
			}

			return &AudioInfo{
				Format:     media.FormatWAV,
				SampleRate: sampleRate,
				Channels:   channels,
				BitDepth:   bitDepth,
				Duration:   duration,
			}, nil
		}

		offset += 8 + chunkSize
		if chunkSize%2 == 1 {
			offset++ // Padding byte
		}
	}

	return nil, ErrCorruptedHeader
}

// parseFLACInfo extracts info from FLAC STREAMINFO block.
func (f *FormatDetector) parseFLACInfo(data []byte) (*AudioInfo, error) {
	// FLAC: 4 byte magic + metadata blocks
	if len(data) < 42 {
		return nil, ErrInsufficientData
	}

	// First metadata block should be STREAMINFO
	blockType := data[4] & 0x7F
	if blockType != 0 { // STREAMINFO type is 0
		return nil, ErrCorruptedHeader
	}

	blockSize := int(data[5])<<16 | int(data[6])<<8 | int(data[7])
	if blockSize < 34 || len(data) < 8+blockSize {
		return nil, ErrInsufficientData
	}

	// Parse STREAMINFO (34 bytes)
	// Bits 0-15: minimum block size
	// Bits 16-31: maximum block size
	// Bits 32-55: minimum frame size
	// Bits 56-79: maximum frame size
	// Bits 80-99: sample rate (20 bits)
	// Bits 100-102: channels - 1 (3 bits)
	// Bits 103-107: bits per sample - 1 (5 bits)
	// Bits 108-143: total samples (36 bits)

	// Sample rate: bytes 18-20 (20 bits)
	sampleRate := int(data[18])<<12 | int(data[19])<<4 | int(data[20]>>4)

	// Channels and bits per sample
	channelsBits := (data[20] & 0x0E) >> 1
	channels := int(channelsBits) + 1

	bpsByte := ((data[20] & 0x01) << 4) | (data[21] >> 4)
	bitDepth := int(bpsByte) + 1

	// Total samples: bytes 21-25 (36 bits)
	totalSamples := int64(data[21]&0x0F)<<32 | int64(data[22])<<24 |
		int64(data[23])<<16 | int64(data[24])<<8 | int64(data[25])

	duration := 0.0
	if sampleRate > 0 {
		duration = float64(totalSamples) / float64(sampleRate)
	}

	return &AudioInfo{
		Format:     media.FormatFLAC,
		SampleRate: sampleRate,
		Channels:   channels,
		BitDepth:   bitDepth,
		Duration:   duration,
	}, nil
}

// parseMP3Info extracts info from MP3 frame header.
func (f *FormatDetector) parseMP3Info(data []byte) (*AudioInfo, error) {
	if len(data) < 10 {
		return nil, ErrInsufficientData
	}

	offset := 0

	// Skip ID3v2 tag if present
	if bytes.HasPrefix(data, id3Magic) {
		if len(data) < 10 {
			return nil, ErrInsufficientData
		}
		// ID3 size is stored in 4 bytes with 7-bit encoding (syncsafe integer)
		tagSize := int(data[6]&0x7F)<<21 | int(data[7]&0x7F)<<14 |
			int(data[8]&0x7F)<<7 | int(data[9]&0x7F)
		offset = 10 + tagSize
	}

	// Find first valid frame header
	for offset < len(data)-4 {
		if data[offset] == 0xFF && (data[offset+1]&0xE0) == 0xE0 {
			// Found frame sync
			version := (data[offset+1] >> 3) & 0x03
			layer := (data[offset+1] >> 1) & 0x03
			sampleRateIdx := (data[offset+2] >> 2) & 0x03
			channelMode := (data[offset+3] >> 6) & 0x03

			// Sample rate table
			sampleRates := [][]int{
				{11025, 12000, 8000, 0},  // MPEG 2.5
				{0, 0, 0, 0},             // Reserved
				{22050, 24000, 16000, 0}, // MPEG 2
				{44100, 48000, 32000, 0}, // MPEG 1
			}

			sampleRate := 0
			if version < 4 && sampleRateIdx < 3 {
				sampleRate = sampleRates[version][sampleRateIdx]
			}

			channels := 2 // stereo by default
			if channelMode == 3 {
				channels = 1 // mono
			}

			// MP3 is always 16-bit output (internally)
			bitDepth := 16

			// Layer III is the most common (layer = 1 means Layer III)
			if layer != 1 {
				// Could be Layer I or II, which is unusual
			}

			return &AudioInfo{
				Format:     media.FormatMP3,
				SampleRate: sampleRate,
				Channels:   channels,
				BitDepth:   bitDepth,
				Duration:   0, // Would need to scan entire file
			}, nil
		}
		offset++
	}

	return nil, ErrCorruptedHeader
}
