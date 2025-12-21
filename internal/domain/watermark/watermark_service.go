// Package watermark provides forensic watermarking services
package watermark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
	domain_ec "github.com/greysquirr3l/shadowforge/internal/domain/errorcorrection"
	infra_crypto "github.com/greysquirr3l/shadowforge/internal/infrastructure/crypto"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/distribution"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/errorcorrection"
	"github.com/greysquirr3l/shadowforge/internal/infrastructure/media"
	stego_impl "github.com/greysquirr3l/shadowforge/internal/infrastructure/stego_impl"
	"github.com/greysquirr3l/shadowforge/pkg/logger"
)

// WatermarkService provides forensic watermarking operations
type WatermarkService struct {
	cryptoService          *infra_crypto.CirclCryptoService
	errorCorrectionService *errorcorrection.RSService
	stegoService           *stego_impl.StegoService
	mediaService           *media.MediaService
	distributionService    *distribution.DistributionService
}

// NewWatermarkService creates a new watermark service
func NewWatermarkService(
	cryptoService *infra_crypto.CirclCryptoService,
	errorCorrectionService *errorcorrection.RSService,
	stegoService *stego_impl.StegoService,
	mediaService *media.MediaService,
	distributionService *distribution.DistributionService,
) *WatermarkService {
	return &WatermarkService{
		cryptoService:          cryptoService,
		errorCorrectionService: errorCorrectionService,
		stegoService:           stegoService,
		mediaService:           mediaService,
		distributionService:    distributionService,
	}
}

// EmbedRequest contains parameters for watermark embedding
type EmbedRequest struct {
	ForensicData  *ForensicData
	Config        *WatermarkConfig
	InputDir      string
	OutputDir     string
	PublicKeyPath string
	ReceiptPath   string
}

// EmbedResult contains results from watermark embedding
type EmbedResult struct {
	WatermarkID      string
	ImagesProcessed  int
	Strategy         WatermarkStrategy
	Manifest         map[string]interface{}
	OutputPaths      []string
	TotalDataSize    int
	DistributedBytes int
}

