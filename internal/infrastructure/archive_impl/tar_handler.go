package archive_impl

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/domain/archive"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// TARHandler implements TAR archive operations with security protections
type TARHandler struct {
	maxFileSize        int64   // Maximum size for individual files
	maxTotalSize       int64   // Maximum size for entire archive
	maxFiles           int     // Maximum number of files in archive
	maxCompressionRate float64 // Maximum compression ratio (anti tar-bomb)
}

// NewTARHandler creates a new TAR handler with default security limits
func NewTARHandler() *TARHandler {
	return &TARHandler{
		maxFileSize:        100 * 1024 * 1024,  // 100MB per file
		maxTotalSize:       1024 * 1024 * 1024, // 1GB total
		maxFiles:           10000,              // 10,000 files max
		maxCompressionRate: 100.0,              // 100:1 compression ratio
	}
}

// NewCustomTARHandler creates a TAR handler with custom limits
func NewCustomTARHandler(maxFileSize, maxTotalSize int64, maxFiles int, maxCompressionRate float64) *TARHandler {
	return &TARHandler{
		maxFileSize:        maxFileSize,
		maxTotalSize:       maxTotalSize,
		maxFiles:           maxFiles,
		maxCompressionRate: maxCompressionRate,
	}
}

// Extract extracts a TAR archive with security validation
func (h *TARHandler) Extract(archiveData []byte, outputDir string, format archive.ArchiveFormat) (*archive.Archive, error) {
	logger.Log.WithFields(map[string]interface{}{
		"output_dir":   outputDir,
		"archive_size": len(archiveData),
		"format":       format.String(),
	}).Info("Extracting TAR archive")

	// Decompress if needed
	reader, err := h.createReader(archiveData, format)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar reader: %w", err)
	}

	// Create output directory
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Extract files
	tarReader := tar.NewReader(reader)
	entries := make([]*archive.ArchiveEntry, 0)
	var totalExtracted int64
	var fileCount int

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Security checks
		fileCount++
		if fileCount > h.maxFiles {
			return nil, fmt.Errorf("too many files in archive: %d exceeds limit %d", fileCount, h.maxFiles)
		}

		if header.Size > h.maxFileSize {
			return nil, fmt.Errorf("file %s too large: %d exceeds limit %d", header.Name, header.Size, h.maxFileSize)
		}

		totalExtracted += header.Size
		if totalExtracted > h.maxTotalSize {
			return nil, fmt.Errorf("total size exceeds limit: %d > %d", totalExtracted, h.maxTotalSize)
		}

		// Validate path (tar-slip protection)
		if err := h.validatePath(header.Name); err != nil {
			return nil, fmt.Errorf("invalid path in archive: %w", err)
		}

		// Extract file
		entry, err := h.extractFile(tarReader, header, outputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract file %s: %w", header.Name, err)
		}

		entries = append(entries, entry)

		logger.Log.WithFields(map[string]interface{}{
			"file": header.Name,
			"size": header.Size,
		}).Debug("Extracted file from TAR")
	}

	// Create archive entity
	extractedAt := time.Now()
	archiveEntity := &archive.Archive{
		Format:      format,
		TotalSize:   totalExtracted,
		EntryCount:  len(entries),
		Entries:     entries,
		ExtractedAt: &extractedAt,
	}

	logger.Log.WithFields(map[string]interface{}{
		"total_files": len(entries),
		"total_size":  totalExtracted,
	}).Info("TAR extraction complete")

	return archiveEntity, nil
}

