package watermark

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

func TestEncryptionReceipt_WriteAndRead_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	receiptPath := filepath.Join(tmpDir, "receipt.md")

	fd := &ForensicData{
		WatermarkID: "wm-123",
		Recipient:   "Test Recipient <test@example.com>",
		DocumentID:  "DOC-1",
	}

	kyberPubSize := 1568
	kyberPrivSize := 3168

	pub := make([]byte, kyberPubSize+16)  // extra bytes simulate concatenated keys
	priv := make([]byte, kyberPrivSize+8) // extra bytes simulate concatenated keys
	for i := range pub {
		pub[i] = byte(i % 251)
	}
	for i := range priv {
		priv[i] = byte((i + 7) % 251)
	}

	keyPair, err := domain_crypto.NewKeyPair(domain_crypto.Kyber1024, pub, priv)
	require.NoError(t, err)

	err = writeEncryptionKeyReceiptMarkdown(receiptPath, fd, keyPair, domain_crypto.Kyber1024)
	require.NoError(t, err)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(receiptPath)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}

	recovered, err := readEncryptionKeyReceiptMarkdown(receiptPath)
	require.NoError(t, err)

	require.Len(t, recovered.PublicKey, kyberPubSize)
	require.Len(t, recovered.PrivateKey, kyberPrivSize)
	require.Equal(t, pub[:kyberPubSize], recovered.PublicKey)
	require.Equal(t, priv[:kyberPrivSize], recovered.PrivateKey)
}

func TestEncryptionReceipt_Read_MissingFields(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "receipt.md")

	// Missing private_key_base64.
	err := os.WriteFile(path, []byte("# x\n```"+receiptFenceLang+"\nalgorithm: kyber1024\npublic_key_base64: Zm9v\n```\n"), 0600)
	require.NoError(t, err)

	_, err = readEncryptionKeyReceiptMarkdown(path)
	require.ErrorIs(t, err, ErrReceiptMissingPrivateKey)
}
