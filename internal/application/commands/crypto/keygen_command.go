// Package crypto provides commands for cryptographic operations.
package crypto

import (
	"context"
	"fmt"

	"github.com/greysquirr3l/shadowforge/internal/application/commands"
	"github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// GenerateKeyPairCommand is a command to generate a new cryptographic key pair.
type GenerateKeyPairCommand struct {
	request crypto.KeyPairRequest
}

// NewGenerateKeyPairCommand creates a new GenerateKeyPairCommand.
func NewGenerateKeyPairCommand(request crypto.KeyPairRequest) *GenerateKeyPairCommand {
	return &GenerateKeyPairCommand{
		request: request,
	}
}

// CommandName returns the name of this command.
func (c *GenerateKeyPairCommand) CommandName() string {
	return "crypto.GenerateKeyPair"
}

// Validate validates the command inputs.
func (c *GenerateKeyPairCommand) Validate() error {
	if !c.request.Algorithm.IsValid() {
		return fmt.Errorf("invalid algorithm: %s", c.request.Algorithm.Name())
	}

	// KeySize is optional - if provided, validate it matches algorithm
	if c.request.KeySize > 0 {
		expectedSize := c.request.Algorithm.KeySize()
		if c.request.KeySize != expectedSize {
			return fmt.Errorf("invalid key size: expected %d for %s, got %d",
				expectedSize, c.request.Algorithm.Name(), c.request.KeySize)
		}
	}

	return nil
}

// Request returns the key generation request.
func (c *GenerateKeyPairCommand) Request() crypto.KeyPairRequest {
	return c.request
}

// GenerateKeyPairCommandHandler handles GenerateKeyPairCommand.
type GenerateKeyPairCommandHandler struct {
	cryptoService domain_crypto.CryptoService
}

// NewGenerateKeyPairCommandHandler creates a new GenerateKeyPairCommandHandler.
func NewGenerateKeyPairCommandHandler(cryptoService domain_crypto.CryptoService) *GenerateKeyPairCommandHandler {
	return &GenerateKeyPairCommandHandler{
		cryptoService: cryptoService,
	}
}

// Handle executes the GenerateKeyPairCommand.
func (h *GenerateKeyPairCommandHandler) Handle(ctx context.Context, cmd commands.Command) (crypto.KeyPairResponse, error) {
	// Type assert to GenerateKeyPairCommand
	keygenCmd, ok := cmd.(*GenerateKeyPairCommand)
	if !ok {
		return crypto.KeyPairResponse{}, fmt.Errorf("expected *GenerateKeyPairCommand, got %T", cmd)
	}

	req := keygenCmd.Request()

	// Generate key pair using domain service
	keyPair, err := h.cryptoService.GenerateKeyPair(ctx, req.Algorithm)
	if err != nil {
		return crypto.KeyPairResponse{}, fmt.Errorf("key generation failed: %w", err)
	}

	// Build response
	response := crypto.KeyPairResponse{
		PublicKey:   keyPair.PublicKey,
		PrivateKey:  keyPair.PrivateKey,
		Algorithm:   keyPair.Algorithm,
		GeneratedAt: keyPair.CreatedAt,
	}

	return response, nil
}
