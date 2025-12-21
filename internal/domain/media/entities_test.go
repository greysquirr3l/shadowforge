package media_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
)

func TestMediaAsset_NewMediaAsset_Success(t *testing.T) {
	// Arrange
	id := media.NewAssetID()
	data := []byte("test media data")

	// Act
	asset, err := media.NewMediaAsset(id, media.MediaTypeImage, media.FormatPNG, data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, id, asset.ID)
	assert.Equal(t, media.MediaTypeImage, asset.Type)
	assert.Equal(t, media.FormatPNG, asset.Format)
	assert.Equal(t, data, asset.Data)
	assert.NotNil(t, asset.Metadata)
	assert.False(t, asset.IsSanitized)
	assert.False(t, asset.FormatValidated)
	assert.WithinDuration(t, time.Now(), asset.LoadedAt, time.Second)
}
func TestMediaAsset_NewMediaAsset_Errors(t *testing.T) {
	tests := []struct {
		name        string
		id          media.AssetID
		mediaType   media.MediaType
		format      media.MediaFormat
		data        []byte
		expectedErr error
	}{
		{
			name:        "zero_id",
			id:          media.AssetID{},
			mediaType:   media.MediaTypeImage,
			format:      media.FormatPNG,
			data:        []byte("data"),
			expectedErr: media.ErrInvalidAssetID,
		},
		{
			name:        "invalid_media_type",
			id:          media.NewAssetID(),
			mediaType:   media.MediaType("invalid"),
			format:      media.FormatPNG,
			data:        []byte("data"),
			expectedErr: media.ErrInvalidMediaType,
		},
		{
			name:        "invalid_format",
			id:          media.NewAssetID(),
			mediaType:   media.MediaTypeImage,
			format:      media.MediaFormat("invalid"),
			data:        []byte("data"),
			expectedErr: media.ErrInvalidFormat,
		},
		{
			name:        "empty_data",
			id:          media.NewAssetID(),
			mediaType:   media.MediaTypeImage,
			format:      media.FormatPNG,
			data:        []byte{},
			expectedErr: media.ErrEmptyMediaData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			asset, err := media.NewMediaAsset(tt.id, tt.mediaType, tt.format, tt.data)

			// Assert
			assert.Nil(t, asset)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestMediaAsset_SetDimensions_Success(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
	dims, _ := media.NewDimensions(1920, 1080)

	// Act
	err := asset.SetDimensions(dims)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, dims, asset.Dimensions)
}

func TestMediaAsset_SetDimensions_Errors(t *testing.T) {
	tests := []struct {
		name        string
		mediaType   media.MediaType
		dimensions  *media.Dimensions
		expectedErr error
	}{
		{
			name:        "not_image_type",
			mediaType:   media.MediaTypeAudio,
			dimensions:  &media.Dimensions{Width: 100, Height: 100},
			expectedErr: media.ErrInvalidMediaType,
		},
		{
			name:        "nil_dimensions",
			mediaType:   media.MediaTypeImage,
			dimensions:  nil,
			expectedErr: media.ErrInvalidDimensions,
		},
		{
			name:        "invalid_dimensions",
			mediaType:   media.MediaTypeImage,
			dimensions:  &media.Dimensions{Width: -1, Height: 100},
			expectedErr: media.ErrInvalidDimensions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset, _ := media.NewMediaAsset(media.NewAssetID(), tt.mediaType, media.FormatPNG, []byte("data"))

			// Act
			err := asset.SetDimensions(tt.dimensions)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestMediaAsset_SetResolution_Success(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
	res, _ := media.NewResolution(300, 300)

	// Act
	err := asset.SetResolution(res)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, res, asset.Resolution)
}

func TestMediaAsset_SetResolution_Errors(t *testing.T) {
	tests := []struct {
		name        string
		mediaType   media.MediaType
		resolution  *media.Resolution
		expectedErr error
	}{
		{
			name:        "not_image_type",
			mediaType:   media.MediaTypeAudio,
			resolution:  &media.Resolution{Horizontal: 300, Vertical: 300},
			expectedErr: media.ErrInvalidMediaType,
		},
		{
			name:        "nil_resolution",
			mediaType:   media.MediaTypeImage,
			resolution:  nil,
			expectedErr: media.ErrInvalidResolution,
		},
		{
			name:        "invalid_resolution",
			mediaType:   media.MediaTypeImage,
			resolution:  &media.Resolution{Horizontal: -1, Vertical: 300},
			expectedErr: media.ErrInvalidResolution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset, _ := media.NewMediaAsset(media.NewAssetID(), tt.mediaType, media.FormatPNG, []byte("data"))

			// Act
			err := asset.SetResolution(tt.resolution)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestMediaAsset_SetSampleRate_Success(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeAudio, media.FormatWAV, []byte("data"))
	sr, _ := media.NewSampleRate(44100, 16, 2)

	// Act
	err := asset.SetSampleRate(sr)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, sr, asset.SampleRate)
}

func TestMediaAsset_SetSampleRate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		mediaType   media.MediaType
		sampleRate  *media.SampleRate
		expectedErr error
	}{
		{
			name:        "not_audio_type",
			mediaType:   media.MediaTypeImage,
			sampleRate:  &media.SampleRate{Hz: 44100, BitDepth: 16, Channels: 2},
			expectedErr: media.ErrInvalidMediaType,
		},
		{
			name:        "nil_sample_rate",
			mediaType:   media.MediaTypeAudio,
			sampleRate:  nil,
			expectedErr: media.ErrInvalidSampleRate,
		},
		{
			name:        "invalid_sample_rate",
			mediaType:   media.MediaTypeAudio,
			sampleRate:  &media.SampleRate{Hz: -1, BitDepth: 16, Channels: 2},
			expectedErr: media.ErrInvalidSampleRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset, _ := media.NewMediaAsset(media.NewAssetID(), tt.mediaType, media.FormatWAV, []byte("data"))

			// Act
			err := asset.SetSampleRate(tt.sampleRate)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestMediaAsset_SetColorSpace_Success(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
	cs := media.ColorSpaceRGB

	// Act
	err := asset.SetColorSpace(&cs)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &cs, asset.ColorSpace)
}

func TestMediaAsset_SetColorSpace_Errors(t *testing.T) {
	tests := []struct {
		name        string
		mediaType   media.MediaType
		colorSpace  *media.ColorSpace
		expectedErr error
	}{
		{
			name:        "not_image_type",
			mediaType:   media.MediaTypeAudio,
			colorSpace:  func() *media.ColorSpace { cs := media.ColorSpaceRGB; return &cs }(),
			expectedErr: media.ErrInvalidMediaType,
		},
		{
			name:        "nil_color_space",
			mediaType:   media.MediaTypeImage,
			colorSpace:  nil,
			expectedErr: media.ErrInvalidColorSpace,
		},
		{
			name:        "invalid_color_space",
			mediaType:   media.MediaTypeImage,
			colorSpace:  func() *media.ColorSpace { cs := media.ColorSpace("invalid"); return &cs }(),
			expectedErr: media.ErrInvalidColorSpace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset, _ := media.NewMediaAsset(media.NewAssetID(), tt.mediaType, media.FormatPNG, []byte("data"))

			// Act
			err := asset.SetColorSpace(tt.colorSpace)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestMediaAsset_SetCapacity(t *testing.T) {
	tests := []struct {
		name        string
		capacity    int64
		expectedErr error
	}{
		{"valid_zero", 0, nil},
		{"valid_positive", 1024, nil},
		{"invalid_negative", -1, media.ErrInvalidCapacity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))

			// Act
			err := asset.SetCapacity(tt.capacity)

			// Assert
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.capacity, asset.Capacity)
			}
		})
	}
}

func TestMediaAsset_Sanitize(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
	asset.Metadata["key"] = "value"

	// Act
	asset.Sanitize()

	// Assert
	assert.True(t, asset.IsSanitized)
	assert.Empty(t, asset.Metadata)
}

func TestMediaAsset_ValidateFormat(t *testing.T) {
	// Arrange
	asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))

	// Act
	asset.ValidateFormat()

	// Assert
	assert.True(t, asset.FormatValidated)
}

func TestMediaAsset_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupAsset  func() *media.MediaAsset
		expectedErr error
	}{
		{
			name: "valid_image_with_dimensions",
			setupAsset: func() *media.MediaAsset {
				asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
				dims, _ := media.NewDimensions(100, 100)
				if err := asset.SetDimensions(dims); err != nil {
					panic(err)
				}
				return asset
			},
			expectedErr: nil,
		},
		{
			name: "valid_audio_with_sample_rate",
			setupAsset: func() *media.MediaAsset {
				asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeAudio, media.FormatWAV, []byte("data"))
				sr, _ := media.NewSampleRate(44100, 16, 2)
				if err := asset.SetSampleRate(sr); err != nil {
					panic(err)
				}
				return asset
			},
			expectedErr: nil,
		},
		{
			name: "valid_text",
			setupAsset: func() *media.MediaAsset {
				asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeText, media.FormatTXT, []byte("data"))
				return asset
			},
			expectedErr: nil,
		},
		{
			name: "image_missing_dimensions",
			setupAsset: func() *media.MediaAsset {
				asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeImage, media.FormatPNG, []byte("data"))
				return asset
			},
			expectedErr: media.ErrInvalidDimensions,
		},
		{
			name: "audio_missing_sample_rate",
			setupAsset: func() *media.MediaAsset {
				asset, _ := media.NewMediaAsset(media.NewAssetID(), media.MediaTypeAudio, media.FormatWAV, []byte("data"))
				return asset
			},
			expectedErr: media.ErrInvalidSampleRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			asset := tt.setupAsset()

			// Act
			err := asset.Validate()

			// Assert
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCapacityInfo_NewCapacityInfo_Success(t *testing.T) {
	// Arrange
	id := media.NewAssetID()

	// Act
	info, err := media.NewCapacityInfo(id, media.MediaTypeImage, 10000, 8000, 6000)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, id, info.AssetID)
	assert.Equal(t, media.MediaTypeImage, info.MediaType)
	assert.Equal(t, int64(10000), info.TotalCapacity)
	assert.Equal(t, int64(8000), info.UsableCapacity)
	assert.Equal(t, int64(6000), info.RecommendedMax)
}

