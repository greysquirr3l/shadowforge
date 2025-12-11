package media

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// AssetID Tests
// ============================================================================

func TestAssetID_NewAssetID_Success(t *testing.T) {
	id := NewAssetID()

	assert.NotEmpty(t, id.String())
	// Verify it's a valid UUID
	_, err := uuid.Parse(id.String())
	assert.NoError(t, err)
	assert.False(t, id.IsZero())
}

func TestAssetID_NewAssetIDFromString_Success(t *testing.T) {
	validUUID := uuid.New().String()

	id, err := NewAssetIDFromString(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
	assert.False(t, id.IsZero())
}

func TestAssetID_NewAssetIDFromString_EmptyString(t *testing.T) {
	id, err := NewAssetIDFromString("")

	assert.ErrorIs(t, err, ErrInvalidAssetID)
	assert.True(t, id.IsZero())
}

func TestAssetID_NewAssetIDFromString_InvalidUUID(t *testing.T) {
	id, err := NewAssetIDFromString("not-a-uuid")

	assert.ErrorIs(t, err, ErrInvalidAssetID)
	assert.True(t, id.IsZero())
}

func TestAssetID_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		id       AssetID
		expected bool
	}{
		{
			name:     "zero_value",
			id:       AssetID{},
			expected: true,
		},
		{
			name:     "with_value",
			id:       NewAssetID(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsZero())
		})
	}
}

func TestAssetID_Equals(t *testing.T) {
	id1 := NewAssetID()
	id2 := NewAssetID()
	id3, _ := NewAssetIDFromString(id1.String())

	assert.False(t, id1.Equals(id2), "Different IDs should not be equal")
	assert.True(t, id1.Equals(id3), "Same UUID string should be equal")
}

// ============================================================================
// MediaType Tests
// ============================================================================

func TestMediaType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		mediaType MediaType
		expected  bool
	}{
		{"valid_image", MediaTypeImage, true},
		{"valid_audio", MediaTypeAudio, true},
		{"valid_text", MediaTypeText, true},
		{"invalid_type", MediaType("video"), false},
		{"empty_type", MediaType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mediaType.IsValid())
		})
	}
}

func TestMediaType_String(t *testing.T) {
	assert.Equal(t, "image", MediaTypeImage.String())
	assert.Equal(t, "audio", MediaTypeAudio.String())
	assert.Equal(t, "text", MediaTypeText.String())
}

// ============================================================================
// MediaFormat Tests
// ============================================================================

func TestMediaFormat_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		format   MediaFormat
		expected bool
	}{
		// Image formats
		{"valid_png", FormatPNG, true},
		{"valid_jpeg", FormatJPEG, true},
		{"valid_bmp", FormatBMP, true},
		{"valid_gif", FormatGIF, true},
		// Audio formats
		{"valid_wav", FormatWAV, true},
		{"valid_flac", FormatFLAC, true},
		{"valid_mp3", FormatMP3, true},
		// Text formats
		{"valid_txt", FormatTXT, true},
		{"valid_markdown", FormatMarkdown, true},
		// Invalid
		{"invalid_format", MediaFormat("docx"), false},
		{"empty_format", MediaFormat(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.IsValid())
		})
	}
}

func TestMediaFormat_String(t *testing.T) {
	tests := []struct {
		format   MediaFormat
		expected string
	}{
		{FormatPNG, "png"},
		{FormatJPEG, "jpeg"},
		{FormatBMP, "bmp"},
		{FormatGIF, "gif"},
		{FormatWAV, "wav"},
		{FormatFLAC, "flac"},
		{FormatMP3, "mp3"},
		{FormatTXT, "txt"},
		{FormatMarkdown, "md"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.String())
		})
	}
}

func TestMediaFormat_IsImage(t *testing.T) {
	tests := []struct {
		name     string
		format   MediaFormat
		expected bool
	}{
		{"png_is_image", FormatPNG, true},
		{"jpeg_is_image", FormatJPEG, true},
		{"bmp_is_image", FormatBMP, true},
		{"gif_is_image", FormatGIF, true},
		{"wav_not_image", FormatWAV, false},
		{"txt_not_image", FormatTXT, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.IsImage())
		})
	}
}

func TestMediaFormat_IsAudio(t *testing.T) {
	tests := []struct {
		name     string
		format   MediaFormat
		expected bool
	}{
		{"wav_is_audio", FormatWAV, true},
		{"flac_is_audio", FormatFLAC, true},
		{"mp3_is_audio", FormatMP3, true},
		{"png_not_audio", FormatPNG, false},
		{"txt_not_audio", FormatTXT, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.IsAudio())
		})
	}
}

func TestMediaFormat_IsText(t *testing.T) {
	tests := []struct {
		name     string
		format   MediaFormat
		expected bool
	}{
		{"txt_is_text", FormatTXT, true},
		{"markdown_is_text", FormatMarkdown, true},
		{"png_not_text", FormatPNG, false},
		{"wav_not_text", FormatWAV, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.IsText())
		})
	}
}

// ============================================================================
// Dimensions Tests
// ============================================================================

func TestDimensions_NewDimensions_Success(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small_100x100", 100, 100},
		{"hd_1920x1080", 1920, 1080},
		{"4k_3840x2160", 3840, 2160},
		{"wide_2560x1080", 2560, 1080},
		{"tall_1080x1920", 1080, 1920},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := NewDimensions(tt.width, tt.height)

			require.NoError(t, err)
			assert.Equal(t, tt.width, dims.Width)
			assert.Equal(t, tt.height, dims.Height)
			assert.True(t, dims.IsValid())
		})
	}
}

