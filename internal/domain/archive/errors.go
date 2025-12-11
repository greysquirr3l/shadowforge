package archive

import "errors"

// Archive domain errors - all sentinel errors for archive operations.
var (
	// ID and validation errors
	ErrInvalidArchiveID        = errors.New("invalid archive ID")
	ErrInvalidEntryID          = errors.New("invalid entry ID")
	ErrInvalidFormat           = errors.New("invalid archive format")
	ErrUnsupportedFormat       = errors.New("unsupported archive format")
	ErrInvalidCompressionLevel = errors.New("invalid compression level")

	// Entry errors
	ErrInvalidEntry       = errors.New("invalid archive entry")
	ErrInvalidEntryName   = errors.New("invalid entry name")
	ErrInvalidEntrySize   = errors.New("invalid entry size")
	ErrInvalidPath        = errors.New("invalid path")
	ErrInvalidMediaType   = errors.New("invalid media type")
	ErrEntryNotFound      = errors.New("entry not found")
	ErrEntryAlreadyExists = errors.New("entry already exists in archive")

	// Archive state errors
	ErrArchiveEmpty     = errors.New("archive is empty")
	ErrArchiveCorrupted = errors.New("archive is corrupted")
	ErrArchiveTooLarge  = errors.New("archive exceeds size limit")
	ErrArchiveNotFound  = errors.New("archive not found")

	// Security errors
	ErrPathTraversal    = errors.New("path traversal attack detected")
	ErrZipBombDetected  = errors.New("zip bomb detected")
	ErrExtractionFailed = errors.New("archive extraction failed")
	ErrCreationFailed   = errors.New("archive creation failed")
)