// Embed embeds forensic watermark data into images
func (s *WatermarkService) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResult, error) {
	// Validate forensic data
	if err := req.ForensicData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid forensic data: %w", err)
	}

	// Generate watermark ID
	req.ForensicData.WatermarkID = GenerateWatermarkID()

	// Serialize forensic data to JSON
	serialized, err := json.Marshal(req.ForensicData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize forensic data: %w", err)
	}

	logger.Log.Infof("Serialized forensic data: %d bytes", len(serialized))

	// Encrypt payload if enabled
	var payload []byte
	if req.Config.Encrypt {
		// Use Kyber-1024 for encryption
		// Load public key (for now, generate ephemeral key pair)
		keyPair, err := s.cryptoService.GenerateKeyPair(ctx, domain_crypto.Kyber1024)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		if req.ReceiptPath != "" {
			if err := writeEncryptionKeyReceiptMarkdown(req.ReceiptPath, req.ForensicData, keyPair, domain_crypto.Kyber1024); err != nil {
				return nil, fmt.Errorf("write encryption receipt: %w", err)
			}
			logger.Log.WithField("receipt_path", req.ReceiptPath).Info("Wrote watermark encryption receipt")
		}

		encryptedPayload, err := s.cryptoService.Encrypt(ctx, serialized, keyPair.PublicKey, domain_crypto.Kyber1024)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}

		// Sign ciphertext (nonce + ciphertext) if enabled.
		// This allows verification prior to decryption, and survives JSON serialization.
		if req.Config.Sign {
			msg := watermarkSignatureMessageV1(encryptedPayload.Nonce, encryptedPayload.Data)
			signature, err := s.cryptoService.Sign(ctx, msg, keyPair.PrivateKey, domain_crypto.Dilithium3)
			if err != nil {
				return nil, fmt.Errorf("sign encrypted payload: %w", err)
			}
			encryptedPayload.Signature = signature
			encryptedPayload.PublicKey = keyPair.PublicKey
		}

		// Serialize the CryptoPayload to bytes
		payload, err = json.Marshal(encryptedPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize encrypted payload: %w", err)
		}
		logger.Log.Infof("Encrypted payload: %d bytes", len(payload))
	} else {
		if req.Config.Sign {
			keyPair, err := s.cryptoService.GenerateKeyPair(ctx, domain_crypto.Kyber1024)
			if err != nil {
				return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
			}
			msg := watermarkSignatureMessageV1(nil, serialized)
			signature, err := s.cryptoService.Sign(ctx, msg, keyPair.PrivateKey, domain_crypto.Dilithium3)
			if err != nil {
				return nil, fmt.Errorf("sign payload: %w", err)
			}
			envelope := signedWatermarkPayloadV1{
				Version:   signedWatermarkPayloadV1Version,
				Data:      serialized,
				Signature: signature,
				PublicKey: keyPair.PublicKey,
			}
			payload, err = json.Marshal(&envelope)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize signed payload: %w", err)
			}
		} else {
			payload = serialized
		}
	}

	// Load images from input directory
	images, err := s.loadImages(req.InputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no PNG images found in %s", req.InputDir)
	}

	logger.Log.Infof("Loaded %d images", len(images))

	// Create Reed-Solomon shards
	dataShards := len(images)
	parityShards := int(float64(dataShards) * req.Config.Redundancy)
	if parityShards < 1 {
		parityShards = 1
	}

	totalShards := dataShards + parityShards
	logger.Log.Infof("Reed-Solomon configuration: %d data + %d parity = %d total shards",
		dataShards, parityShards, totalShards)

	// Encode with Reed-Solomon
	shardConfig := &domain_ec.ShardConfiguration{
		DataShards:   dataShards,
		ParityShards: parityShards,
	}
	protectedMsg, err := s.errorCorrectionService.Encode(ctx, payload, shardConfig)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	shards := protectedMsg.Shards
	logger.Log.Infof("Created %d shards", len(shards))

	// Embed shards in images
	outputPaths := make([]string, 0, len(images))
	for i, img := range images {
		if i >= len(shards) {
			break
		}

		// Convert technique to StegoTechnique (it's already a StegoTechnique)
		stegoTech := req.Config.Technique

		// Embed shard in image using configured technique
		stegoContainer, err := s.stegoService.Embed(ctx, img.Data, shards[i].Data, stegoTech)
		if err != nil {
			return nil, fmt.Errorf("failed to embed in image %d: %w", i, err)
		}

		// Save watermarked image
		outputFilename := filepath.Base(img.Path)
		outputPath := filepath.Join(req.OutputDir, outputFilename)

		// Write the stego media to file (CoverMedia contains the stego output)
		if err := os.WriteFile(outputPath, stegoContainer.CoverMedia, 0644); err != nil {
			return nil, fmt.Errorf("failed to save watermarked image: %w", err)
		}

		outputPaths = append(outputPaths, outputPath)
		logger.Log.Infof("Saved watermarked image: %s", outputPath)
	}

	// Generate manifest
	manifest := map[string]interface{}{
		"watermark_id":     req.ForensicData.WatermarkID,
		"recipient":        req.ForensicData.Recipient,
		"prepared_date":    req.ForensicData.PreparedDate,
		"document_id":      req.ForensicData.DocumentID,
		"strategy":         req.Config.Strategy,
		"technique":        req.Config.Technique,
		"redundancy":       req.Config.Redundancy,
		"total_shards":     totalShards,
		"data_shards":      dataShards,
		"parity_shards":    parityShards,
		"images_processed": len(outputPaths),
		"encrypted":        req.Config.Encrypt,
		"signed":           req.Config.Sign,
	}

	return &EmbedResult{
		WatermarkID:      req.ForensicData.WatermarkID,
		ImagesProcessed:  len(outputPaths),
		Strategy:         req.Config.Strategy,
		Manifest:         manifest,
		OutputPaths:      outputPaths,
		TotalDataSize:    len(payload),
		DistributedBytes: len(shards) * len(shards[0].Data),
	}, nil
}

// ExtractRequest contains parameters for watermark extraction
type ExtractRequest struct {
	InputDir    string
	Config      *WatermarkConfig
	Threshold   int
	ReceiptPath string
}

// ExtractResult contains results from watermark extraction
type ExtractResult struct {
	ForensicData    *ForensicData
	ShardsRecovered int
	ShardsNeeded    int
	Success         bool
	Verified        bool
}

