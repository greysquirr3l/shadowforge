package archive

import (
"fmt"
"os"

"github.com/google/uuid"
)

// ArchiveID is a unique identifier for archives.
type ArchiveID struct {
value string
}

// NewArchiveID creates a new archive ID.
func NewArchiveID() ArchiveID {
return ArchiveID{value: uuid.New().String()}
}

// NewArchiveIDFromString creates an archive ID from a string.
func NewArchiveIDFromString(id string) (ArchiveID, error) {
if id == "" {
return ArchiveID{}, ErrInvalidArchiveID
}
return ArchiveID{value: id}, nil
}

// String returns the string representation.
func (a ArchiveID) String() string {
return a.value
}

// IsValid checks if the ID is valid.
func (a ArchiveID) IsValid() bool {
return a.value != ""
}

// Equals checks equality with another ArchiveID.
func (a ArchiveID) Equals(other ArchiveID) bool {
return a.value == other.value
}

// EntryID is a unique identifier for archive entries.
type EntryID struct {
value string
}

// NewEntryID creates a new entry ID.
func NewEntryID() EntryID {
return EntryID{value: uuid.New().String()}
}

// NewEntryIDFromString creates an entry ID from a string.
func NewEntryIDFromString(id string) (EntryID, error) {
if id == "" {
return EntryID{}, ErrInvalidEntryID
}
return EntryID{value: id}, nil
}

// String returns the string representation.
func (e EntryID) String() string {
return e.value
}

// IsValid checks if the ID is valid.
func (e EntryID) IsValid() bool {
return e.value != ""
}

// Equals checks equality with another EntryID.
func (e EntryID) Equals(other EntryID) bool {
return e.value == other.value
}

// ArchiveFormat represents the archive file format.
type ArchiveFormat string

const (
FormatZIP    ArchiveFormat = "zip"
FormatTAR    ArchiveFormat = "tar"
FormatTARGZ  ArchiveFormat = "tar.gz"
FormatTARBZ2 ArchiveFormat = "tar.bz2"
FormatTARXZ  ArchiveFormat = "tar.xz"
FormatAuto   ArchiveFormat = "auto" // Detect from file extension
)

// String returns the string representation.
func (f ArchiveFormat) String() string {
return string(f)
}

// IsValid checks if the format is valid.
func (f ArchiveFormat) IsValid() bool {
switch f {
case FormatZIP, FormatTAR, FormatTARGZ, FormatTARBZ2, FormatTARXZ, FormatAuto:
return true
default:
return false
}
}

// Extension returns the file extension for the format.
func (f ArchiveFormat) Extension() string {
switch f {
case FormatZIP:
return ".zip"
case FormatTAR:
return ".tar"
case FormatTARGZ:
return ".tar.gz"
case FormatTARBZ2:
return ".tar.bz2"
case FormatTARXZ:
return ".tar.xz"
default:
return ""
}
}

// DetectFormat detects the archive format from file extension.
func DetectFormat(filename string) (ArchiveFormat, error) {
switch {
case len(filename) > 7 && filename[len(filename)-7:] == ".tar.gz":
return FormatTARGZ, nil
case len(filename) > 8 && filename[len(filename)-8:] == ".tar.bz2":
return FormatTARBZ2, nil
case len(filename) > 7 && filename[len(filename)-7:] == ".tar.xz":
return FormatTARXZ, nil
case len(filename) > 4 && filename[len(filename)-4:] == ".tar":
return FormatTAR, nil
case len(filename) > 4 && filename[len(filename)-4:] == ".zip":
return FormatZIP, nil
default:
return "", ErrUnsupportedFormat
}
}

// CompressionLevel represents the compression level (0-9).
type CompressionLevel int

const (
CompressionNone    CompressionLevel = 0
CompressionFastest CompressionLevel = 1
CompressionDefault CompressionLevel = 6
CompressionBest    CompressionLevel = 9
)

// String returns the string representation.
func (c CompressionLevel) String() string {
switch c {
case CompressionNone:
return "none"
case CompressionFastest:
return "fastest"
case CompressionDefault:
return "default"
case CompressionBest:
return "best"
default:
return fmt.Sprintf("level-%d", c)
}
}