func TestDimensions_NewDimensions_InvalidWidth(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"zero_width", 0, 100},
		{"negative_width", -10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := NewDimensions(tt.width, tt.height)

			assert.ErrorIs(t, err, ErrInvalidDimensions)
			assert.Nil(t, dims)
		})
	}
}

func TestDimensions_NewDimensions_InvalidHeight(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"zero_height", 100, 0},
		{"negative_height", 100, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := NewDimensions(tt.width, tt.height)

			assert.ErrorIs(t, err, ErrInvalidDimensions)
			assert.Nil(t, dims)
		})
	}
}

func TestDimensions_AspectRatio(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		ratio  float64
	}{
		{"square_1_1", 1000, 1000, 1.0},
		{"16_9", 1920, 1080, 1.778},
		{"4_3", 1024, 768, 1.333},
		{"21_9", 2560, 1080, 2.370},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, _ := NewDimensions(tt.width, tt.height)

			assert.InDelta(t, tt.ratio, dims.AspectRatio(), 0.001)
		})
	}
}

func TestDimensions_Area(t *testing.T) {
	dims, _ := NewDimensions(1920, 1080)

	assert.Equal(t, 2073600, dims.Area())
}

// ============================================================================
// Resolution Tests
// ============================================================================

func TestResolution_NewResolution_Success(t *testing.T) {
	tests := []struct {
		name       string
		horizontal int
		vertical   int
	}{
		{"low_72dpi", 72, 72},
		{"standard_96dpi", 96, 96},
		{"print_300dpi", 300, 300},
		{"high_600dpi", 600, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewResolution(tt.horizontal, tt.vertical)

			require.NoError(t, err)
			assert.Equal(t, tt.horizontal, res.Horizontal)
			assert.Equal(t, tt.vertical, res.Vertical)
			assert.True(t, res.IsValid())
		})
	}
}

func TestResolution_NewResolution_InvalidHorizontal(t *testing.T) {
	res, err := NewResolution(0, 96)

	assert.ErrorIs(t, err, ErrInvalidResolution)
	assert.Nil(t, res)
}

func TestResolution_NewResolution_InvalidVertical(t *testing.T) {
	res, err := NewResolution(96, -10)

	assert.ErrorIs(t, err, ErrInvalidResolution)
	assert.Nil(t, res)
}

// ============================================================================
// SampleRate Tests
// ============================================================================

func TestSampleRate_NewSampleRate_Success(t *testing.T) {
	tests := []struct {
		name     string
		hz       int
		bitDepth int
		channels int
	}{
		{"cd_quality", 44100, 16, 2},
		{"high_quality", 48000, 24, 2},
		{"studio_quality", 96000, 24, 2},
		{"archive_quality", 192000, 32, 2},
		{"mono_8bit", 8000, 8, 1},
		{"surround_5_1", 48000, 24, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr, err := NewSampleRate(tt.hz, tt.bitDepth, tt.channels)

			require.NoError(t, err)
			assert.Equal(t, tt.hz, sr.Hz)
			assert.Equal(t, tt.bitDepth, sr.BitDepth)
			assert.Equal(t, tt.channels, sr.Channels)
			assert.True(t, sr.IsValid())
		})
	}
}

func TestSampleRate_NewSampleRate_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		hz       int
		bitDepth int
		channels int
	}{
		{"zero_hz", 0, 16, 2},
		{"negative_hz", -1, 16, 2},
		{"invalid_bitdepth", 44100, 12, 2},
		{"zero_channels", 44100, 16, 0},
		{"too_many_channels", 44100, 16, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr, err := NewSampleRate(tt.hz, tt.bitDepth, tt.channels)

			assert.ErrorIs(t, err, ErrInvalidSampleRate)
			assert.Nil(t, sr)
		})
	}
}

// ============================================================================
// ColorSpace Tests
// ============================================================================

func TestColorSpace_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		colorSpace ColorSpace
		expected   bool
	}{
		{"valid_rgb", ColorSpaceRGB, true},
		{"valid_rgba", ColorSpaceRGBA, true},
		{"valid_gray", ColorSpaceGray, true},
		{"valid_cmyk", ColorSpaceCMYK, true},
		{"valid_ycbcr", ColorSpaceYCbCr, true},
		{"invalid_space", ColorSpace("hsv"), false},
		{"empty_space", ColorSpace(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colorSpace.IsValid())
		})
	}
}

func TestColorSpace_String(t *testing.T) {
	tests := []struct {
		colorSpace ColorSpace
		expected   string
	}{
		{ColorSpaceRGB, "rgb"},
		{ColorSpaceRGBA, "rgba"},
		{ColorSpaceGray, "gray"},
		{ColorSpaceCMYK, "cmyk"},
		{ColorSpaceYCbCr, "ycbcr"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colorSpace.String())
		})
	}
}

func TestColorSpace_HasAlpha(t *testing.T) {
	tests := []struct {
		name       string
		colorSpace ColorSpace
		expected   bool
	}{
		{"rgb_no_alpha", ColorSpaceRGB, false},
		{"rgba_has_alpha", ColorSpaceRGBA, true},
		{"gray_no_alpha", ColorSpaceGray, false},
		{"cmyk_no_alpha", ColorSpaceCMYK, false},
		{"ycbcr_no_alpha", ColorSpaceYCbCr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.colorSpace.HasAlpha())
		})
	}
}
