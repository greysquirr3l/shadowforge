// Package stego implements the Steganography bounded context.
package stego

import "errors"

var (
	// ErrInvalidContainerID indicates an invalid or empty container ID.
	ErrInvalidContainerID = errors.New("invalid container ID")

	// ErrEmptyCoverMedia indicates the cover media is empty.
	ErrEmptyCoverMedia = errors.New("empty cover media")

	// ErrInvalidTechnique indicates an unsupported or invalid steganography technique.
	ErrInvalidTechnique = errors.New("invalid steganography technique")

	// ErrEmptyPayload indicates an attempt to embed empty data.
	ErrEmptyPayload = errors.New("empty payload")

	// ErrAlreadyEmbedded indicates the container already has embedded data.
	ErrAlreadyEmbedded = errors.New("container already has embedded data")

	// ErrCapacityExceeded indicates the payload exceeds available capacity.
	ErrCapacityExceeded = errors.New("payload exceeds available capacity")

	// ErrNoEmbeddedData indicates no data has been embedded in the container.
	ErrNoEmbeddedData = errors.New("no embedded data found")

	// ErrInvalidCapacity indicates an invalid capacity value.
	ErrInvalidCapacity = errors.New("invalid capacity")

	// ErrInvalidQuality indicates an invalid quality score.
	ErrInvalidQuality = errors.New("invalid quality score")

	// ErrInvalidPayloadSize indicates an invalid payload size.
	ErrInvalidPayloadSize = errors.New("invalid payload size")

	// ErrInvalidScore indicates an invalid score value (must be 0.0-1.0).
	ErrInvalidScore = errors.New("invalid score value")

	// ErrEmbeddingFailed indicates the embedding operation failed.
	ErrEmbeddingFailed = errors.New("embedding operation failed")

	// ErrExtractionFailed indicates the extraction operation failed.
	ErrExtractionFailed = errors.New("extraction operation failed")

	// ErrInvalidEmbeddingMap indicates an invalid or corrupted embedding map.
	ErrInvalidEmbeddingMap = errors.New("invalid embedding map")

	// ErrCorruptedContainer indicates the container data is corrupted.
	ErrCorruptedContainer = errors.New("corrupted container")

	// ErrIncompatibleMedia indicates the media type is incompatible with the technique.
	ErrIncompatibleMedia = errors.New("incompatible media type")
)
