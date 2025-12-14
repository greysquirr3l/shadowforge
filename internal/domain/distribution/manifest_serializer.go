// Package distribution provides distribution strategy and shard coordination logic.
package distribution

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ManifestSerializer handles JSON serialization and HMAC integrity protection for manifests.
type ManifestSerializer struct {
	hmacKey []byte
}

// manifestJSON represents the JSON-serializable format of ShardManifest.
// This structure includes an HMAC signature field for integrity verification.
type manifestJSON struct {
	Version        string              `json:"version"`
	ID             string              `json:"id"`
	StrategyID     string              `json:"strategy_id"`
	ShardMetadata  []shardMetadataJSON `json:"shards"`
	TotalShards    int                 `json:"total_shards"`
	RequiredShards int                 `json:"required_shards"`
	CreatedAt      string              `json:"created_at"`
	ExpiresAt      *string             `json:"expires_at,omitempty"`
	Signature      string              `json:"signature"`
}

// shardMetadataJSON represents the JSON-serializable format of ShardMetadata.
type shardMetadataJSON struct {
	Index       int    `json:"index"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
	MediaID     string `json:"media_id"`
	RecipientID string `json:"recipient_id,omitempty"`
}

const (
	// ManifestVersion is the current manifest schema version.
	ManifestVersion = "1.0"
)

// NewManifestSerializer creates a new manifest serializer with the provided HMAC key.
// The HMAC key is used to sign and verify manifest integrity.
func NewManifestSerializer(hmacKey []byte) (*ManifestSerializer, error) {
	if len(hmacKey) == 0 {
		return nil, fmt.Errorf("HMAC key cannot be empty")
	}
	if len(hmacKey) < 32 {
		return nil, fmt.Errorf("HMAC key must be at least 32 bytes")
	}

	return &ManifestSerializer{
		hmacKey: hmacKey,
	}, nil
}

// Marshal serializes a ShardManifest to JSON with HMAC signature.
// Returns the JSON-encoded manifest with signature, or an error if serialization fails.
func (s *ManifestSerializer) Marshal(manifest *ShardManifest) ([]byte, error) {
	if manifest == nil {
		return nil, fmt.Errorf("manifest cannot be nil")
	}

	// Validate manifest before serialization
	if err := s.validateManifest(manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Convert domain model to JSON structure
	jsonManifest := s.toJSON(manifest)

	// Serialize without signature first
	jsonManifest.Signature = ""
	dataWithoutSig, err := json.Marshal(jsonManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Compute HMAC signature over the data (without signature field)
	signature := s.computeHMAC(dataWithoutSig)
	jsonManifest.Signature = signature

	// Serialize final manifest with signature
	finalData, err := json.Marshal(jsonManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest with signature: %w", err)
	}

	return finalData, nil
}

// Unmarshal deserializes JSON data into a ShardManifest and verifies HMAC signature.
// Returns the deserialized manifest or an error if verification fails.
func (s *ManifestSerializer) Unmarshal(data []byte) (*ShardManifest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var jsonManifest manifestJSON
	if err := json.Unmarshal(data, &jsonManifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Verify schema version compatibility
	if jsonManifest.Version != ManifestVersion {
		return nil, fmt.Errorf("unsupported manifest version: %s (expected %s)",
			jsonManifest.Version, ManifestVersion)
	}

	// Verify HMAC signature
	providedSignature := jsonManifest.Signature
	jsonManifest.Signature = ""

	dataWithoutSig, err := json.Marshal(jsonManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal for verification: %w", err)
	}

	expectedSignature := s.computeHMAC(dataWithoutSig)
	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("HMAC signature verification failed: manifest may have been tampered")
	}

	// Convert JSON structure back to domain model
	manifest, err := s.fromJSON(&jsonManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to domain model: %w", err)
	}

	// Validate reconstructed manifest
	if err := s.validateManifest(manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest structure: %w", err)
	}

	return manifest, nil
}

// Sign generates an HMAC signature for the manifest.
// This is useful for separate signature generation/verification workflows.
func (s *ManifestSerializer) Sign(manifest *ShardManifest) (string, error) {
	if manifest == nil {
		return "", fmt.Errorf("manifest cannot be nil")
	}

	// Convert to JSON without signature
	jsonManifest := s.toJSON(manifest)
	jsonManifest.Signature = ""

	data, err := json.Marshal(jsonManifest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return s.computeHMAC(data), nil
}

// Verify checks if the provided signature matches the manifest's HMAC.
func (s *ManifestSerializer) Verify(manifest *ShardManifest, signature string) bool {
	if manifest == nil || signature == "" {
		return false
	}

	expectedSignature, err := s.Sign(manifest)
	if err != nil {
		return false
	}

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// computeHMAC computes the HMAC-SHA256 signature of the data.
func (s *ManifestSerializer) computeHMAC(data []byte) string {
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// toJSON converts a ShardManifest to JSON structure.
func (s *ManifestSerializer) toJSON(manifest *ShardManifest) *manifestJSON {
	shards := make([]shardMetadataJSON, len(manifest.ShardMetadata))
	for i, shard := range manifest.ShardMetadata {
		shards[i] = shardMetadataJSON{
			Index:       shard.Index,
			Size:        shard.Size,
			Checksum:    shard.Checksum,
			MediaID:     shard.MediaID,
			RecipientID: shard.RecipientID,
		}
	}

	var expiresAt *string
	if manifest.ExpiresAt != nil {
		exp := manifest.ExpiresAt.Format(time.RFC3339)
		expiresAt = &exp
	}

	return &manifestJSON{
		Version:        ManifestVersion,
		ID:             manifest.ID.String(),
		StrategyID:     manifest.StrategyID.String(),
		ShardMetadata:  shards,
		TotalShards:    manifest.TotalShards,
		RequiredShards: manifest.RequiredShards,
		CreatedAt:      manifest.CreatedAt.Format(time.RFC3339),
		ExpiresAt:      expiresAt,
	}
}

// fromJSON converts JSON structure back to ShardManifest.
func (s *ManifestSerializer) fromJSON(jsonManifest *manifestJSON) (*ShardManifest, error) {
	// Parse ID
	id, err := ParseManifestID(jsonManifest.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest ID: %w", err)
	}

	// Parse strategy ID
	strategyID, err := ParseStrategyID(jsonManifest.StrategyID)
	if err != nil {
		return nil, fmt.Errorf("invalid strategy ID: %w", err)
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, jsonManifest.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	var expiresAt *time.Time
	if jsonManifest.ExpiresAt != nil {
		expTime, err := time.Parse(time.RFC3339, *jsonManifest.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at timestamp: %w", err)
		}
		expiresAt = &expTime
	}

	// Convert shard metadata
	shards := make([]ShardMetadata, len(jsonManifest.ShardMetadata))
	for i, jsonShard := range jsonManifest.ShardMetadata {
		shard, err := NewShardMetadata(
			jsonShard.Index,
			jsonShard.Size,
			jsonShard.Checksum,
			jsonShard.MediaID,
			jsonShard.RecipientID,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid shard metadata at index %d: %w", i, err)
		}
		shards[i] = shard
	}

	return &ShardManifest{
		ID:             id,
		StrategyID:     strategyID,
		ShardMetadata:  shards,
		TotalShards:    jsonManifest.TotalShards,
		RequiredShards: jsonManifest.RequiredShards,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
	}, nil
}

// validateManifest checks if a manifest is valid for serialization.
func (s *ManifestSerializer) validateManifest(manifest *ShardManifest) error {
	if manifest.ID.IsZero() {
		return fmt.Errorf("manifest ID cannot be zero")
	}
	if manifest.StrategyID.IsZero() {
		return fmt.Errorf("strategy ID cannot be zero")
	}
	if len(manifest.ShardMetadata) == 0 {
		return ErrNoShardMetadata
	}
	if manifest.TotalShards != len(manifest.ShardMetadata) {
		return fmt.Errorf("total shards (%d) does not match metadata count (%d)",
			manifest.TotalShards, len(manifest.ShardMetadata))
	}
	if manifest.RequiredShards < 1 || manifest.RequiredShards > manifest.TotalShards {
		return ErrInvalidThreshold
	}
	if manifest.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp cannot be zero")
	}

	// Validate each shard metadata
	for i, shard := range manifest.ShardMetadata {
		if err := shard.Validate(); err != nil {
			return fmt.Errorf("shard metadata %d is invalid: %w", i, err)
		}
	}

	return nil
}
