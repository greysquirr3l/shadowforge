package archive

import "context"

// ArchiveService provides domain operations for archive management.
type ArchiveService struct {
repo ArchiveRepository
}

// NewArchiveService creates a new archive service.
func NewArchiveService(repo ArchiveRepository) *ArchiveService {
return &ArchiveService{
repo: repo,
}
}

// CreateArchive creates a new archive with validation.
func (s *ArchiveService) CreateArchive(ctx context.Context, format ArchiveFormat, compressionLevel CompressionLevel) (*Archive, error) {
return NewArchive(format, compressionLevel)
}

// AddEntryToArchive adds an entry to an archive.
func (s *ArchiveService) AddEntryToArchive(ctx context.Context, archive *Archive, entry *ArchiveEntry) error {
return archive.AddEntry(entry)
}

// ValidateArchive performs comprehensive validation.
func (s *ArchiveService) ValidateArchive(ctx context.Context, archive *Archive) error {
return archive.Validate()
}

// CheckZipBomb checks if archive is a potential zip bomb.
func (s *ArchiveService) CheckZipBomb(ctx context.Context, archive *Archive, threshold float64) bool {
return archive.DetectZipBomb(threshold)
}

// ExtractArchive marks archive as extracted.
func (s *ArchiveService) ExtractArchive(ctx context.Context, archive *Archive) error {
return archive.Extract()
}

// GetArchiveEntry retrieves an entry from archive.
func (s *ArchiveService) GetArchiveEntry(ctx context.Context, archive *Archive, entryID EntryID) (*ArchiveEntry, error) {
return archive.GetEntry(entryID)
}

// ListArchiveEntries returns all entries in archive.
func (s *ArchiveService) ListArchiveEntries(ctx context.Context, archive *Archive) []*ArchiveEntry {
return archive.GetEntries()
}

// CalculateCompressionRatio calculates the compression ratio.
func (s *ArchiveService) CalculateCompressionRatio(ctx context.Context, archive *Archive) float64 {
return archive.CalculateCompressionRatio()
}

// DetectMediaType detects the media type of a file.
func (s *ArchiveService) DetectMediaType(ctx context.Context, filename string) MediaType {
return DetectMediaType(filename)
}
