// Package media domain errors.
package media

import "errors"

var (
	// ErrInvalidAssetID indicates the asset ID is invalid.
	ErrInvalidAssetID = errors.New("invalid asset ID")

	// ErrInvalidMediaType indicates the media type is invalid or unsupported.
	ErrInvalidMediaType = errors.New("invalid media type")

	// ErrInvalidFormat indicates the media format is invalid or unsupported.
	ErrInvalidFormat = errors.New("invalid media format")

	// ErrEmptyMediaData indicates the media data is empty.
	ErrEmptyMediaData = errors.New("empty media data")

	// ErrInvalidDimensions indicates image dimensions are invalid.
	ErrInvalidDimensions = errors.New("invalid dimensions")

	// ErrInvalidResolution indicates image resolution is invalid.
	ErrInvalidResolution = errors.New("invalid resolution")

	// ErrInvalidSampleRate indicates audio sample rate is invalid.
	ErrInvalidSampleRate = errors.New("invalid sample rate")

	// ErrInvalidColorSpace indicates image color space is invalid.
	ErrInvalidColorSpace = errors.New("invalid color space")

	// ErrInvalidCapacity indicates embedding capacity is invalid.
	ErrInvalidCapacity = errors.New("invalid capacity")

	// ErrUnsupportedMediaType indicates the media type is not supported.
	ErrUnsupportedMediaType = errors.New("unsupported media type")

	// ErrInvalidQualityScore indicates quality score is out of range.
	ErrInvalidQualityScore = errors.New("invalid quality score")

	// ErrMediaNotLoaded indicates media asset has not been loaded.
	ErrMediaNotLoaded = errors.New("media not loaded")

	// ErrMediaAlreadyLoaded indicates media asset is already loaded.
	ErrMediaAlreadyLoaded = errors.New("media already loaded")

	// ErrAssetNotFound indicates the media asset was not found.
	ErrAssetNotFound = errors.New("asset not found")
)
