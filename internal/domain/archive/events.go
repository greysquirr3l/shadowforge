package archive

import "time"

// ArchiveCreated is emitted when an archive is created.
type ArchiveCreated struct {
ArchiveID        ArchiveID
Format           ArchiveFormat
CompressionLevel CompressionLevel
Timestamp        time.Time
}

// NewArchiveCreated creates a new ArchiveCreated event.
func NewArchiveCreated(archiveID ArchiveID, format ArchiveFormat, compressionLevel CompressionLevel) ArchiveCreated {
return ArchiveCreated{
ArchiveID:        archiveID,
Format:           format,
CompressionLevel: compressionLevel,
Timestamp:        time.Now(),
}
}

// EntryAdded is emitted when an entry is added to an archive.
type EntryAdded struct {
ArchiveID  ArchiveID
EntryID    EntryID
EntryName  string
EntryPath  string
Size       int64
MediaType  MediaType
Timestamp  time.Time
}

// NewEntryAdded creates a new EntryAdded event.
func NewEntryAdded(archiveID ArchiveID, entry *ArchiveEntry) EntryAdded {
return EntryAdded{
ArchiveID:  archiveID,
EntryID:    entry.ID,
EntryName:  entry.Name,
EntryPath:  entry.Path,
Size:       entry.Size,
MediaType:  entry.MediaType,
Timestamp:  time.Now(),
}
}

// ArchiveExtracted is emitted when an archive is extracted.
type ArchiveExtracted struct {
ArchiveID   ArchiveID
EntryCount  int
TotalSize   int64
Timestamp   time.Time
}

// NewArchiveExtracted creates a new ArchiveExtracted event.
func NewArchiveExtracted(archive *Archive) ArchiveExtracted {
return ArchiveExtracted{
ArchiveID:   archive.ID,
EntryCount:  archive.EntryCount,
TotalSize:   archive.TotalSize,
Timestamp:   time.Now(),
}
}

// MediaExtractedFromArchive is emitted when media is extracted from an archive.
type MediaExtractedFromArchive struct {
ArchiveID     ArchiveID
EntryID       EntryID
MediaType     MediaType
Size          int64
OutputPath    string
Timestamp     time.Time
}

// NewMediaExtractedFromArchive creates a new MediaExtractedFromArchive event.
func NewMediaExtractedFromArchive(archiveID ArchiveID, entry *ArchiveEntry, outputPath string) MediaExtractedFromArchive {
return MediaExtractedFromArchive{
ArchiveID:  archiveID,
EntryID:    entry.ID,
MediaType:  entry.MediaType,
Size:       entry.Size,
OutputPath: outputPath,
Timestamp:  time.Now(),
}
}

// StegoMediaPackaged is emitted when stego media is packaged into an archive.
type StegoMediaPackaged struct {
ArchiveID        ArchiveID
StegoMediaCount  int
TotalSize        int64
Timestamp        time.Time
}

// NewStegoMediaPackaged creates a new StegoMediaPackaged event.
func NewStegoMediaPackaged(archiveID ArchiveID, mediaCount int, totalSize int64) StegoMediaPackaged {
return StegoMediaPackaged{
ArchiveID:       archiveID,
StegoMediaCount: mediaCount,
TotalSize:       totalSize,
Timestamp:       time.Now(),
}
}

// ZipBombDetected is emitted when a potential zip bomb is detected.
type ZipBombDetected struct {
ArchiveID        ArchiveID
CompressionRatio float64
TotalSize        int64
CompressedSize   int64
Threshold        float64
Timestamp        time.Time
}

// NewZipBombDetected creates a new ZipBombDetected event.
func NewZipBombDetected(archive *Archive, threshold float64) ZipBombDetected {
return ZipBombDetected{
ArchiveID:        archive.ID,
CompressionRatio: archive.CalculateCompressionRatio(),
TotalSize:        archive.TotalSize,
CompressedSize:   archive.CompressedSize,
Threshold:        threshold,
Timestamp:        time.Now(),
}
}

// ArchiveValidated is emitted when an archive passes validation.
type ArchiveValidated struct {
ArchiveID   ArchiveID
EntryCount  int
TotalSize   int64
Timestamp   time.Time
}

// NewArchiveValidated creates a new ArchiveValidated event.
func NewArchiveValidated(archive *Archive) ArchiveValidated {
return ArchiveValidated{
ArchiveID:  archive.ID,
EntryCount: archive.EntryCount,
TotalSize:  archive.TotalSize,
Timestamp:  time.Now(),
}
}

// ArchiveDeleted is emitted when an archive is deleted.
type ArchiveDeleted struct {
ArchiveID  ArchiveID
Timestamp  time.Time
}

// NewArchiveDeleted creates a new ArchiveDeleted event.
func NewArchiveDeleted(archiveID ArchiveID) ArchiveDeleted {
return ArchiveDeleted{
ArchiveID: archiveID,
Timestamp: time.Now(),
}
}
