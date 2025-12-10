// Package media domain events.
package media

import "time"

// MediaLoaded is emitted when media is successfully loaded.
type MediaLoaded struct {
AssetID   AssetID
MediaType MediaType
Format    MediaFormat
Size      int64
LoadedAt  time.Time
}

// FormatDetected is emitted when media format is automatically detected.
type FormatDetected struct {
AssetID        AssetID
DetectedFormat MediaFormat
Confidence     float64 // 0.0 to 1.0
DetectedAt     time.Time
}

// CapacityCalculated is emitted when embedding capacity is calculated.
type CapacityCalculated struct {
AssetID        AssetID
TotalCapacity  int64
UsableCapacity int64
QualityImpact  float64
CalculatedAt   time.Time
}

// MetadataSanitized is emitted when metadata is removed from media.
type MetadataSanitized struct {
AssetID       AssetID
FieldsRemoved []string
SanitizedAt   time.Time
}

// MediaValidated is emitted when media validation completes.
type MediaValidated struct {
AssetID          AssetID
IsValid          bool
ValidationErrors []string
ValidatedAt      time.Time
}

// QualityAnalyzed is emitted when quality analysis completes.
type QualityAnalyzed struct {
AssetID      AssetID
QualityScore float64
Metrics      map[string]float64
AnalyzedAt   time.Time
}
