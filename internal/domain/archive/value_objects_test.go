package archive

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ArchiveID Tests
// ============================================================================

func TestArchiveID_NewArchiveID(t *testing.T) {
	id := NewArchiveID()

	assert.True(t, id.IsValid())
	assert.NotEmpty(t, id.String())
}

func TestArchiveID_NewArchiveIDFromString_Valid(t *testing.T) {
	original := NewArchiveID()

	id, err := NewArchiveIDFromString(original.String())

	require.NoError(t, err)
	assert.True(t, id.IsValid())
	assert.Equal(t, original.String(), id.String())
}

func TestArchiveID_NewArchiveIDFromString_Empty(t *testing.T) {
	id, err := NewArchiveIDFromString("")

	assert.ErrorIs(t, err, ErrInvalidArchiveID)
	assert.False(t, id.IsValid())
}

func TestArchiveID_Equals(t *testing.T) {
	id1 := NewArchiveID()
	id2 := NewArchiveID()

	assert.True(t, id1.Equals(id1))
	assert.False(t, id1.Equals(id2))
}

// ============================================================================
// EntryID Tests
// ============================================================================

func TestEntryID_NewEntryID(t *testing.T) {
	id := NewEntryID()

	assert.True(t, id.IsValid())
	assert.NotEmpty(t, id.String())
}

func TestEntryID_NewEntryIDFromString_Valid(t *testing.T) {
	original := NewEntryID()

	id, err := NewEntryIDFromString(original.String())

	require.NoError(t, err)
	assert.True(t, id.IsValid())
	assert.Equal(t, original.String(), id.String())
}

func TestEntryID_NewEntryIDFromString_Empty(t *testing.T) {
	id, err := NewEntryIDFromString("")

	assert.ErrorIs(t, err, ErrInvalidEntryID)
	assert.False(t, id.IsValid())
}

func TestEntryID_Equals(t *testing.T) {
	id1 := NewEntryID()
	id2 := NewEntryID()

	assert.True(t, id1.Equals(id1))
	assert.False(t, id1.Equals(id2))
}

// ============================================================================
// ArchiveFormat Tests
// ============================================================================

func TestArchiveFormat_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		format  ArchiveFormat
		isValid bool
	}{
		{"zip", FormatZIP, true},
		{"tar", FormatTAR, true},
		{"tar_gz", FormatTARGZ, true},
		{"tar_bz2", FormatTARBZ2, true},
		{"tar_xz", FormatTARXZ, true},
		{"auto", FormatAuto, true},
		{"invalid", ArchiveFormat("invalid"), false},
		{"empty", ArchiveFormat(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.format.IsValid())
		})
	}
}

func TestArchiveFormat_String(t *testing.T) {
	assert.Equal(t, "zip", FormatZIP.String())
	assert.Equal(t, "tar", FormatTAR.String())
	assert.Equal(t, "tar.gz", FormatTARGZ.String())
	assert.Equal(t, "tar.bz2", FormatTARBZ2.String())
	assert.Equal(t, "tar.xz", FormatTARXZ.String())
	assert.Equal(t, "auto", FormatAuto.String())
}

func TestArchiveFormat_Extension(t *testing.T) {
	tests := []struct {
		format    ArchiveFormat
		extension string
	}{
		{FormatZIP, ".zip"},
		{FormatTAR, ".tar"},
		{FormatTARGZ, ".tar.gz"},
		{FormatTARBZ2, ".tar.bz2"},
		{FormatTARXZ, ".tar.xz"},
		{FormatAuto, ""},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			assert.Equal(t, tt.extension, tt.format.Extension())
		})
	}
}

func TestArchiveFormat_DetectFormat(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		expectedFormat ArchiveFormat
		expectError    bool
	}{
		{"zip_file", "archive.zip", FormatZIP, false},
		{"tar_file", "archive.tar", FormatTAR, false},
		{"tar_gz_file", "archive.tar.gz", FormatTARGZ, false},
		{"tar_bz2_file", "backup.tar.bz2", FormatTARBZ2, false},
		{"tar_xz_file", "data.tar.xz", FormatTARXZ, false},
		{"path_with_tar_gz", "/path/to/file.tar.gz", FormatTARGZ, false},

		{"unsupported", "file.rar", ArchiveFormat(""), true},
		{"no_extension", "archive", ArchiveFormat(""), true},
		{"short_name", "a.z", ArchiveFormat(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := DetectFormat(tt.filename)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedFormat, format)
			}
		})
	}
}

// ============================================================================
// CompressionLevel Tests
// ============================================================================

func TestCompressionLevel_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		level   CompressionLevel
		isValid bool
	}{
		{"min", 0, true},
		{"level_1", 1, true},
		{"level_5", 5, true},
		{"level_9", 9, true},
		{"negative", -1, false},
		{"too_high", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.level.IsValid())
		})
	}
}

func TestCompressionLevel_String(t *testing.T) {
	tests := []struct {
		level    CompressionLevel
		expected string
	}{
		{CompressionNone, "none"},
		{CompressionFastest, "fastest"},
		{CompressionDefault, "default"},
		{CompressionBest, "best"},
		{CompressionLevel(2), "level-2"},
		{CompressionLevel(5), "level-5"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

// ============================================================================
// MediaType Tests
// ============================================================================

func TestMediaType_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		mt      MediaType
		isValid bool
	}{
		{"image", TypeImage, true},
		{"audio", TypeAudio, true},
		{"text", TypeText, true},
		{"video", TypeVideo, true},
		{"document", TypeDocument, true},
		{"directory", TypeDirectory, true},
		{"unknown", TypeUnknown, true},
		{"invalid", MediaType("invalid"), false},
		{"empty", MediaType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.mt.IsValid())
		})
	}
}

