package crypto_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/greysquirr3l/shadowforge/internal/application/commands/crypto"
	dto_crypto "github.com/greysquirr3l/shadowforge/internal/application/dto/crypto"
	domain_crypto "github.com/greysquirr3l/shadowforge/internal/domain/crypto"
)

// MockCryptoService is a mock implementation of CryptoService
type MockCryptoService struct {
	mock.Mock
}

func (m *MockCryptoService) GenerateKeyPair(ctx context.Context, algorithm domain_crypto.PQCAlgorithm) (*domain_crypto.KeyPair, error) {
	args := m.Called(ctx, algorithm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain_crypto.KeyPair), args.Error(1)
}

func (m *MockCryptoService) Encrypt(ctx context.Context, data []byte, publicKey []byte, algorithm domain_crypto.PQCAlgorithm) (*domain_crypto.CryptoPayload, error) {
	args := m.Called(ctx, data, publicKey, algorithm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain_crypto.CryptoPayload), args.Error(1)
}

func (m *MockCryptoService) Decrypt(ctx context.Context, payload *domain_crypto.CryptoPayload, privateKey []byte) ([]byte, error) {
	args := m.Called(ctx, payload, privateKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) Sign(ctx context.Context, data []byte, privateKey []byte, algorithm domain_crypto.PQCAlgorithm) ([]byte, error) {
	args := m.Called(ctx, data, privateKey, algorithm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) Verify(ctx context.Context, data []byte, signature []byte, publicKey []byte, algorithm domain_crypto.PQCAlgorithm) (bool, error) {
	args := m.Called(ctx, data, signature, publicKey, algorithm)
	return args.Bool(0), args.Error(1)
}

func (m *MockCryptoService) DeriveKey(ctx context.Context, password string, salt []byte) ([]byte, error) {
	args := m.Called(ctx, password, salt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test EncryptDataCommand

func TestEncryptDataCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request dto_crypto.EncryptionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request",
			request: dto_crypto.EncryptionRequest{
				Data:               []byte("test data"),
				Algorithm:          domain_crypto.Kyber1024,
				RecipientPublicKey: []byte("public-key"),
				SenderPrivateKey:   []byte("private-key"),
			},
			wantErr: false,
		},
		{
			name: "empty_data",
			request: dto_crypto.EncryptionRequest{
				Data:               []byte{},
				Algorithm:          domain_crypto.Kyber1024,
				RecipientPublicKey: []byte("public-key"),
				SenderPrivateKey:   []byte("private-key"),
			},
			wantErr: true,
			errMsg:  "data cannot be empty",
		},
		{
			name: "invalid_algorithm",
			request: dto_crypto.EncryptionRequest{
				Data:               []byte("test data"),
				Algorithm:          domain_crypto.UnknownAlgo,
				RecipientPublicKey: []byte("public-key"),
				SenderPrivateKey:   []byte("private-key"),
			},
			wantErr: true,
			errMsg:  "invalid algorithm",
		},
		{
			name: "empty_recipient_public_key",
			request: dto_crypto.EncryptionRequest{
				Data:               []byte("test data"),
				Algorithm:          domain_crypto.Kyber1024,
				RecipientPublicKey: []byte{},
				SenderPrivateKey:   []byte("private-key"),
			},
			wantErr: true,
			errMsg:  "recipient public key cannot be empty",
		},
		{
			name: "empty_sender_private_key",
			request: dto_crypto.EncryptionRequest{
				Data:               []byte("test data"),
				Algorithm:          domain_crypto.Kyber1024,
				RecipientPublicKey: []byte("public-key"),
				SenderPrivateKey:   []byte{},
			},
			wantErr: true,
			errMsg:  "sender private key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := crypto.NewEncryptDataCommand(tt.request)
			err := cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEncryptDataCommandHandler_Handle(t *testing.T) {
	ctx := context.Background()
	testData := []byte("test secret data")
	testAlgorithm := domain_crypto.Kyber1024

	t.Run("successful_encryption", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewEncryptDataCommandHandler(mockService)

		request := dto_crypto.EncryptionRequest{
			Data:               testData,
			Algorithm:          testAlgorithm,
			RecipientPublicKey: []byte("recipient-public"),
			SenderPrivateKey:   []byte("sender-private"),
		}

		// Mock payload
		payloadID := domain_crypto.GeneratePayloadID()
		mockPayload := &domain_crypto.CryptoPayload{
			ID:          payloadID,
			Data:        []byte("encrypted-data"),
			Algorithm:   testAlgorithm,
			Nonce:       []byte("test-nonce"),
			EncryptedAt: time.Now(),
		}

		mockService.On("Encrypt", ctx, testData, request.RecipientPublicKey, testAlgorithm).
			Return(mockPayload, nil)
		mockService.On("Sign", ctx, mockPayload.Data, request.SenderPrivateKey, testAlgorithm).
			Return([]byte("signature"), nil)

		cmd := crypto.NewEncryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		require.NoError(t, err)
		assert.Equal(t, mockPayload.Data, response.EncryptedData)
		assert.Equal(t, []byte("signature"), response.Signature)
		assert.Equal(t, mockPayload.Nonce, response.Nonce)
		assert.Equal(t, testAlgorithm, response.Algorithm)

		mockService.AssertExpectations(t)
	})

	t.Run("encryption_failure", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewEncryptDataCommandHandler(mockService)

		request := dto_crypto.EncryptionRequest{
			Data:               testData,
			Algorithm:          testAlgorithm,
			RecipientPublicKey: []byte("recipient-public"),
			SenderPrivateKey:   []byte("sender-private"),
		}

		expectedErr := errors.New("encryption service error")
		mockService.On("Encrypt", ctx, testData, request.RecipientPublicKey, testAlgorithm).
			Return(nil, expectedErr)

		cmd := crypto.NewEncryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption failed")
		assert.Empty(t, response.EncryptedData)

		mockService.AssertExpectations(t)
	})

	t.Run("signing_failure", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewEncryptDataCommandHandler(mockService)

		request := dto_crypto.EncryptionRequest{
			Data:               testData,
			Algorithm:          testAlgorithm,
			RecipientPublicKey: []byte("recipient-public"),
			SenderPrivateKey:   []byte("sender-private"),
		}

		payloadID := domain_crypto.GeneratePayloadID()
		mockPayload := &domain_crypto.CryptoPayload{
			ID:          payloadID,
			Data:        []byte("encrypted-data"),
			Algorithm:   testAlgorithm,
			Nonce:       []byte("test-nonce"),
			EncryptedAt: time.Now(),
		}

		expectedErr := errors.New("signing service error")
		mockService.On("Encrypt", ctx, testData, request.RecipientPublicKey, testAlgorithm).
			Return(mockPayload, nil)
		mockService.On("Sign", ctx, mockPayload.Data, request.SenderPrivateKey, testAlgorithm).
			Return(nil, expectedErr)

		cmd := crypto.NewEncryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signing failed")
		assert.Empty(t, response.EncryptedData)

		mockService.AssertExpectations(t)
	})
}

// Test DecryptDataCommand

func TestDecryptDataCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request dto_crypto.DecryptionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request",
			request: dto_crypto.DecryptionRequest{
				EncryptedData:       []byte("encrypted-data"),
				Algorithm:           domain_crypto.Kyber1024,
				RecipientPrivateKey: []byte("private-key"),
				SenderPublicKey:     []byte("public-key"),
				Signature:           []byte("signature"),
				Nonce:               []byte("nonce"),
			},
			wantErr: false,
		},
		{
			name: "empty_encrypted_data",
			request: dto_crypto.DecryptionRequest{
				EncryptedData:       []byte{},
				Algorithm:           domain_crypto.Kyber1024,
				RecipientPrivateKey: []byte("private-key"),
				SenderPublicKey:     []byte("public-key"),
				Signature:           []byte("signature"),
			},
			wantErr: true,
			errMsg:  "encrypted data cannot be empty",
		},
		{
			name: "empty_signature",
			request: dto_crypto.DecryptionRequest{
				EncryptedData:       []byte("encrypted-data"),
				Algorithm:           domain_crypto.Kyber1024,
				RecipientPrivateKey: []byte("private-key"),
				SenderPublicKey:     []byte("public-key"),
				Signature:           []byte{},
			},
			wantErr: true,
			errMsg:  "signature cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := crypto.NewDecryptDataCommand(tt.request)
			err := cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecryptDataCommandHandler_Handle(t *testing.T) {
	ctx := context.Background()
	testData := []byte("test secret data")
	testAlgorithm := domain_crypto.Kyber1024

	t.Run("successful_decryption_valid_signature", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewDecryptDataCommandHandler(mockService)

		request := dto_crypto.DecryptionRequest{
			EncryptedData:       []byte("encrypted-data"),
			Algorithm:           testAlgorithm,
			RecipientPrivateKey: []byte("recipient-private"),
			SenderPublicKey:     []byte("sender-public"),
			Signature:           []byte("signature"),
			Nonce:               []byte("nonce"),
		}

		mockService.On("Verify", ctx, request.EncryptedData, request.Signature, request.SenderPublicKey, testAlgorithm).
			Return(true, nil)
		mockService.On("Decrypt", ctx, mock.AnythingOfType("*crypto.CryptoPayload"), request.RecipientPrivateKey).
			Return(testData, nil)

		cmd := crypto.NewDecryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		require.NoError(t, err)
		assert.Equal(t, testData, response.Data)
		assert.True(t, response.SignatureValid)
		assert.WithinDuration(t, time.Now(), response.DecryptedAt, time.Second)

		mockService.AssertExpectations(t)
	})

	t.Run("successful_decryption_invalid_signature", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewDecryptDataCommandHandler(mockService)

		request := dto_crypto.DecryptionRequest{
			EncryptedData:       []byte("encrypted-data"),
			Algorithm:           testAlgorithm,
			RecipientPrivateKey: []byte("recipient-private"),
			SenderPublicKey:     []byte("sender-public"),
			Signature:           []byte("bad-signature"),
			Nonce:               []byte("nonce"),
		}

		mockService.On("Verify", ctx, request.EncryptedData, request.Signature, request.SenderPublicKey, testAlgorithm).
			Return(false, nil)
		mockService.On("Decrypt", ctx, mock.AnythingOfType("*crypto.CryptoPayload"), request.RecipientPrivateKey).
			Return(testData, nil)

		cmd := crypto.NewDecryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		require.NoError(t, err)
		assert.Equal(t, testData, response.Data)
		assert.False(t, response.SignatureValid)

		mockService.AssertExpectations(t)
	})

	t.Run("verification_failure", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewDecryptDataCommandHandler(mockService)

		request := dto_crypto.DecryptionRequest{
			EncryptedData:       []byte("encrypted-data"),
			Algorithm:           testAlgorithm,
			RecipientPrivateKey: []byte("recipient-private"),
			SenderPublicKey:     []byte("sender-public"),
			Signature:           []byte("signature"),
			Nonce:               []byte("nonce"),
		}

		expectedErr := errors.New("verification service error")
		mockService.On("Verify", ctx, request.EncryptedData, request.Signature, request.SenderPublicKey, testAlgorithm).
			Return(false, expectedErr)

		cmd := crypto.NewDecryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
		assert.Empty(t, response.Data)

		mockService.AssertExpectations(t)
	})

	t.Run("decryption_failure", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewDecryptDataCommandHandler(mockService)

		request := dto_crypto.DecryptionRequest{
			EncryptedData:       []byte("encrypted-data"),
			Algorithm:           testAlgorithm,
			RecipientPrivateKey: []byte("recipient-private"),
			SenderPublicKey:     []byte("sender-public"),
			Signature:           []byte("signature"),
			Nonce:               []byte("nonce"),
		}

		expectedErr := errors.New("decryption service error")
		mockService.On("Verify", ctx, request.EncryptedData, request.Signature, request.SenderPublicKey, testAlgorithm).
			Return(true, nil)
		mockService.On("Decrypt", ctx, mock.AnythingOfType("*crypto.CryptoPayload"), request.RecipientPrivateKey).
			Return(nil, expectedErr)

		cmd := crypto.NewDecryptDataCommand(request)
		response, err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decryption failed")
		assert.Empty(t, response.Data)

		mockService.AssertExpectations(t)
	})
}

