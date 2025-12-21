package watermark

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

const (
	receiptFenceLang = "shadowforge-keypair"
)

var (
	ErrReceiptMissingPrivateKey = errors.New("receipt missing private key")
	ErrReceiptMissingPublicKey  = errors.New("receipt missing public key")
	ErrReceiptMissingAlgorithm  = errors.New("receipt missing algorithm")
)

func writeEncryptionKeyReceiptMarkdown(path string, forensicData *ForensicData, keyPair *domain_crypto.KeyPair, algorithm domain_crypto.PQCAlgorithm) error {
	if path == "" {
		return fmt.Errorf("receipt path is empty")
	}
	if forensicData == nil {
		return fmt.Errorf("forensic data is nil")
	}
	if keyPair == nil {
		return fmt.Errorf("key pair is nil")
	}

	kyberPub, kyberPriv, err := extractKyberKeyMaterial(keyPair)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create receipt directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open receipt for write: %w", err)
	}
	defer func() { _ = f.Close() }()

	createdAt := time.Now().UTC()

	// SECURITY: This file contains sensitive key material.
	content := strings.Builder{}
	content.WriteString("# Shadowforge Watermark Encryption Receipt\n\n")
	content.WriteString("This file contains sensitive key material. Treat it like a private key.\n")
	content.WriteString("If disclosed, encrypted watermarks can be decrypted.\n\n")
	content.WriteString("## Watermark\n")
	content.WriteString(fmt.Sprintf("- Watermark ID: %s\n", forensicData.WatermarkID))
	content.WriteString(fmt.Sprintf("- Recipient: %s\n", forensicData.Recipient))
	if forensicData.DocumentID != "" {
		content.WriteString(fmt.Sprintf("- Document ID: %s\n", forensicData.DocumentID))
	}
	content.WriteString(fmt.Sprintf("- Receipt Created At (UTC): %s\n\n", createdAt.Format(time.RFC3339)))

	content.WriteString("## Encryption Keypair\n")
	content.WriteString("```" + receiptFenceLang + "\n")
	content.WriteString(fmt.Sprintf("algorithm: %s\n", strings.ToLower(algorithm.Name())))
	content.WriteString(fmt.Sprintf("public_key_base64: %s\n", base64.StdEncoding.EncodeToString(kyberPub)))
	content.WriteString(fmt.Sprintf("private_key_base64: %s\n", base64.StdEncoding.EncodeToString(kyberPriv)))
	content.WriteString("```\n")

	if _, err := f.WriteString(content.String()); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}

	// Best-effort: ensure permissions are strict even if umask interfered.
	_ = os.Chmod(path, 0600)

	return nil
}

func readEncryptionKeyReceiptMarkdown(path string) (*domain_crypto.KeyPair, error) {
	if path == "" {
		return nil, fmt.Errorf("receipt path is empty")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open receipt: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	inFence := false
	var algorithm string
	var publicKeyB64 string
	var privateKeyB64 string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "```") {
			if inFence {
				inFence = false
				continue
			}
			lang := strings.TrimPrefix(line, "```")
			lang = strings.TrimSpace(lang)
			if lang == receiptFenceLang {
				inFence = true
			}
			continue
		}

		if !inFence {
			continue
		}

		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)

		switch key {
		case "algorithm":
			algorithm = strings.ToLower(val)
		case "public_key_base64":
			publicKeyB64 = val
		case "private_key_base64":
			privateKeyB64 = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read receipt: %w", err)
	}

	if algorithm == "" {
		return nil, ErrReceiptMissingAlgorithm
	}
	if publicKeyB64 == "" {
		return nil, ErrReceiptMissingPublicKey
	}
	if privateKeyB64 == "" {
		return nil, ErrReceiptMissingPrivateKey
	}

	pub, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode receipt public key: %w", err)
	}
	priv, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode receipt private key: %w", err)
	}

	// Receipt is watermark-only, currently only supports Kyber-1024.
	// We still validate algorithm name to catch accidental receipt misuse.
	switch algorithm {
	case "kyber-1024", "kyber1024":
		kp, err := domain_crypto.NewKeyPair(domain_crypto.Kyber1024, pub, priv)
		if err != nil {
			return nil, fmt.Errorf("create key pair from receipt: %w", err)
		}
		return kp, nil
	default:
		return nil, fmt.Errorf("unsupported receipt algorithm: %s", algorithm)
	}
}

func extractKyberKeyMaterial(keyPair *domain_crypto.KeyPair) (publicKey []byte, privateKey []byte, err error) {
	if keyPair == nil {
		return nil, nil, fmt.Errorf("key pair is nil")
	}

	// CirclCryptoService.GenerateKeyPair concatenates Kyber + Dilithium keys.
	// For watermark encryption receipts we only persist Kyber material.
	kyberPubSize := 1568
	kyberPrivSize := 3168
	if len(keyPair.PublicKey) < kyberPubSize {
		return nil, nil, fmt.Errorf("public key too short for Kyber-1024")
	}
	if len(keyPair.PrivateKey) < kyberPrivSize {
		return nil, nil, fmt.Errorf("private key too short for Kyber-1024")
	}

	pub := make([]byte, kyberPubSize)
	copy(pub, keyPair.PublicKey[:kyberPubSize])

	priv := make([]byte, kyberPrivSize)
	copy(priv, keyPair.PrivateKey[:kyberPrivSize])

	return pub, priv, nil
}