func TestMediaType_String(t *testing.T) {
	assert.Equal(t, "image", TypeImage.String())
	assert.Equal(t, "audio", TypeAudio.String())
	assert.Equal(t, "text", TypeText.String())
	assert.Equal(t, "video", TypeVideo.String())
	assert.Equal(t, "document", TypeDocument.String())
	assert.Equal(t, "directory", TypeDirectory.String())
	assert.Equal(t, "unknown", TypeUnknown.String())
}

func TestMediaType_DetectMediaType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType MediaType
	}{
		// Images
		{"photo.png", TypeImage},
		{"image.jpg", TypeImage},
		{"picture.jpeg", TypeImage},
		{"animation.gif", TypeImage},
		{"icon.bmp", TypeImage},
		{"modern.webp", TypeImage},

		// Audio
		{"song.wav", TypeAudio},
		{"music.mp3", TypeAudio},
		{"track.flac", TypeAudio},
		{"audio.ogg", TypeAudio},
		{"file.m4a", TypeAudio},

		// Text
		{"readme.txt", TypeText},
		{"notes.md", TypeText},
		{"data.csv", TypeText},
		{"config.json", TypeText},
		{"schema.xml", TypeText},
		{"settings.yaml", TypeText},
		{"deploy.yml", TypeText},

		// Video
		{"movie.mp4", TypeVideo},
		{"clip.avi", TypeVideo},
		{"video.mkv", TypeVideo},
		{"recording.mov", TypeVideo},
		{"film.wmv", TypeVideo},

		// Documents
		{"report.pdf", TypeDocument},
		{"letter.doc", TypeDocument},
		{"document.docx", TypeDocument},
		{"spreadsheet.xls", TypeDocument},
		{"workbook.xlsx", TypeDocument},
		{"presentation.ppt", TypeDocument},
		{"slides.pptx", TypeDocument},

		// Unknown
		{"archive.zip", TypeUnknown},
		{"binary.bin", TypeUnknown},
		{"executable.exe", TypeUnknown},
		{"noext", TypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			mt := DetectMediaType(tt.filename)
			assert.Equal(t, tt.expectedType, mt)
		})
	}
}

// ============================================================================
// EntryPermissions Tests
// ============================================================================

func TestEntryPermissions_NewEntryPermissions(t *testing.T) {
	perms := NewEntryPermissions(0755)

	assert.Equal(t, os.FileMode(0755), perms.Mode)
}

func TestEntryPermissions_IsExecutable(t *testing.T) {
	tests := []struct {
		name       string
		mode       os.FileMode
		executable bool
	}{
		{"755_rwxr_xr_x", 0755, true},
		{"644_rw_r__r__", 0644, false},
		{"111_all_exec", 0111, true},
		{"100_owner_exec", 0100, true},
		{"010_group_exec", 0010, true},
		{"001_other_exec", 0001, true},
		{"444_no_exec", 0444, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := NewEntryPermissions(tt.mode)
			assert.Equal(t, tt.executable, perms.IsExecutable())
		})
	}
}

func TestEntryPermissions_IsReadable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		readable bool
	}{
		{"644_rw_r__r__", 0644, true},
		{"444_all_read", 0444, true},
		{"400_owner_read", 0400, true},
		{"040_group_read", 0040, true},
		{"004_other_read", 0004, true},
		{"200_write_only", 0200, false},
		{"111_exec_only", 0111, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := NewEntryPermissions(tt.mode)
			assert.Equal(t, tt.readable, perms.IsReadable())
		})
	}
}

func TestEntryPermissions_IsWritable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		writable bool
	}{
		{"644_rw_r__r__", 0644, true},
		{"222_all_write", 0222, true},
		{"200_owner_write", 0200, true},
		{"020_group_write", 0020, true},
		{"002_other_write", 0002, true},
		{"444_read_only", 0444, false},
		{"111_exec_only", 0111, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := NewEntryPermissions(tt.mode)
			assert.Equal(t, tt.writable, perms.IsWritable())
		})
	}
}

func TestEntryPermissions_String(t *testing.T) {
	perms := NewEntryPermissions(0755)

	str := perms.String()

	assert.NotEmpty(t, str)
	assert.Contains(t, str, "rwx")
}

// ============================================================================
// Helper Struct Tests
// ============================================================================

func TestDefaultExtractionOptions(t *testing.T) {
	opts := DefaultExtractionOptions()

	assert.Equal(t, ".", opts.OutputDirectory)
	assert.True(t, opts.PreservePaths)
	assert.True(t, opts.PreservePermissions)
	assert.False(t, opts.AllowAbsolutePaths)                  // Security: prevent path traversal
	assert.Equal(t, int64(100*1024*1024), opts.MaxFileSize)   // 100MB
	assert.Equal(t, int64(1024*1024*1024), opts.MaxTotalSize) // 1GB
	assert.InDelta(t, 0.01, opts.ZipBombThreshold, 0.001)     // 1%
}

func TestDefaultCreationOptions(t *testing.T) {
	opts := DefaultCreationOptions()

	assert.Equal(t, FormatTARGZ, opts.Format)
	assert.Equal(t, CompressionDefault, opts.CompressionLevel)
	assert.False(t, opts.IncludeHidden)
	assert.False(t, opts.FollowSymlinks)
	assert.True(t, opts.PreservePaths)
	assert.Equal(t, "", opts.BaseDirectory)
}
