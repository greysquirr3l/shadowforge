package media

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatDetector(t *testing.T) {
	detector := NewFormatDetector()
	assert.NotNil(t, detector)
	assert.Equal(t, 12, detector.minImageData)
	assert.Equal(t, 12, detector.minAudioData)
}

func TestFormatDetector_DetectFormat_PNG(t *testing.T) {
	detector := NewFormatDetector()
	pngData := createMinimalPNG(100, 100)

	format, mediaType, err := detector.DetectFormat(pngData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatPNG, format)
	assert.Equal(t, media.MediaTypeImage, mediaType)
}

func TestFormatDetector_DetectFormat_JPEG(t *testing.T) {
	detector := NewFormatDetector()
	jpegData := createMinimalJPEG(320, 240)

	format, mediaType, err := detector.DetectFormat(jpegData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatJPEG, format)
	assert.Equal(t, media.MediaTypeImage, mediaType)
}

func TestFormatDetector_DetectFormat_GIF(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name  string
		magic []byte
		valid bool
	}{
		{"GIF87a", []byte("GIF87a\x00\x00\x00\x00\x00\x00\x00"), true},
		{"GIF89a", []byte("GIF89a\x00\x00\x00\x00\x00\x00\x00"), true},
		{"InvalidGIF", []byte("GIF00a\x00\x00\x00\x00\x00\x00\x00"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, mediaType, err := detector.DetectFormat(tt.magic)
			if tt.valid {
				require.NoError(t, err)
				assert.Equal(t, media.FormatGIF, format)
				assert.Equal(t, media.MediaTypeImage, mediaType)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFormatDetector_DetectFormat_BMP(t *testing.T) {
	detector := NewFormatDetector()
	bmpData := createMinimalBMP(256, 256)

	format, mediaType, err := detector.DetectFormat(bmpData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatBMP, format)
	assert.Equal(t, media.MediaTypeImage, mediaType)
}

func TestFormatDetector_DetectFormat_WAV(t *testing.T) {
	detector := NewFormatDetector()
	wavData := createMinimalWAV(44100, 16, 2, 1000)

	format, mediaType, err := detector.DetectFormat(wavData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatWAV, format)
	assert.Equal(t, media.MediaTypeAudio, mediaType)
}

func TestFormatDetector_DetectFormat_FLAC(t *testing.T) {
	detector := NewFormatDetector()
	flacData := createMinimalFLAC(44100, 16, 2)

	format, mediaType, err := detector.DetectFormat(flacData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatFLAC, format)
	assert.Equal(t, media.MediaTypeAudio, mediaType)
}

func TestFormatDetector_DetectFormat_MP3(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"MP3_FFB", []byte{0xFF, 0xFB, 0x90, 0x00}, true},
		{"MP3_FFA", []byte{0xFF, 0xFA, 0x90, 0x00}, true},
		{"MP3_FFF", []byte{0xFF, 0xF3, 0x90, 0x00}, true},
		{"MP3_ID3", append([]byte("ID3"), make([]byte, 10)...), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, mediaType, err := detector.DetectFormat(tt.data)
			if tt.valid {
				require.NoError(t, err)
				assert.Equal(t, media.FormatMP3, format)
				assert.Equal(t, media.MediaTypeAudio, mediaType)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFormatDetector_DetectFormat_Text(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name     string
		data     []byte
		format   media.MediaFormat
		hasError bool
	}{
		{
			name:   "PlainText",
			data:   []byte("Hello, this is plain text content."),
			format: media.FormatTXT,
		},
		{
			name:   "Markdown",
			data:   []byte("# Heading\n\nSome text with **bold** and `code`.\n\n- Item 1\n- Item 2"),
			format: media.FormatMarkdown,
		},
		{
			name:   "MarkdownWithCodeBlock",
			data:   []byte("# Title\n\n```go\nfunc main() {}\n```\n"),
			format: media.FormatMarkdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, mediaType, err := detector.DetectFormat(tt.data)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.format, format)
				assert.Equal(t, media.MediaTypeText, mediaType)
			}
		})
	}
}

func TestFormatDetector_DetectFormat_Errors(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name string
		data []byte
		err  error
	}{
		{"Empty", []byte{}, ErrInsufficientData},
		{"TooShort", []byte{0x00, 0x01, 0x02}, ErrInsufficientData},
		{"Binary", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}, ErrUnknownFormat},
		{"RandomBinary", []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x00}, ErrUnknownFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := detector.DetectFormat(tt.data)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestFormatDetector_ParseImageInfo_PNG(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name       string
		width      int
		height     int
		colorSpace media.ColorSpace
	}{
		{"Small", 100, 100, media.ColorSpaceRGBA},
		{"Wide", 1920, 1080, media.ColorSpaceRGBA},
		{"Tall", 480, 800, media.ColorSpaceRGBA},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pngData := createMinimalPNG(tt.width, tt.height)
			info, err := detector.ParseImageInfo(pngData)

			require.NoError(t, err)
			assert.Equal(t, media.FormatPNG, info.Format)
			assert.Equal(t, tt.width, info.Width)
			assert.Equal(t, tt.height, info.Height)
		})
	}
}