// Test GenerateKeyPairCommand

func TestGenerateKeyPairCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request dto_crypto.KeyPairRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_kyber1024",
			request: dto_crypto.KeyPairRequest{
				Algorithm: domain_crypto.Kyber1024,
				KeySize:   domain_crypto.Kyber1024KeySize,
			},
			wantErr: false,
		},
		{
			name: "valid_without_keysize",
			request: dto_crypto.KeyPairRequest{
				Algorithm: domain_crypto.Kyber1024,
				KeySize:   0,
			},
			wantErr: false,
		},
		{
			name: "invalid_algorithm",
			request: dto_crypto.KeyPairRequest{
				Algorithm: domain_crypto.UnknownAlgo,
			},
			wantErr: true,
			errMsg:  "invalid algorithm",
		},
		{
			name: "invalid_keysize",
			request: dto_crypto.KeyPairRequest{
				Algorithm: domain_crypto.Kyber1024,
				KeySize:   999,
			},
			wantErr: true,
			errMsg:  "invalid key size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := crypto.NewGenerateKeyPairCommand(tt.request)
			err := cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateKeyPairCommandHandler_Handle(t *testing.T) {
	ctx := context.Background()
	testAlgorithm := domain_crypto.Kyber1024

	t.Run("successful_key_generation", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewGenerateKeyPairCommandHandler(mockService)

		request := dto_crypto.KeyPairRequest{
			Algorithm: testAlgorithm,
			KeySize:   domain_crypto.Kyber1024KeySize,
		}

		mockKeyPair := &domain_crypto.KeyPair{
			Algorithm:  testAlgorithm,
			PublicKey:  []byte("public-key-data"),
			PrivateKey: []byte("private-key-data"),
			CreatedAt:  time.Now(),
		}

		mockService.On("GenerateKeyPair", ctx, testAlgorithm).
			Return(mockKeyPair, nil)

		cmd := crypto.NewGenerateKeyPairCommand(request)
		response, err := handler.Handle(ctx, cmd)

		require.NoError(t, err)
		assert.Equal(t, mockKeyPair.PublicKey, response.PublicKey)
		assert.Equal(t, mockKeyPair.PrivateKey, response.PrivateKey)
		assert.Equal(t, testAlgorithm, response.Algorithm)
		assert.Equal(t, mockKeyPair.CreatedAt, response.GeneratedAt)

		mockService.AssertExpectations(t)
	})

	t.Run("key_generation_failure", func(t *testing.T) {
		mockService := new(MockCryptoService)
		handler := crypto.NewGenerateKeyPairCommandHandler(mockService)

		request := dto_crypto.KeyPairRequest{
			Algorithm: testAlgorithm,
		}

		expectedErr := errors.New("key generation service error")
		mockService.On("GenerateKeyPair", ctx, testAlgorithm).
			Return(nil, expectedErr)

		cmd := crypto.NewGenerateKeyPairCommand(request)
		response, err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key generation failed")
		assert.Empty(t, response.PublicKey)

		mockService.AssertExpectations(t)
	})
}
