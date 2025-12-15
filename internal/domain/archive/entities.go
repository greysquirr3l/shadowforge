// Package archive provides domain entities and logic for managing archive files
// (ZIP, TAR, TAR.GZ, TAR.BZ2) with security validation and media extraction.
package archive

import (
	"time"

	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Archive is the aggregate root for archive file management.
// It coordinates entries, tracks metadata, and enforces security policies.
type Archive struct {
	ID               ArchiveID
	Format           ArchiveFormat
	CompressionLevel CompressionLevel
	Entries          []*ArchiveEntry
	TotalSize        int64
	CompressedSize   int64
	EntryCount       int
	CreatedAt        time.Time
	ExtractedAt      *time.Time
}

// NewArchive creates a new Archive aggregate with validation.
func NewArchive(format ArchiveFormat, compressionLevel CompressionLevel) (*Archive, error) {
	if !format.IsValid() {
		return nil, ErrInvalidFormat
	}
	if !compressionLevel.IsValid() {
		return nil, ErrInvalidCompressionLevel
	}

	archiveID := NewArchiveID()
	archive := &Archive{
		ID:               archiveID,
		Format:           format,
		CompressionLevel: compressionLevel,
		Entries:          make([]*ArchiveEntry, 0),
		TotalSize:        0,
		CompressedSize:   0,
		EntryCount:       0,
		CreatedAt:        time.Now(),
	}

	logger.WithFields(logrus.Fields{
		"archive_id":        archiveID.String(),
		"format":            format.String(),
		"compression_level": compressionLevel.String(),
	}).Info("Archive created")

	return archive, nil
}

// AddEntry adds an ArchiveEntry to the archive with validation.
func (a *Archive) AddEntry(entry *ArchiveEntry) error {
	if entry == nil {
		return ErrInvalidEntry
	}
	if err := entry.Validate(); err != nil {
		return err
	}

	// Check for duplicate paths
	for _, existing := range a.Entries {
		if existing.Path == entry.Path {
			return ErrEntryAlreadyExists
		}
	}

	a.Entries = append(a.Entries, entry)
	a.TotalSize += entry.Size
	a.CompressedSize += entry.CompressedSize
	a.EntryCount++

	logger.WithFields(logrus.Fields{
		"archive_id":  a.ID.String(),
		"entry_id":    entry.ID.String(),
		"entry_path":  entry.Path,
		"entry_size":  entry.Size,
		"entry_count": a.EntryCount,
	}).Debug("Entry added to archive")

	return nil
}

// Extract marks the archive as extracted and records the timestamp.
func (a *Archive) Extract() error {
	now := time.Now()
	a.ExtractedAt = &now

	logger.WithFields(logrus.Fields{
		"archive_id":  a.ID.String(),
		"entry_count": a.EntryCount,
		"total_size":  a.TotalSize,
	}).Info("Archive extracted")

	return nil
}

// Validate performs comprehensive validation on the archive.
func (a *Archive) Validate() error {
	if !a.ID.IsValid() {
		return ErrInvalidArchiveID
	}
	if !a.Format.IsValid() {
		return ErrInvalidFormat
	}
	if a.EntryCount == 0 {
		return ErrArchiveEmpty
	}
	if a.TotalSize == 0 {
		return ErrArchiveCorrupted
	}

	// Validate all entries
	for _, entry := range a.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GetEntry retrieves an entry by ID.
func (a *Archive) GetEntry(entryID EntryID) (*ArchiveEntry, error) {
	for _, entry := range a.Entries {
		if entry.ID == entryID {
			return entry, nil
		}
	}
	return nil, ErrEntryNotFound
}

// GetEntries returns all entries in the archive.
func (a *Archive) GetEntries() []*ArchiveEntry {
	return a.Entries
}

// CalculateCompressionRatio returns the compression ratio (0.0 to 1.0).
func (a *Archive) CalculateCompressionRatio() float64 {
	if a.TotalSize == 0 {
		return 0.0
	}
	return float64(a.CompressedSize) / float64(a.TotalSize)
}

// DetectZipBomb checks if the archive has suspicious compression ratio.
func (a *Archive) DetectZipBomb(threshold float64) bool {
	ratio := a.CalculateCompressionRatio()
	isZipBomb := ratio < threshold

	if isZipBomb {
		logger.WithFields(logrus.Fields{
			"archive_id":        a.ID.String(),
			"compression_ratio": ratio,
			"threshold":         threshold,
			"total_size":        a.TotalSize,
			"compressed_size":   a.CompressedSize,
		}).Warn("Potential zip bomb detected")
	}

	return isZipBomb
}

// ArchiveEntry represents a single file or directory within an archive.
type ArchiveEntry struct {
	ID             EntryID
	ArchiveID      ArchiveID
	Name           string
	Path           string
	MediaType      MediaType
	Size           int64
	CompressedSize int64
	Checksum       string
	Permissions    uint32
	ModifiedAt     time.Time
	ModTime        time.Time // Alias for ModifiedAt (compatibility)
	ExtractedAt    *time.Time
	Content        []byte // File content (for in-memory operations)
}

// NewArchiveEntry creates a new ArchiveEntry with validation.
func NewArchiveEntry(archiveID ArchiveID, name, path string, mediaType MediaType, size int64) (*ArchiveEntry, error) {
	if !archiveID.IsValid() {
		return nil, ErrInvalidArchiveID
	}
	if name == "" {
		return nil, ErrInvalidEntryName
	}
	if path == "" {
		return nil, ErrInvalidPath
	}
	if !mediaType.IsValid() {
		return nil, ErrInvalidMediaType
	}
	if size < 0 {
		return nil, ErrInvalidEntrySize
	}

	entry := &ArchiveEntry{
		ID:          NewEntryID(),
		ArchiveID:   archiveID,
		Name:        name,
		Path:        path,
		MediaType:   mediaType,
		Size:        size,
		Permissions: 0644,
		ModifiedAt:  time.Now(),
	}

	return entry, nil
}

// IsDirectory checks if the entry represents a directory.
func (e *ArchiveEntry) IsDirectory() bool {
	return e.MediaType == TypeDirectory
}

// IsFile checks if the entry represents a file.
func (e *ArchiveEntry) IsFile() bool {
	return !e.IsDirectory()
}

// Validate performs validation on the entry.
func (e *ArchiveEntry) Validate() error {
	if !e.ID.IsValid() {
		return ErrInvalidEntryID
	}
	if !e.ArchiveID.IsValid() {
		return ErrInvalidArchiveID
	}
	if e.Name == "" {
		return ErrInvalidEntryName
	}
	if e.Path == "" {
		return ErrInvalidPath
	}
	if !e.MediaType.IsValid() {
		return ErrInvalidMediaType
	}
	if e.Size < 0 {
		return ErrInvalidEntrySize
	}

	return nil
}

// MarkExtracted records the extraction timestamp.
func (e *ArchiveEntry) MarkExtracted() {
	now := time.Now()
	e.ExtractedAt = &now
}