func TestFormatDetector_ParseImageInfo_JPEG(t *testing.T) {
	detector := NewFormatDetector()
	jpegData := createMinimalJPEG(640, 480)

	info, err := detector.ParseImageInfo(jpegData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatJPEG, info.Format)
	assert.Equal(t, 640, info.Width)
	assert.Equal(t, 480, info.Height)
}

func TestFormatDetector_ParseImageInfo_GIF(t *testing.T) {
	detector := NewFormatDetector()
	gifData := createMinimalGIF(320, 240)

	info, err := detector.ParseImageInfo(gifData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatGIF, info.Format)
	assert.Equal(t, 320, info.Width)
	assert.Equal(t, 240, info.Height)
}

func TestFormatDetector_ParseImageInfo_BMP(t *testing.T) {
	detector := NewFormatDetector()
	bmpData := createMinimalBMP(800, 600)

	info, err := detector.ParseImageInfo(bmpData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatBMP, info.Format)
	assert.Equal(t, 800, info.Width)
	assert.Equal(t, 600, info.Height)
}

func TestFormatDetector_ParseAudioInfo_WAV(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name       string
		sampleRate int
		bitDepth   int
		channels   int
	}{
		{"CD_Quality", 44100, 16, 2},
		{"HighRes", 96000, 24, 2},
		{"Mono", 22050, 8, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wavData := createMinimalWAV(tt.sampleRate, tt.bitDepth, tt.channels, 1000)
			info, err := detector.ParseAudioInfo(wavData)

			require.NoError(t, err)
			assert.Equal(t, media.FormatWAV, info.Format)
			assert.Equal(t, tt.sampleRate, info.SampleRate)
			assert.Equal(t, tt.bitDepth, info.BitDepth)
			assert.Equal(t, tt.channels, info.Channels)
		})
	}
}

func TestFormatDetector_ParseAudioInfo_FLAC(t *testing.T) {
	detector := NewFormatDetector()
	flacData := createMinimalFLAC(48000, 24, 2)

	info, err := detector.ParseAudioInfo(flacData)
	require.NoError(t, err)
	assert.Equal(t, media.FormatFLAC, info.Format)
	assert.Equal(t, 48000, info.SampleRate)
	assert.Equal(t, 24, info.BitDepth)
	assert.Equal(t, 2, info.Channels)
}

