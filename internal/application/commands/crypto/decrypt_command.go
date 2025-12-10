// Package crypto provides commands for cryptographic operations.
package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// DecryptDataCommand is a command to decrypt data using post-quantum cryptography.
type DecryptDataCommand struct {
	request crypto.DecryptionRequest
}

// NewDecryptDataCommand creates a new DecryptDataCommand.
func NewDecryptDataCommand(request crypto.DecryptionRequest) *DecryptDataCommand {
	return &DecryptDataCommand{
		request: request,
	}
}

// CommandName returns the name of this command.
func (c *DecryptDataCommand) CommandName() string {
	return "crypto.DecryptData"
}

// Validate validates the command inputs.
func (c *DecryptDataCommand) Validate() error {
	if len(c.request.EncryptedData) == 0 {
		return fmt.Errorf("encrypted data cannot be empty")
	}

	if !c.request.Algorithm.IsValid() {
		return fmt.Errorf("invalid algorithm: %s", c.request.Algorithm.Name())
	}

	if len(c.request.RecipientPrivateKey) == 0 {
		return fmt.Errorf("recipient private key cannot be empty")
	}

	if len(c.request.SenderPublicKey) == 0 {
		return fmt.Errorf("sender public key cannot be empty")
	}

	if len(c.request.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	return nil
}

// Request returns the decryption request.
func (c *DecryptDataCommand) Request() crypto.DecryptionRequest {
	return c.request
}

// DecryptDataCommandHandler handles DecryptDataCommand.
type DecryptDataCommandHandler struct {
	cryptoService domain_crypto.CryptoService
}

// NewDecryptDataCommandHandler creates a new DecryptDataCommandHandler.
func NewDecryptDataCommandHandler(cryptoService domain_crypto.CryptoService) *DecryptDataCommandHandler {
	return &DecryptDataCommandHandler{
		cryptoService: cryptoService,
	}
}

// Handle executes the DecryptDataCommand.
func (h *DecryptDataCommandHandler) Handle(ctx context.Context, cmd commands.Command) (crypto.DecryptionResponse, error) {
	// Type assert to DecryptDataCommand
	decryptCmd, ok := cmd.(*DecryptDataCommand)
	if !ok {
		return crypto.DecryptionResponse{}, fmt.Errorf("expected *DecryptDataCommand, got %T", cmd)
	}

	req := decryptCmd.Request()

	// Create CryptoPayload from request
	payloadID := domain_crypto.GeneratePayloadID()
	payload := &domain_crypto.CryptoPayload{
		ID:        payloadID,
		Data:      req.EncryptedData,
		Algorithm: req.Algorithm,
		Nonce:     req.Nonce,
		Signature: req.Signature,
		PublicKey: req.SenderPublicKey,
	}

	// Verify signature first
	signatureValid, err := h.cryptoService.Verify(ctx, req.EncryptedData, req.Signature, req.SenderPublicKey, req.Algorithm)
	if err != nil {
		return crypto.DecryptionResponse{}, fmt.Errorf("signature verification failed: %w", err)
	}

	// Decrypt the data
	decryptedData, err := h.cryptoService.Decrypt(ctx, payload, req.RecipientPrivateKey)
	if err != nil {
		return crypto.DecryptionResponse{}, fmt.Errorf("decryption failed: %w", err)
	}

	// Build response
	response := crypto.DecryptionResponse{
		Data:           decryptedData,
		SignatureValid: signatureValid,
		DecryptedAt:    time.Now(),
	}

	return response, nil
}
