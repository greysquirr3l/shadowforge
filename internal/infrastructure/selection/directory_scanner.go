package selection

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/greysquirr3l/shadowforge/internal/domain/media"
	"github.com/greysquirr3l/shadowforge/internal/domain/selection"
	"github.com/greysquirr3l/shadowforge/internal/domain/stego"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
)

// DirectoryScanner scans directories for media files and analyzes them.
// Uses local capacity estimation rather than external services.
type DirectoryScanner struct{}

// NewDirectoryScanner creates a new directory scanner.
func NewDirectoryScanner(mediaService, stegoService interface{}) *DirectoryScanner {
	// Type assertions are removed - DirectoryScanner now uses local capacity estimation
	return &DirectoryScanner{}
}

// ScanOptions defines options for directory scanning.
type ScanOptions struct {
	Recursive     bool
	IncludeHidden bool
	MaxDepth      int
	FilePatterns  []string // e.g., []string{"*.png", "*.jpg"}
}

// ScanDirectory scans a directory and returns analysis for all media files.
func (s *DirectoryScanner) ScanDirectory(
	ctx context.Context,
	dirPath string,
	opts ScanOptions,
) (*selection.AnalysisResult, error) {
	logger.Log.WithFields(logrus.Fields{
		"directory": dirPath,
		"recursive": opts.Recursive,
	}).Info("Starting directory scan")

	files := make([]*selection.MediaAnalysis, 0)

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Log.WithError(err).Warn("Error accessing path")
			return nil // Skip errors, continue scanning
		}

		// Check depth limit
		if opts.MaxDepth > 0 {
			relPath, _ := filepath.Rel(dirPath, path)
			depth := len(filepath.SplitList(relPath))
			if depth > opts.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip directories (unless we need to traverse)
		if d.IsDir() {
			if !opts.Recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files unless requested
		if !opts.IncludeHidden && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return nil
		}

		// Check if it's a media file
		mediaType := selection.DetermineMediaType(path)
		if mediaType == "" {
			return nil
		}

		// Analyze the file
		analysis, err := s.analyzeFile(ctx, path, mediaType)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"file":  path,
				"error": err,
			}).Warn("Failed to analyze file")
			return nil // Skip this file, continue
		}

		files = append(files, analysis)
		return nil
	}

	if err := filepath.WalkDir(dirPath, walkFunc); err != nil {
		return nil, fmt.Errorf("directory scan failed: %w", err)
	}

	// Compile results
	result := &selection.AnalysisResult{
		Files:             files,
		TotalRawCapacity:  0,
		TotalSafeCapacity: 0,
		MediaTypeCounts:   make(map[media.MediaType]int),
	}

	totalDetectScore := 0.0
	for _, file := range files {
		result.TotalRawCapacity += file.RawCapacity
		result.TotalSafeCapacity += file.SafeCapacity
		totalDetectScore += file.DetectabilityScore
		result.MediaTypeCounts[file.Type]++
	}

	if len(files) > 0 {
		result.AverageDetectScore = totalDetectScore / float64(len(files))
	}

	logger.Log.WithFields(logrus.Fields{
		"files_found":    len(files),
		"total_capacity": result.TotalSafeCapacity,
		"image_count":    result.MediaTypeCounts[media.MediaTypeImage],
		"audio_count":    result.MediaTypeCounts[media.MediaTypeAudio],
		"text_count":     result.MediaTypeCounts[media.MediaTypeText],
	}).Info("Directory scan completed")

	return result, nil
}

// analyzeFile performs detailed analysis on a single media file.
func (s *DirectoryScanner) analyzeFile(
	ctx context.Context,
	filePath string,
	mediaType media.MediaType,
) (*selection.MediaAnalysis, error) {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}

	format := selection.DetermineMediaFormat(filePath)
	technique := selection.RecommendTechniqueForMedia(mediaType, format)

	// Calculate capacity (simplified - would use actual stego service)
	rawCap, safeCap := s.estimateCapacity(info.Size(), mediaType, technique)

	analysis := &selection.MediaAnalysis{
		Path:                 filePath,
		Type:                 mediaType,
		Format:               format,
		FileSize:             info.Size(),
		RawCapacity:          rawCap,
		SafeCapacity:         safeCap,
		DetectabilityScore:   selection.CalculateDetectabilityScore(safeCap, safeCap/2, technique),
		RecommendedTechnique: technique,
		Dimensions:           s.getDimensions(filePath, mediaType),
	}

	return analysis, nil
}

// estimateCapacity provides a rough capacity estimate based on file size.
func (s *DirectoryScanner) estimateCapacity(fileSize int64, mediaType media.MediaType, technique stego.StegoTechnique) (raw, safe int64) {
	switch mediaType {
	case media.MediaTypeImage:
		// Images: ~10-30% of file size depending on technique
		raw = int64(float64(fileSize) * 0.2)
		safe = int64(float64(raw) * 0.7)
	case media.MediaTypeAudio:
		// Audio: ~1-5% of file size
		raw = int64(float64(fileSize) * 0.03)
		safe = int64(float64(raw) * 0.7)
	case media.MediaTypeText:
		// Text: Very low capacity
		raw = fileSize / 100
		safe = raw / 2
	}
	return
}

// getDimensions gets dimension information for the media file.
func (s *DirectoryScanner) getDimensions(filePath string, mediaType media.MediaType) string {
	// This would use actual image/audio libraries in full implementation
	// For now, return placeholder
	switch mediaType {
	case media.MediaTypeImage:
		return "Unknown dimensions" // Would read actual dimensions
	case media.MediaTypeAudio:
		return "Unknown duration" // Would read actual duration
	default:
		return "N/A"
	}
}
