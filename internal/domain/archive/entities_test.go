package archive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Archive Aggregate Tests
// ============================================================================

func TestArchive_NewArchive_Success(t *testing.T) {
	tests := []struct {
		name             string
		format           ArchiveFormat
		compressionLevel CompressionLevel
	}{
		{"zip_default", FormatZIP, CompressionDefault},
		{"tar_gz_best", FormatTARGZ, CompressionBest},
		{"tar_fastest", FormatTAR, CompressionFastest},
		{"tar_bz2_none", FormatTARBZ2, CompressionNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := NewArchive(tt.format, tt.compressionLevel)

			require.NoError(t, err)
			assert.NotNil(t, archive)
			assert.True(t, archive.ID.IsValid())
			assert.Equal(t, tt.format, archive.Format)
			assert.Equal(t, tt.compressionLevel, archive.CompressionLevel)
			assert.Empty(t, archive.Entries)
			assert.Equal(t, int64(0), archive.TotalSize)
			assert.Equal(t, int64(0), archive.CompressedSize)
			assert.Equal(t, 0, archive.EntryCount)
			assert.WithinDuration(t, time.Now(), archive.CreatedAt, time.Second)
			assert.Nil(t, archive.ExtractedAt)
		})
	}
}

func TestArchive_NewArchive_InvalidFormat(t *testing.T) {
	archive, err := NewArchive(ArchiveFormat("invalid"), CompressionDefault)

	assert.ErrorIs(t, err, ErrInvalidFormat)
	assert.Nil(t, archive)
}

func TestArchive_NewArchive_InvalidCompressionLevel(t *testing.T) {
	archive, err := NewArchive(FormatZIP, CompressionLevel(99))

	assert.ErrorIs(t, err, ErrInvalidCompressionLevel)
	assert.Nil(t, archive)
}

func TestArchive_AddEntry_Success(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry, _ := NewArchiveEntry(archive.ID, "test.txt", "files/test.txt", TypeText, 1024)
	entry.CompressedSize = 512

	// Act
	err := archive.AddEntry(entry)

	// Assert
	require.NoError(t, err)
	assert.Len(t, archive.Entries, 1)
	assert.Equal(t, int64(1024), archive.TotalSize)
	assert.Equal(t, int64(512), archive.CompressedSize)
	assert.Equal(t, 1, archive.EntryCount)
}

func TestArchive_AddEntry_MultipleEntries(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry1, _ := NewArchiveEntry(archive.ID, "file1.txt", "file1.txt", TypeText, 1024)
	entry2, _ := NewArchiveEntry(archive.ID, "file2.png", "file2.png", TypeImage, 2048)
	entry3, _ := NewArchiveEntry(archive.ID, "file3.wav", "file3.wav", TypeAudio, 4096)

	// Act
	err1 := archive.AddEntry(entry1)
	err2 := archive.AddEntry(entry2)
	err3 := archive.AddEntry(entry3)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.Len(t, archive.Entries, 3)
	assert.Equal(t, int64(7168), archive.TotalSize) // 1024+2048+4096
	assert.Equal(t, 3, archive.EntryCount)
}

func TestArchive_AddEntry_NilEntry(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)

	err := archive.AddEntry(nil)

	assert.ErrorIs(t, err, ErrInvalidEntry)
}

func TestArchive_AddEntry_DuplicatePath(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry1, _ := NewArchiveEntry(archive.ID, "test.txt", "files/test.txt", TypeText, 1024)
	entry2, _ := NewArchiveEntry(archive.ID, "test.txt", "files/test.txt", TypeText, 2048)

	// Act
	err1 := archive.AddEntry(entry1)
	err2 := archive.AddEntry(entry2)

	// Assert
	require.NoError(t, err1)
	assert.ErrorIs(t, err2, ErrEntryAlreadyExists)
	assert.Len(t, archive.Entries, 1)
}

func TestArchive_Extract_Success(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)

	err := archive.Extract()

	require.NoError(t, err)
	require.NotNil(t, archive.ExtractedAt)
	assert.WithinDuration(t, time.Now(), *archive.ExtractedAt, time.Second)
}