func TestFormatDetector_ParseAudioInfo_Errors(t *testing.T) {
	detector := NewFormatDetector()

	tests := []struct {
		name string
		data []byte
		err  error
	}{
		{"TooShort", []byte{0x00, 0x01}, ErrInsufficientData},
		{"NotAudio", createMinimalPNG(10, 10), ErrUnknownFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := detector.ParseAudioInfo(tt.data)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestIsValidUTF8Text(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"EmptyData", []byte{}, false},
		{"ValidASCII", []byte("Hello World"), true},
		{"ValidUTF8", []byte("Hello 世界 🌍"), true},
		{"WithNewlines", []byte("Line 1\nLine 2\r\nLine 3"), true},
		{"WithTabs", []byte("Col1\tCol2\tCol3"), true},
		{"NullByte", []byte("Hello\x00World"), false},
		{"ManyControlChars", bytes.Repeat([]byte{0x01}, 200), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUTF8Text(tt.data)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestHasMarkdownIndicators(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"PlainText", []byte("Just some plain text here."), false},
		{"SingleIndicator", []byte("# Just a heading"), false},
		{"TwoIndicators", []byte("# Heading\n\n**Bold text**"), true},
		{"CodeBlock", []byte("# Title\n```code```"), true},
		{"TaskList", []byte("- [ ] Todo 1\n- [x] Done"), true},
		{"Table", []byte("| Col1 | Col2 |\n- Item"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasMarkdownIndicators(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions to create minimal valid file headers

func createMinimalPNG(width, height int) []byte {
	data := make([]byte, 33)
	copy(data[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	binary.BigEndian.PutUint32(data[8:12], 13)
	copy(data[12:16], []byte("IHDR"))
	binary.BigEndian.PutUint32(data[16:20], uint32(width))
	binary.BigEndian.PutUint32(data[20:24], uint32(height))
	data[24] = 8 // Bit depth
	data[25] = 6 // Color type (RGBA)
	data[26] = 0
	data[27] = 0
	data[28] = 0
	binary.BigEndian.PutUint32(data[29:33], 0)
	return data
}

func createMinimalJPEG(width, height int) []byte {
	// JPEG structure: SOI + APP0 + SOF0
	// SOI: FF D8
	// APP0: FF E0 [length 2 bytes] [data]
	// SOF0: FF C0 [length 2 bytes] [precision 1 byte] [height 2 bytes] [width 2 bytes] [components 1 byte]
	data := make([]byte, 24)

	// SOI marker
	data[0] = 0xFF
	data[1] = 0xD8

	// APP0 marker with minimal length
	data[2] = 0xFF
	data[3] = 0xE0
	data[4] = 0x00
	data[5] = 0x02 // Length = 2 (just the length field itself, minimal)

	// SOF0 marker - Start of Frame baseline DCT
	data[6] = 0xFF
	data[7] = 0xC0
	data[8] = 0x00
	data[9] = 0x0B // Length = 11 bytes (length + precision + height + width + components + component data)
	data[10] = 8   // Precision (8 bits)
	binary.BigEndian.PutUint16(data[11:13], uint16(height))
	binary.BigEndian.PutUint16(data[13:15], uint16(width))
	data[15] = 3 // Number of components (3 = YCbCr)

	return data
}

func createMinimalGIF(width, height int) []byte {
	data := make([]byte, 13)
	copy(data[0:6], []byte("GIF89a"))
	binary.LittleEndian.PutUint16(data[6:8], uint16(width))
	binary.LittleEndian.PutUint16(data[8:10], uint16(height))
	data[10] = 0x87
	data[11] = 0
	data[12] = 0
	return data
}

func createMinimalBMP(width, height int) []byte {
	data := make([]byte, 54)
	data[0] = 'B'
	data[1] = 'M'
	pixelDataSize := width * height * 3
	fileSize := 54 + pixelDataSize
	binary.LittleEndian.PutUint32(data[2:6], uint32(fileSize))
	binary.LittleEndian.PutUint32(data[10:14], 54)
	binary.LittleEndian.PutUint32(data[14:18], 40)
	binary.LittleEndian.PutUint32(data[18:22], uint32(width))
	binary.LittleEndian.PutUint32(data[22:26], uint32(height))
	binary.LittleEndian.PutUint16(data[26:28], 1)
	binary.LittleEndian.PutUint16(data[28:30], 24)
	binary.LittleEndian.PutUint32(data[30:34], 0)
	binary.LittleEndian.PutUint32(data[34:38], uint32(pixelDataSize))
	return data
}

// Note: createMinimalWAV is defined in audio_processor_test.go with signature (sampleRate, bitDepth, channels, numSamples int)

func createMinimalFLAC(sampleRate, bitDepth, channels int) []byte {
	data := make([]byte, 42)
	copy(data[0:4], []byte("fLaC"))
	data[4] = 0x80
	data[5] = 0
	data[6] = 0
	data[7] = 34
	binary.BigEndian.PutUint16(data[8:10], 4096)
	binary.BigEndian.PutUint16(data[10:12], 4096)
	data[18] = byte(sampleRate >> 12)
	data[19] = byte(sampleRate >> 4)
	data[20] = byte((sampleRate&0x0F)<<4) | byte((channels-1)<<1) | byte((bitDepth-1)>>4)
	data[21] = byte((bitDepth - 1) << 4)
	binary.BigEndian.PutUint32(data[22:26], 0)
	return data
}

// Additional coverage tests for low-coverage parsing functions

func TestParseMP3Info(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_MP3_ID3v2",
			data: []byte{
				// ID3v2.3 header (10 bytes)
				0x49, 0x44, 0x33, // ID3
				0x03, 0x00, // Version 2.3
				0x00,                   // Flags
				0x00, 0x00, 0x00, 0x00, // Size: 0 (no tag data beyond header)
				// MP3 frame header at offset 10 (MPEG-1 Layer 3, 128kbps, 44.1kHz)
				// Need 5 bytes because loop checks: offset < len(data)-4, so 10 < 15-4 = true
				0xFF, 0xFB, 0x90, 0x00, 0x00,
			},
			expectError: false,
		},
		{
			name: "Valid_MP3_NoID3",
			data: []byte{
				// MP3 frame header only (MPEG-1 Layer 3, 128kbps, 44.1kHz)
				0xFF, 0xFB, 0x90, 0x00,
				// Need padding to meet minimum length
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0xFF, 0xFB},
			expectError: true,
		},
		{
			name:        "InvalidFrameSync_Error",
			data:        []byte{0x00, 0x00, 0x00, 0x00},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFormatDetector()
			info, err := detector.parseMP3Info(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, media.FormatMP3, info.Format)
			}
		})
	}
}

func TestParsePNGInfo_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		expectError  bool
		expectWidth  int
		expectHeight int
	}{
		{
			name: "Valid_PNG_1x1",
			data: []byte{
				// PNG signature
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
				// IHDR chunk length (13)
				0x00, 0x00, 0x00, 0x0D,
				// IHDR chunk type
				0x49, 0x48, 0x44, 0x52,
				// Width: 1
				0x00, 0x00, 0x00, 0x01,
				// Height: 1
				0x00, 0x00, 0x00, 0x01,
				// Bit depth, color type, compression, filter, interlace
				0x08, 0x02, 0x00, 0x00, 0x00,
				// CRC
				0x00, 0x00, 0x00, 0x00,
			},
			expectError:  false,
			expectWidth:  1,
			expectHeight: 1,
		},
		{
			name: "TooShort_Error",
			data: []byte{
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
				0x00, 0x00, 0x00, 0x0D,
			},
			expectError: true,
		},
		{
			name: "MissingIHDR_Error",
			data: []byte{
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
				0x00, 0x00, 0x00, 0x0D,
				0x00, 0x00, 0x00, 0x00, // Wrong chunk type
				0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x01,
				0x08, 0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFormatDetector()
			info, err := detector.parsePNGInfo(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, tt.expectWidth, info.Width)
				assert.Equal(t, tt.expectHeight, info.Height)
			}
		})
	}
}

func TestParseJPEGInfo_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_JPEG_WithSOF",
			data: []byte{
				// SOI marker
				0xFF, 0xD8,
				// SOF0 marker
				0xFF, 0xC0,
				// Length (17 bytes)
				0x00, 0x11,
				// Precision
				0x08,
				// Height: 100
				0x00, 0x64,
				// Width: 100
				0x00, 0x64,
				// Components
				0x03,
				// Component data
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0xFF, 0xD8, 0xFF},
			expectError: true,
		},
		{
			name: "NoSOF_Error",
			data: []byte{
				0xFF, 0xD8, // SOI
				0xFF, 0xD9, // EOI (no SOF)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFormatDetector()
			info, err := detector.parseJPEGInfo(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
			}
		})
	}
}

