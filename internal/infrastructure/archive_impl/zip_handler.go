package archive_impl

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/archive"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// ZIPHandler implements ZIP archive operations with security protections
type ZIPHandler struct {
	maxFileSize        int64   // Maximum size for individual files
	maxTotalSize       int64   // Maximum size for entire archive
	maxFiles           int     // Maximum number of files in archive
	maxCompressionRate float64 // Maximum compression ratio (anti zip-bomb)
}

// NewZIPHandler creates a new ZIP handler with default security limits
func NewZIPHandler() *ZIPHandler {
	return &ZIPHandler{
		maxFileSize:        100 * 1024 * 1024,  // 100MB per file
		maxTotalSize:       1024 * 1024 * 1024, // 1GB total
		maxFiles:           10000,              // 10,000 files max
		maxCompressionRate: 100.0,              // 100:1 compression ratio
	}
}

// NewCustomZIPHandler creates a ZIP handler with custom limits
func NewCustomZIPHandler(maxFileSize, maxTotalSize int64, maxFiles int, maxCompressionRate float64) *ZIPHandler {
	return &ZIPHandler{
		maxFileSize:        maxFileSize,
		maxTotalSize:       maxTotalSize,
		maxFiles:           maxFiles,
		maxCompressionRate: maxCompressionRate,
	}
}

// Extract extracts a ZIP archive with security validation
func (h *ZIPHandler) Extract(archiveData []byte, outputDir string) (*archive.Archive, error) {
	logger.Log.WithFields(map[string]interface{}{
		"output_dir":   outputDir,
		"archive_size": len(archiveData),
	}).Info("Extracting ZIP archive")

	// Create zip reader
	reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Validate archive security
	if err := h.validateArchive(reader); err != nil {
		return nil, fmt.Errorf("archive validation failed: %w", err)
	}

	// Create output directory
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Extract files
	entries := make([]*archive.ArchiveEntry, 0, len(reader.File))
	var totalExtracted int64

	for _, file := range reader.File {
		entry, size, err := h.extractFile(file, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}

		entries = append(entries, entry)
		totalExtracted += size

		logger.Log.WithFields(map[string]interface{}{
			"file": file.Name,
			"size": size,
		}).Debug("Extracted file from ZIP")
	}

	// Create archive entity
	extractedAt := time.Now()
	archiveEntity := &archive.Archive{
		Format:      archive.FormatZIP,
		TotalSize:   totalExtracted,
		EntryCount:  len(entries),
		Entries:     entries,
		ExtractedAt: &extractedAt,
	}

	logger.Log.WithFields(map[string]interface{}{
		"total_files": len(entries),
		"total_size":  totalExtracted,
	}).Info("ZIP extraction complete")

	return archiveEntity, nil
}

// Create creates a ZIP archive from the given entries
func (h *ZIPHandler) Create(archiveEntry *archive.Archive, outputPath string) error {
	logger.Log.WithFields(map[string]interface{}{
		"output_path": outputPath,
		"file_count":  len(archiveEntry.Entries),
	}).Info("Creating ZIP archive")

	// Validate inputs
	if len(archiveEntry.Entries) == 0 {
		return fmt.Errorf("no entries to archive")
	}

	if len(archiveEntry.Entries) > h.maxFiles {
		return fmt.Errorf("too many files: %d exceeds limit %d", len(archiveEntry.Entries), h.maxFiles)
	}

	// Create output buffer
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Add files to archive
	var totalSize int64
	for _, entry := range archiveEntry.Entries {
		if entry.IsDirectory() {
			// Create directory entry
			header := &zip.FileHeader{
				Name:   entry.Path + "/",
				Method: zip.Deflate,
			}
			header.SetModTime(entry.ModifiedAt)

			if _, err := writer.CreateHeader(header); err != nil {
				return fmt.Errorf("failed to create directory header: %w", err)
			}
		} else {
			// Check file size limits
			if entry.Size > h.maxFileSize {
				return fmt.Errorf("file %s too large: %d exceeds limit %d", entry.Path, entry.Size, h.maxFileSize)
			}

			totalSize += entry.Size
			if totalSize > h.maxTotalSize {
				return fmt.Errorf("total size exceeds limit: %d > %d", totalSize, h.maxTotalSize)
			}

			// Create file header
			header := &zip.FileHeader{
				Name:   entry.Path,
				Method: zip.Deflate,
			}
			header.SetModTime(entry.ModifiedAt)

			// Write file content
			fileWriter, err := writer.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("failed to create file header: %w", err)
			}

			if _, err := fileWriter.Write(entry.Content); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}

			logger.Log.WithFields(map[string]interface{}{
				"file": entry.Path,
				"size": entry.Size,
			}).Debug("Added file to ZIP")
		}
	}

	// Close zip writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	logger.Log.WithFields(map[string]interface{}{
		"file_count":  len(archiveEntry.Entries),
		"total_size":  totalSize,
		"output_path": outputPath,
	}).Info("ZIP archive created successfully")

	return nil
}

