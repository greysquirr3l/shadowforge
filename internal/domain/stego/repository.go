// Package stego implements the Steganography bounded context.
package stego

import "context"

// StegoRepository defines the persistence interface for steganographic operations.
// Implementations are in the infrastructure layer.
type StegoRepository interface {
	// SaveContainer persists a steganographic container.
	SaveContainer(ctx context.Context, container *StegoContainer) error

	// GetContainer retrieves a container by its ID.
	GetContainer(ctx context.Context, id ContainerID) (*StegoContainer, error)

	// DeleteContainer removes a container from storage.
	DeleteContainer(ctx context.Context, id ContainerID) error

	// SaveMetadata persists embedding metadata.
	SaveMetadata(ctx context.Context, metadata *EmbeddingMetadata) error

	// GetMetadata retrieves metadata for a container.
	GetMetadata(ctx context.Context, containerID ContainerID) (*EmbeddingMetadata, error)

	// ListContainers lists all containers with optional filtering.
	ListContainers(ctx context.Context, technique StegoTechnique, limit int) ([]*StegoContainer, error)
}