// Extract extracts forensic watermark data from images
func (s *WatermarkService) Extract(ctx context.Context, req *ExtractRequest) (*ExtractResult, error) {
	// Load watermarked images
	images, err := s.loadImages(req.InputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no PNG images found in %s", req.InputDir)
	}

	logger.Log.Infof("Loaded %d watermarked images", len(images))

	// The watermark embed path currently stores only shard data inside each image.
	// Reed-Solomon decoding requires non-zero shard/message IDs for validation, so
	// we synthesize them during extraction.
	messageID := domain_ec.GenerateMessageID()

	// Extract shards from images
	shards := make([]*domain_ec.Shard, 0, len(images))
	for i, img := range images {
		// Technique is already StegoTechnique
		stegoTech := req.Config.Technique

		extractedData, err := s.stegoService.Extract(ctx, img.Data, stegoTech)
		if err != nil {
			logger.Log.Warnf("Failed to extract from image %d: %v", i, err)
			continue // Try next image
		}

		// Create a Shard with the extracted data
		shard := &domain_ec.Shard{
			ID:        domain_ec.GenerateShardID(),
			MessageID: messageID,
			Index:     i,
			Data:      extractedData,
		}
		shards = append(shards, shard)
		logger.Log.Infof("Extracted shard %d: %d bytes", i, len(extractedData))
	}

	logger.Log.Infof("Recovered %d shards", len(shards))

	// Check if we have enough shards
	if len(shards) < req.Threshold {
		return &ExtractResult{
			ShardsRecovered: len(shards),
			ShardsNeeded:    req.Threshold,
			Success:         false,
		}, fmt.Errorf("insufficient shards: have %d, need %d", len(shards), req.Threshold)
	}

	// Decode with Reed-Solomon
	shardConfig := &domain_ec.ShardConfiguration{
		DataShards:   req.Threshold, // Minimum needed
		ParityShards: len(shards) - req.Threshold,
	}
	payload, err := s.errorCorrectionService.Decode(ctx, shards, shardConfig)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon decoding failed: %w", err)
	}

	logger.Log.Infof("Decoded payload: %d bytes", len(payload))

	verified := false

	// Decrypt payload if enabled
	var decrypted []byte
	if req.Config.Encrypt {
		// Decrypt using Kyber-1024
		// First unmarshal the CryptoPayload
		var cryptoPayload domain_crypto.CryptoPayload
		if err := json.Unmarshal(payload, &cryptoPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal encrypted payload: %w", err)
		}

		// Verify signature (over nonce + ciphertext) if enabled.
		if req.Config.Sign {
			verified = s.verifyCiphertextSignature(ctx, cryptoPayload.Nonce, cryptoPayload.Data, cryptoPayload.Signature, cryptoPayload.PublicKey)
		}

		if req.ReceiptPath == "" {
			return nil, fmt.Errorf("encrypted watermark requires --receipt (Markdown encryption receipt) to decrypt")
		}

		receiptKeyPair, err := readEncryptionKeyReceiptMarkdown(req.ReceiptPath)
		if err != nil {
			return nil, fmt.Errorf("read encryption receipt: %w", err)
		}
		decrypted, err = s.cryptoService.Decrypt(ctx, &cryptoPayload, receiptKeyPair.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt payload: %w", err)
		}
		logger.Log.Infof("Decrypted payload: %d bytes", len(decrypted))
	} else {
		if req.Config.Sign {
			var envelope signedWatermarkPayloadV1
			if err := json.Unmarshal(payload, &envelope); err != nil {
				return nil, fmt.Errorf("failed to unmarshal signed payload: %w", err)
			}
			if envelope.Version != signedWatermarkPayloadV1Version {
				return nil, fmt.Errorf("unsupported signed payload version: %s", envelope.Version)
			}
			verified = s.verifyCiphertextSignature(ctx, nil, envelope.Data, envelope.Signature, envelope.PublicKey)
			decrypted = envelope.Data
		} else {
			decrypted = payload
		}
	}

	// Deserialize forensic data
	var forensicData ForensicData
	if err := json.Unmarshal(decrypted, &forensicData); err != nil {
		return nil, fmt.Errorf("failed to deserialize forensic data: %w", err)
	}

	return &ExtractResult{
		ForensicData:    &forensicData,
		ShardsRecovered: len(shards),
		ShardsNeeded:    req.Threshold,
		Success:         true,
		Verified:        verified,
	}, nil
}

const signedWatermarkPayloadV1Version = "shadowforge-watermark-signed:v1"

type signedWatermarkPayloadV1 struct {
	Version   string `json:"version"`
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
	PublicKey []byte `json:"public_key"`
}

func watermarkSignatureMessageV1(nonce, ciphertextOrPlaintext []byte) []byte {
	// Stable message format that survives JSON round-trips.
	// We avoid relying on CryptoPayload fields that don't serialize cleanly.
	var b bytes.Buffer
	b.WriteString("shadowforge-watermark-sig:v1")
	b.WriteByte(0)
	b.Write(nonce)
	b.WriteByte(0)
	b.Write(ciphertextOrPlaintext)
	return b.Bytes()
}

func (s *WatermarkService) verifyCiphertextSignature(ctx context.Context, nonce, data, signature, publicKey []byte) bool {
	if len(signature) == 0 || len(publicKey) == 0 {
		logger.Log.Warn("Signature/public key missing; cannot verify watermark signature")
		return false
	}

	msg := watermarkSignatureMessageV1(nonce, data)
	valid, err := s.cryptoService.Verify(ctx, msg, signature, publicKey, domain_crypto.Dilithium3)
	if err != nil {
		logger.Log.WithField("error", err.Error()).Warn("Signature verification failed")
		return false
	}

	return valid
}

// MediaFile represents a loaded media file
type MediaFile struct {
	Path string
	Data []byte
}

// loadImages loads all PNG images from a directory
func (s *WatermarkService) loadImages(dir string) ([]*MediaFile, error) {
	var images []*MediaFile

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".png" {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		// Read the file data directly
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Log.Warnf("Failed to read %s: %v", path, err)
			continue
		}

		img := &MediaFile{
			Path: path,
			Data: data,
		}

		images = append(images, img)
	}

	return images, nil
}

// GenerateWatermarkID creates a unique watermark identifier
func GenerateWatermarkID() string {
	return uuid.New().String()
}