func TestArchive_Validate_Success(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry, _ := NewArchiveEntry(archive.ID, "test.txt", "test.txt", TypeText, 1024)
	archive.AddEntry(entry)

	// Act
	err := archive.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestArchive_Validate_InvalidID(t *testing.T) {
	archive := &Archive{
		ID:         ArchiveID{},
		Format:     FormatZIP,
		EntryCount: 1,
		TotalSize:  1024,
	}

	err := archive.Validate()

	assert.ErrorIs(t, err, ErrInvalidArchiveID)
}

func TestArchive_Validate_InvalidFormat(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	archive.Format = ArchiveFormat("invalid")
	entry, _ := NewArchiveEntry(archive.ID, "test.txt", "test.txt", TypeText, 1024)
	archive.AddEntry(entry)

	err := archive.Validate()

	assert.ErrorIs(t, err, ErrInvalidFormat)
}

func TestArchive_Validate_Empty(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)

	err := archive.Validate()

	assert.ErrorIs(t, err, ErrArchiveEmpty)
}

func TestArchive_Validate_Corrupted(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry, _ := NewArchiveEntry(archive.ID, "test.txt", "test.txt", TypeText, 1024)
	archive.AddEntry(entry)
	archive.TotalSize = 0 // Corrupt the size

	err := archive.Validate()

	assert.ErrorIs(t, err, ErrArchiveCorrupted)
}

func TestArchive_GetEntry_Success(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry1, _ := NewArchiveEntry(archive.ID, "file1.txt", "file1.txt", TypeText, 1024)
	entry2, _ := NewArchiveEntry(archive.ID, "file2.txt", "file2.txt", TypeText, 2048)
	archive.AddEntry(entry1)
	archive.AddEntry(entry2)

	// Act
	found, err := archive.GetEntry(entry2.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entry2.ID, found.ID)
	assert.Equal(t, "file2.txt", found.Name)
}

func TestArchive_GetEntry_NotFound(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	nonExistentID := NewEntryID()

	entry, err := archive.GetEntry(nonExistentID)

	assert.ErrorIs(t, err, ErrEntryNotFound)
	assert.Nil(t, entry)
}

func TestArchive_GetEntries_ReturnsAll(t *testing.T) {
	// Arrange
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	entry1, _ := NewArchiveEntry(archive.ID, "file1.txt", "file1.txt", TypeText, 1024)
	entry2, _ := NewArchiveEntry(archive.ID, "file2.txt", "file2.txt", TypeText, 2048)
	entry3, _ := NewArchiveEntry(archive.ID, "file3.txt", "file3.txt", TypeText, 4096)
	archive.AddEntry(entry1)
	archive.AddEntry(entry2)
	archive.AddEntry(entry3)

	// Act
	entries := archive.GetEntries()

	// Assert
	assert.Len(t, entries, 3)
}

func TestArchive_CalculateCompressionRatio(t *testing.T) {
	tests := []struct {
		name           string
		totalSize      int64
		compressedSize int64
		expectedRatio  float64
	}{
		{"50_percent", 1000, 500, 0.5},
		{"25_percent", 2000, 500, 0.25},
		{"90_percent", 1000, 900, 0.9},
		{"no_compression", 1000, 1000, 1.0},
		{"zero_total", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, _ := NewArchive(FormatZIP, CompressionDefault)
			archive.TotalSize = tt.totalSize
			archive.CompressedSize = tt.compressedSize

			ratio := archive.CalculateCompressionRatio()

			assert.InDelta(t, tt.expectedRatio, ratio, 0.001)
		})
	}
}

func TestArchive_DetectZipBomb_NotDetected(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	archive.TotalSize = 1000
	archive.CompressedSize = 500 // 50% ratio, above 1% threshold

	isZipBomb := archive.DetectZipBomb(0.01)

	assert.False(t, isZipBomb)
}

func TestArchive_DetectZipBomb_Detected(t *testing.T) {
	archive, _ := NewArchive(FormatZIP, CompressionDefault)
	archive.TotalSize = 100000000  // 100MB uncompressed
	archive.CompressedSize = 50000 // 50KB compressed (0.0005 ratio)

	isZipBomb := archive.DetectZipBomb(0.01) // 1% threshold

	assert.True(t, isZipBomb)
}

// ============================================================================
// ArchiveEntry Entity Tests
// ============================================================================

func TestArchiveEntry_NewArchiveEntry_Success(t *testing.T) {
	archiveID := NewArchiveID()

	entry, err := NewArchiveEntry(archiveID, "test.txt", "files/test.txt", TypeText, 1024)

	require.NoError(t, err)
	assert.True(t, entry.ID.IsValid())
	assert.True(t, entry.ArchiveID.Equals(archiveID))
	assert.Equal(t, "test.txt", entry.Name)
	assert.Equal(t, "files/test.txt", entry.Path)
	assert.Equal(t, TypeText, entry.MediaType)
	assert.Equal(t, int64(1024), entry.Size)
	assert.WithinDuration(t, time.Now(), entry.ModifiedAt, time.Second)
	assert.Nil(t, entry.ExtractedAt)
}

