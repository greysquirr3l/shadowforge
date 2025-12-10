// Package stego implements the Steganography bounded context.
package stego

import "time"

// DomainEvent is the interface that all steganography domain events must implement.
type DomainEvent interface {
	OccurredAt() time.Time
	EventType() string
}

// DataEmbedded is emitted when data has been successfully embedded.
type DataEmbedded struct {
	ContainerID  ContainerID
	Technique    StegoTechnique
	PayloadSize  int64
	CapacityUsed int64
	QualityScore float64
	Timestamp    time.Time
}

func (e DataEmbedded) OccurredAt() time.Time {
	return e.Timestamp
}

func (e DataEmbedded) EventType() string {
	return "stego.data_embedded"
}

// DataExtracted is emitted when data has been successfully extracted.
type DataExtracted struct {
	ContainerID ContainerID
	Technique   StegoTechnique
	PayloadSize int64
	Timestamp   time.Time
}

func (e DataExtracted) OccurredAt() time.Time {
	return e.Timestamp
}

func (e DataExtracted) EventType() string {
	return "stego.data_extracted"
}

// CapacityCalculated is emitted when capacity analysis is complete.
type CapacityCalculated struct {
	ContainerID   ContainerID
	Technique     StegoTechnique
	TotalCapacity int64
	Timestamp     time.Time
}

func (e CapacityCalculated) OccurredAt() time.Time {
	return e.Timestamp
}

func (e CapacityCalculated) EventType() string {
	return "stego.capacity_calculated"
}

// QualityAnalyzed is emitted when quality analysis is complete.
type QualityAnalyzed struct {
	ContainerID        ContainerID
	QualityScore       float64
	DetectabilityScore float64
	FidelityScore      float64
	Timestamp          time.Time
}

func (e QualityAnalyzed) OccurredAt() time.Time {
	return e.Timestamp
}

func (e QualityAnalyzed) EventType() string {
	return "stego.quality_analyzed"
}

// ContainerCorrupted is emitted when container corruption is detected.
type ContainerCorrupted struct {
	ContainerID    ContainerID
	Technique      StegoTechnique
	CorruptionType string
	Timestamp      time.Time
}

func (e ContainerCorrupted) OccurredAt() time.Time {
	return e.Timestamp
}

func (e ContainerCorrupted) EventType() string {
	return "stego.container_corrupted"
}

// TechniqueOptimized is emitted when technique optimization is complete.
type TechniqueOptimized struct {
	OriginalTechnique  StegoTechnique
	OptimizedTechnique StegoTechnique
	ReasonCode         string
	Timestamp          time.Time
}

func (e TechniqueOptimized) OccurredAt() time.Time {
	return e.Timestamp
}

func (e TechniqueOptimized) EventType() string {
	return "stego.technique_optimized"
}