// validateArchive performs security checks on the ZIP archive
func (h *ZIPHandler) validateArchive(reader *zip.Reader) error {
	// Check file count
	if len(reader.File) > h.maxFiles {
		return fmt.Errorf("too many files in archive: %d exceeds limit %d", len(reader.File), h.maxFiles)
	}

	var totalUncompressed int64
	var totalCompressed int64

	for _, file := range reader.File {
		// Check individual file size
		if file.UncompressedSize64 > uint64(h.maxFileSize) {
			return fmt.Errorf("file %s too large: %d exceeds limit %d", file.Name, file.UncompressedSize64, h.maxFileSize)
		}

		totalUncompressed += int64(file.UncompressedSize64)
		totalCompressed += int64(file.CompressedSize64)

		// Validate path (zip-slip protection)
		if err := h.validatePath(file.Name); err != nil {
			return fmt.Errorf("invalid path in archive: %w", err)
		}
	}

	// Check total size
	if totalUncompressed > h.maxTotalSize {
		return fmt.Errorf("total uncompressed size %d exceeds limit %d", totalUncompressed, h.maxTotalSize)
	}

	// Check compression ratio (anti zip-bomb)
	if totalCompressed > 0 {
		compressionRatio := float64(totalUncompressed) / float64(totalCompressed)
		if compressionRatio > h.maxCompressionRate {
			return fmt.Errorf("suspicious compression ratio: %.2f exceeds limit %.2f (potential zip bomb)", compressionRatio, h.maxCompressionRate)
		}
	}

	return nil
}

// validatePath validates a file path to prevent zip-slip attacks
func (h *ZIPHandler) validatePath(path string) error {
	// Check for absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute path not allowed: %s", path)
	}

	// Check for path traversal attempts
	cleaned := filepath.Clean(path)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
		return fmt.Errorf("path traversal detected: %s", path)
	}

	// Check for suspicious characters
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null byte in path: %s", path)
	}

	return nil
}

// extractFile extracts a single file from the ZIP archive
func (h *ZIPHandler) extractFile(file *zip.File, outputDir string) (*archive.ArchiveEntry, int64, error) {
	// Validate path (zip-slip protection)
	if err := h.validatePath(file.Name); err != nil {
		return nil, 0, err
	}

	// Open file from archive
	reader, err := file.Open()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer reader.Close()

	// Read file content with size limit
	var content []byte
	if !file.FileInfo().IsDir() {
		content, err = io.ReadAll(io.LimitReader(reader, h.maxFileSize))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read file content: %w", err)
		}
	}

	// Create archive entry
	entry := &archive.ArchiveEntry{
		Path:        file.Name,
		Size:        int64(file.UncompressedSize64),
		Checksum:    fmt.Sprintf("crc32:%x", file.CRC32),
		Permissions: 0644, // Default permissions
		ModifiedAt:  file.Modified,
		ModTime:     file.Modified,
		MediaType:   h.detectMediaType(file.Name, file.FileInfo().IsDir()),
		Content:     content,
	}

	// Write to file system if output directory specified
	if outputDir != "" {
		outputPath := filepath.Join(outputDir, file.Name)

		// Additional validation after joining paths
		cleanOutput := filepath.Clean(outputPath)
		cleanBase := filepath.Clean(outputDir)
		if !strings.HasPrefix(cleanOutput, cleanBase) {
			return nil, 0, fmt.Errorf("path traversal attempt detected: %s escapes %s", file.Name, outputDir)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return nil, 0, fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, 0, fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Write file
			if err := os.WriteFile(outputPath, content, 0644); err != nil {
				return nil, 0, fmt.Errorf("failed to write file: %w", err)
			}
		}

		extractedAt := time.Now()
		entry.ExtractedAt = &extractedAt
	}

	return entry, entry.Size, nil
}

// detectMediaType determines the media type based on file extension and info
func (h *ZIPHandler) detectMediaType(path string, isDir bool) archive.MediaType {
	if isDir {
		return archive.TypeDirectory
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".bmp", ".gif":
		return archive.TypeImage
	case ".wav", ".mp3", ".flac", ".ogg":
		return archive.TypeAudio
	case ".txt", ".md", ".html", ".xml", ".json":
		return archive.TypeText
	case ".zip", ".tar", ".gz", ".bz2":
		return archive.TypeDocument
	default:
		return archive.TypeUnknown
	}
}