// Create creates a TAR archive from the given entries
func (h *TARHandler) Create(archiveEntry *archive.Archive, outputPath string) error {
	logger.Log.WithFields(map[string]interface{}{
		"output_path": outputPath,
		"file_count":  len(archiveEntry.Entries),
		"format":      archiveEntry.Format.String(),
	}).Info("Creating TAR archive")

	// Validate inputs
	if len(archiveEntry.Entries) == 0 {
		return fmt.Errorf("no entries to archive")
	}

	if len(archiveEntry.Entries) > h.maxFiles {
		return fmt.Errorf("too many files: %d exceeds limit %d", len(archiveEntry.Entries), h.maxFiles)
	}

	// Create buffer for tar data
	var buf bytes.Buffer

	// Create compression writer if needed
	writer, closeFunc, err := h.createWriter(&buf, archiveEntry.Format)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}
	defer func() { _ = closeFunc() }()

	tarWriter := tar.NewWriter(writer)

	// Add files to archive
	var totalSize int64
	for _, entry := range archiveEntry.Entries {
		if entry.IsDirectory() {
			// Create directory entry
			header := &tar.Header{
				Name:     entry.Path + "/",
				Mode:     0755,
				ModTime:  entry.ModifiedAt,
				Typeflag: tar.TypeDir,
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write directory header: %w", err)
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
			header := &tar.Header{
				Name:     entry.Path,
				Mode:     int64(entry.Permissions),
				Size:     entry.Size,
				ModTime:  entry.ModifiedAt,
				Typeflag: tar.TypeReg,
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write file header: %w", err)
			}

			// Write file content
			if _, err := tarWriter.Write(entry.Content); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}

			logger.Log.WithFields(map[string]interface{}{
				"file": entry.Path,
				"size": entry.Size,
			}).Debug("Added file to TAR")
		}
	}

	// Close tar writer
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Close compression writer
	if err := closeFunc(); err != nil {
		return fmt.Errorf("failed to close compression writer: %w", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	logger.Log.WithFields(map[string]interface{}{
		"file_count":  len(archiveEntry.Entries),
		"total_size":  totalSize,
		"output_path": outputPath,
	}).Info("TAR archive created successfully")

	return nil
}

// createReader creates the appropriate reader based on format
func (h *TARHandler) createReader(data []byte, format archive.ArchiveFormat) (io.Reader, error) {
	reader := bytes.NewReader(data)

	switch format {
	case archive.FormatTAR:
		return reader, nil

	case archive.FormatTARGZ:
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gzReader, nil

	case archive.FormatTARBZ2:
		bz2Reader := bzip2.NewReader(reader)
		return bz2Reader, nil

	default:
		return nil, fmt.Errorf("unsupported tar format: %s", format)
	}
}

// createWriter creates the appropriate writer based on format
func (h *TARHandler) createWriter(buf *bytes.Buffer, format archive.ArchiveFormat) (io.Writer, func() error, error) {
	switch format {
	case archive.FormatTAR:
		return buf, func() error { return nil }, nil

	case archive.FormatTARGZ:
		gzWriter := gzip.NewWriter(buf)
		return gzWriter, gzWriter.Close, nil

	case archive.FormatTARBZ2:
		return nil, nil, fmt.Errorf("bzip2 compression not supported for writing (read-only)")

	default:
		return nil, nil, fmt.Errorf("unsupported tar format: %s", format)
	}
}

// validatePath validates a file path to prevent tar-slip attacks
func (h *TARHandler) validatePath(path string) error {
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

// extractFile extracts a single file from the TAR archive
func (h *TARHandler) extractFile(tarReader *tar.Reader, header *tar.Header, outputDir string) (*archive.ArchiveEntry, error) {
	// Validate path (tar-slip protection)
	if err := h.validatePath(header.Name); err != nil {
		return nil, err
	}

	// Read file content with size limit
	var content []byte
	var err error

	switch header.Typeflag {
	case tar.TypeDir:
		// Directory entry
		content = nil

	case tar.TypeReg:
		// Regular file
		content, err = io.ReadAll(io.LimitReader(tarReader, h.maxFileSize))
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}

	case tar.TypeSymlink, tar.TypeLink:
		// Skip symlinks for security
		logger.Log.WithFields(map[string]interface{}{
			"file": header.Name,
			"type": "symlink",
		}).Warn("Skipping symlink in archive")
		return nil, fmt.Errorf("symlinks not allowed in archive: %s", header.Name)

	default:
		// Skip other types
		logger.Log.WithFields(map[string]interface{}{
			"file":     header.Name,
			"typeflag": header.Typeflag,
		}).Debug("Skipping unsupported file type")
		return nil, fmt.Errorf("unsupported file type: %c for %s", header.Typeflag, header.Name)
	}

	// Create archive entry
	entry := &archive.ArchiveEntry{
		Path:        header.Name,
		Size:        header.Size,
		Checksum:    "", // TAR doesn't have built-in checksums like ZIP
		Permissions: uint32(header.Mode),
		ModifiedAt:  header.ModTime,
		ModTime:     header.ModTime,
		MediaType:   h.detectMediaType(header.Name, header.Typeflag == tar.TypeDir),
		Content:     content,
	}

	// Write to file system if output directory specified
	if outputDir != "" {
		outputPath := filepath.Join(outputDir, header.Name)

		// Additional validation after joining paths
		cleanOutput := filepath.Clean(outputPath)
		cleanBase := filepath.Clean(outputDir)
		if !strings.HasPrefix(cleanOutput, cleanBase) {
			return nil, fmt.Errorf("path traversal attempt detected: %s escapes %s", header.Name, outputDir)
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(outputPath, os.FileMode(header.Mode)); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Write file
			if err := os.WriteFile(outputPath, content, os.FileMode(header.Mode)); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}
		}

		extractedAt := time.Now()
		entry.ExtractedAt = &extractedAt
	}

	return entry, nil
}

// detectMediaType determines the media type based on file extension and info
func (h *TARHandler) detectMediaType(path string, isDir bool) archive.MediaType {
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