func TestCapacityInfo_NewCapacityInfo_Errors(t *testing.T) {
	tests := []struct {
		name           string
		id             media.AssetID
		mediaType      media.MediaType
		totalCapacity  int64
		usableCapacity int64
		recommendedMax int64
		expectedErr    error
	}{
		{
			name:           "zero_id",
			id:             media.AssetID{},
			mediaType:      media.MediaTypeImage,
			totalCapacity:  100,
			usableCapacity: 80,
			recommendedMax: 60,
			expectedErr:    media.ErrInvalidAssetID,
		},
		{
			name:           "invalid_media_type",
			id:             media.NewAssetID(),
			mediaType:      media.MediaType("invalid"),
			totalCapacity:  100,
			usableCapacity: 80,
			recommendedMax: 60,
			expectedErr:    media.ErrInvalidMediaType,
		},
		{
			name:           "negative_total_capacity",
			id:             media.NewAssetID(),
			mediaType:      media.MediaTypeImage,
			totalCapacity:  -1,
			usableCapacity: 80,
			recommendedMax: 60,
			expectedErr:    media.ErrInvalidCapacity,
		},
		{
			name:           "usable_exceeds_total",
			id:             media.NewAssetID(),
			mediaType:      media.MediaTypeImage,
			totalCapacity:  100,
			usableCapacity: 200,
			recommendedMax: 60,
			expectedErr:    media.ErrInvalidCapacity,
		},
		{
			name:           "recommended_exceeds_usable",
			id:             media.NewAssetID(),
			mediaType:      media.MediaTypeImage,
			totalCapacity:  100,
			usableCapacity: 80,
			recommendedMax: 90,
			expectedErr:    media.ErrInvalidCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			info, err := media.NewCapacityInfo(tt.id, tt.mediaType, tt.totalCapacity, tt.usableCapacity, tt.recommendedMax)

			// Assert
			assert.Nil(t, info)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCapacityInfo_SetQualityImpact(t *testing.T) {
	tests := []struct {
		name        string
		impact      float64
		expectedErr error
	}{
		{"valid_zero", 0.0, nil},
		{"valid_half", 0.5, nil},
		{"valid_one", 1.0, nil},
		{"invalid_negative", -0.1, media.ErrInvalidQualityScore},
		{"invalid_greater_than_one", 1.1, media.ErrInvalidQualityScore},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			info, _ := media.NewCapacityInfo(media.NewAssetID(), media.MediaTypeImage, 100, 80, 60)

			// Act
			err := info.SetQualityImpact(tt.impact)

			// Assert
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.impact, info.QualityImpact)
			}
		})
	}
}