func TestArchiveEntry_NewArchiveEntry_InvalidArchiveID(t *testing.T) {
	entry, err := NewArchiveEntry(ArchiveID{}, "test.txt", "test.txt", TypeText, 1024)

	assert.ErrorIs(t, err, ErrInvalidArchiveID)
	assert.Nil(t, entry)
}

func TestArchiveEntry_NewArchiveEntry_EmptyName(t *testing.T) {
	archiveID := NewArchiveID()

	entry, err := NewArchiveEntry(archiveID, "", "test.txt", TypeText, 1024)

	assert.ErrorIs(t, err, ErrInvalidEntryName)
	assert.Nil(t, entry)
}

func TestArchiveEntry_NewArchiveEntry_EmptyPath(t *testing.T) {
	archiveID := NewArchiveID()

	entry, err := NewArchiveEntry(archiveID, "test.txt", "", TypeText, 1024)

	assert.ErrorIs(t, err, ErrInvalidPath)
	assert.Nil(t, entry)
}

func TestArchiveEntry_NewArchiveEntry_InvalidMediaType(t *testing.T) {
	archiveID := NewArchiveID()

	entry, err := NewArchiveEntry(archiveID, "test.txt", "test.txt", MediaType("invalid"), 1024)

	assert.ErrorIs(t, err, ErrInvalidMediaType)
	assert.Nil(t, entry)
}

func TestArchiveEntry_NewArchiveEntry_NegativeSize(t *testing.T) {
	archiveID := NewArchiveID()

	entry, err := NewArchiveEntry(archiveID, "test.txt", "test.txt", TypeText, -100)

	assert.ErrorIs(t, err, ErrInvalidEntrySize)
	assert.Nil(t, entry)
}

func TestArchiveEntry_IsDirectory_True(t *testing.T) {
	archiveID := NewArchiveID()
	entry, _ := NewArchiveEntry(archiveID, "dir", "files/dir", TypeDirectory, 0)

	assert.True(t, entry.IsDirectory())
	assert.False(t, entry.IsFile())
}

func TestArchiveEntry_IsDirectory_False(t *testing.T) {
	archiveID := NewArchiveID()
	entry, _ := NewArchiveEntry(archiveID, "test.txt", "test.txt", TypeText, 1024)

	assert.False(t, entry.IsDirectory())
	assert.True(t, entry.IsFile())
}

func TestArchiveEntry_Validate_Success(t *testing.T) {
	archiveID := NewArchiveID()
	entry, _ := NewArchiveEntry(archiveID, "test.txt", "test.txt", TypeText, 1024)

	err := entry.Validate()

	assert.NoError(t, err)
}

func TestArchiveEntry_Validate_Errors(t *testing.T) {
	archiveID := NewArchiveID()
	validEntry, _ := NewArchiveEntry(archiveID, "test.txt", "test.txt", TypeText, 1024)

	tests := []struct {
		name        string
		mutate      func(*ArchiveEntry)
		expectedErr error
	}{
		{
			"invalid_entry_id",
			func(e *ArchiveEntry) { e.ID = EntryID{} },
			ErrInvalidEntryID,
		},
		{
			"invalid_archive_id",
			func(e *ArchiveEntry) { e.ArchiveID = ArchiveID{} },
			ErrInvalidArchiveID,
		},
		{
			"empty_name",
			func(e *ArchiveEntry) { e.Name = "" },
			ErrInvalidEntryName,
		},
		{
			"empty_path",
			func(e *ArchiveEntry) { e.Path = "" },
			ErrInvalidPath,
		},
		{
			"invalid_media_type",
			func(e *ArchiveEntry) { e.MediaType = MediaType("invalid") },
			ErrInvalidMediaType,
		},
		{
			"negative_size",
			func(e *ArchiveEntry) { e.Size = -1 },
			ErrInvalidEntrySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := *validEntry // Copy
			tt.mutate(&entry)

			err := entry.Validate()

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestArchiveEntry_MarkExtracted(t *testing.T) {
	archiveID := NewArchiveID()
	entry, _ := NewArchiveEntry(archiveID, "test.txt", "test.txt", TypeText, 1024)
	assert.Nil(t, entry.ExtractedAt)

	entry.MarkExtracted()

	require.NotNil(t, entry.ExtractedAt)
	assert.WithinDuration(t, time.Now(), *entry.ExtractedAt, time.Second)
}