func TestParseBMPInfo_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "Valid_BMP_DIBv3",
			data: []byte{
				// BM signature (2 bytes)
				0x42, 0x4D,
				// File size (4 bytes)
				0x36, 0x00, 0x00, 0x00,
				// Reserved (4 bytes)
				0x00, 0x00, 0x00, 0x00,
				// Pixel data offset (4 bytes)
				0x36, 0x00, 0x00, 0x00,
				// DIB header size: 40 bytes (4 bytes)
				0x28, 0x00, 0x00, 0x00,
				// Width: 10 (4 bytes)
				0x0A, 0x00, 0x00, 0x00,
				// Height: 10 (4 bytes)
				0x0A, 0x00, 0x00, 0x00,
				// Color planes: 1 (2 bytes)
				0x01, 0x00,
				// Bits per pixel: 24 (2 bytes)
				0x18, 0x00,
				// Compression: none (4 bytes)
				0x00, 0x00, 0x00, 0x00,
				// Image size (4 bytes)
				0x00, 0x00, 0x00, 0x00,
			},
			expectError: false,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x42, 0x4D, 0x00, 0x00},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFormatDetector()
			info, err := detector.parseBMPInfo(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
			}
		})
	}
}

func TestParseWAVInfo_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name           string
		data           []byte
		expectError    bool
		expectChannels int
	}{
		{
			name: "Valid_WAV_Stereo",
			data: []byte{
				// RIFF header (12 bytes)
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x24, 0x00, 0x00, 0x00, // Chunk size: 36 bytes
				0x57, 0x41, 0x56, 0x45, // "WAVE"
				// fmt subchunk (24 bytes: 8 header + 16 data)
				0x66, 0x6D, 0x74, 0x20, // "fmt "
				0x10, 0x00, 0x00, 0x00, // Subchunk size: 16
				0x01, 0x00, // Audio format: PCM = 1
				0x02, 0x00, // Num channels: 2 (stereo)
				0x44, 0xAC, 0x00, 0x00, // Sample rate: 44100
				0x10, 0xB1, 0x02, 0x00, // Byte rate
				0x04, 0x00, // Block align: 4
				0x10, 0x00, // Bits per sample: 16
				// data subchunk header (8 bytes minimum to reach 44 bytes)
				0x64, 0x61, 0x74, 0x61, // "data"
				0x00, 0x00, 0x00, 0x00, // Data size: 0 (empty)
			},
			expectError:    false,
			expectChannels: 2,
		},
		{
			name:        "TooShort_Error",
			data:        []byte{0x52, 0x49, 0x46, 0x46},
			expectError: true,
		},
		{
			name: "MissingFmt_Error",
			data: []byte{
				0x52, 0x49, 0x46, 0x46,
				0x0C, 0x00, 0x00, 0x00,
				0x57, 0x41, 0x56, 0x45,
				// No fmt chunk
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewFormatDetector()
			info, err := detector.parseWAVInfo(tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				if tt.expectChannels > 0 {
					assert.Equal(t, tt.expectChannels, info.Channels)
				}
			}
		})
	}
}
