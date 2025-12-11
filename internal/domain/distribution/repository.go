// Package distribution provides distribution domain repository interfaces.
package distribution

import "context"

// Repository defines persistence operations for distribution domain.
type Repository interface {
	// SaveStrategy persists a distribution strategy.
	SaveStrategy(ctx context.Context, strategy *DistributionStrategy) error

	// GetStrategy retrieves a distribution strategy by ID.
	GetStrategy(ctx context.Context, id StrategyID) (*DistributionStrategy, error)

	// ListStrategies retrieves a paginated list of distribution strategies.
	ListStrategies(ctx context.Context, limit, offset int) ([]*DistributionStrategy, error)

	// DeleteStrategy removes a distribution strategy.
	DeleteStrategy(ctx context.Context, id StrategyID) error

	// SaveManifest persists a shard manifest.
	SaveManifest(ctx context.Context, manifest *ShardManifest) error

	// GetManifest retrieves a shard manifest by ID.
	GetManifest(ctx context.Context, id ManifestID) (*ShardManifest, error)

	// GetManifestsByStrategy retrieves all manifests for a strategy.
	GetManifestsByStrategy(ctx context.Context, strategyID StrategyID) ([]*ShardManifest, error)
}