// IsValid checks if the compression level is valid.
func (c CompressionLevel) IsValid() bool {
return c >= 0 && c <= 9
}

// MediaType represents the type of media file in an archive.
type MediaType string

const (
TypeImage     MediaType = "image"
TypeAudio     MediaType = "audio"
TypeText      MediaType = "text"
TypeVideo     MediaType = "video"
TypeDocument  MediaType = "document"
TypeDirectory MediaType = "directory"
TypeUnknown   MediaType = "unknown"
)

// String returns the string representation.
func (m MediaType) String() string {
return string(m)
}

// IsValid checks if the media type is valid.
func (m MediaType) IsValid() bool {
switch m {
case TypeImage, TypeAudio, TypeText, TypeVideo, TypeDocument, TypeDirectory, TypeUnknown:
return true
default:
return false
}
}

// DetectMediaType detects the media type from file extension.
func DetectMediaType(filename string) MediaType {
ext := getExtension(filename)
switch ext {
case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp":
return TypeImage
case ".wav", ".mp3", ".flac", ".ogg", ".m4a":
return TypeAudio
case ".txt", ".md", ".csv", ".json", ".xml", ".yaml", ".yml":
return TypeText
case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
return TypeVideo
case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
return TypeDocument
default:
return TypeUnknown
}
}

// getExtension extracts the file extension.
func getExtension(filename string) string {
for i := len(filename) - 1; i >= 0; i-- {
if filename[i] == '.' {
return filename[i:]
}
if filename[i] == '/' || filename[i] == '\\' {
break
}
}
return ""
}

// EntryPermissions represents file permissions for an archive entry.
type EntryPermissions struct {
Mode os.FileMode
}

// NewEntryPermissions creates new permissions.
func NewEntryPermissions(mode os.FileMode) EntryPermissions {
return EntryPermissions{Mode: mode}
}

// IsExecutable checks if the entry is executable.
func (p EntryPermissions) IsExecutable() bool {
return p.Mode&0111 != 0
}

// IsReadable checks if the entry is readable.
func (p EntryPermissions) IsReadable() bool {
return p.Mode&0444 != 0
}

// IsWritable checks if the entry is writable.
func (p EntryPermissions) IsWritable() bool {
return p.Mode&0222 != 0
}

// String returns the string representation.
func (p EntryPermissions) String() string {
return p.Mode.String()
}

// ArchiveMetadata contains metadata about an archive.
type ArchiveMetadata struct {
TotalFiles       int
TotalSize        int64
CompressedSize   int64
CompressionRatio float64
CreatedAt        string
ExtractedAt      string
}

// ExtractionOptions configures archive extraction behavior.
type ExtractionOptions struct {
OutputDirectory     string
PreservePaths       bool
PreservePermissions bool
AllowAbsolutePaths  bool
MaxFileSize         int64
MaxTotalSize        int64
ZipBombThreshold    float64
}

// DefaultExtractionOptions returns secure default extraction options.
func DefaultExtractionOptions() ExtractionOptions {
return ExtractionOptions{
OutputDirectory:     ".",
PreservePaths:       true,
PreservePermissions: true,
AllowAbsolutePaths:  false,                         // Security: prevent path traversal
MaxFileSize:         100 * 1024 * 1024,             // 100MB per file
MaxTotalSize:        1024 * 1024 * 1024,            // 1GB total
ZipBombThreshold:    0.01,                          // 1% compression ratio threshold
}
}

// CreationOptions configures archive creation behavior.
type CreationOptions struct {
Format           ArchiveFormat
CompressionLevel CompressionLevel
IncludeHidden    bool
FollowSymlinks   bool
PreservePaths    bool
BaseDirectory    string
}

// DefaultCreationOptions returns default creation options.
func DefaultCreationOptions() CreationOptions {
return CreationOptions{
Format:           FormatTARGZ,
CompressionLevel: CompressionDefault,
IncludeHidden:    false,
FollowSymlinks:   false,
PreservePaths:    true,
BaseDirectory:    "",
}
}
