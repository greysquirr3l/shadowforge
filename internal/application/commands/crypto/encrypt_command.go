// Package crypto provides commands for cryptographic operations.
package crypto

import (
	"context"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// EncryptDataCommand is a command to encrypt data using post-quantum cryptography.
type EncryptDataCommand struct {
	request crypto.EncryptionRequest
}

// NewEncryptDataCommand creates a new EncryptDataCommand.
func NewEncryptDataCommand(request crypto.EncryptionRequest) *EncryptDataCommand {
	return &EncryptDataCommand{
		request: request,
	}
}

// CommandName returns the name of this command.
func (c *EncryptDataCommand) CommandName() string {
	return "crypto.EncryptData"
}

// Validate validates the command inputs.
func (c *EncryptDataCommand) Validate() error {
	if len(c.request.Data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	if !c.request.Algorithm.IsValid() {
		return fmt.Errorf("invalid algorithm: %s", c.request.Algorithm.Name())
	}

	if len(c.request.RecipientPublicKey) == 0 {
		return fmt.Errorf("recipient public key cannot be empty")
	}

	if len(c.request.SenderPrivateKey) == 0 {
		return fmt.Errorf("sender private key cannot be empty")
	}

	return nil
}

// Request returns the encryption request.
func (c *EncryptDataCommand) Request() crypto.EncryptionRequest {
	return c.request
}

// EncryptDataCommandHandler handles EncryptDataCommand.
type EncryptDataCommandHandler struct {
	cryptoService domain_crypto.CryptoService
}

// NewEncryptDataCommandHandler creates a new EncryptDataCommandHandler.
func NewEncryptDataCommandHandler(cryptoService domain_crypto.CryptoService) *EncryptDataCommandHandler {
	return &EncryptDataCommandHandler{
		cryptoService: cryptoService,
	}
}

// Handle executes the EncryptDataCommand.
func (h *EncryptDataCommandHandler) Handle(ctx context.Context, cmd commands.Command) (crypto.EncryptionResponse, error) {
	// Type assert to EncryptDataCommand
	encryptCmd, ok := cmd.(*EncryptDataCommand)
	if !ok {
		return crypto.EncryptionResponse{}, fmt.Errorf("expected *EncryptDataCommand, got %T", cmd)
	}

	req := encryptCmd.Request()

	// Call domain service to encrypt data
	payload, err := h.cryptoService.Encrypt(ctx, req.Data, req.RecipientPublicKey, req.Algorithm)
	if err != nil {
		return crypto.EncryptionResponse{}, fmt.Errorf("encryption failed: %w", err)
	}

	// Sign the encrypted data using the sender's private key
	signature, err := h.cryptoService.Sign(ctx, payload.Data, req.SenderPrivateKey, req.Algorithm)
	if err != nil {
		return crypto.EncryptionResponse{}, fmt.Errorf("signing failed: %w", err)
	}

	// Build response
	response := crypto.EncryptionResponse{
		EncryptedData: payload.Data,
		Signature:     signature,
		Nonce:         payload.Nonce,
		Algorithm:     req.Algorithm,
		EncryptedAt:   payload.EncryptedAt,
	}

	return response, nil
}
