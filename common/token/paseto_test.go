package token

import (
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestManager(t *testing.T) *PasetoTokenManager {
	// Generate a test key pair
	privateKey := paseto.NewV4AsymmetricSecretKey()
	privateKeyHex := privateKey.ExportHex()

	service, err := NewPasetoTokenManager(privateKeyHex)
	require.NoError(t, err)
	require.NotNil(t, service)

	return service
}

func setupTestManagerAndParser(t *testing.T) (*PasetoTokenManager, *PasetoTokenParser) {
	privateKey := paseto.NewV4AsymmetricSecretKey()
	privateKeyHex := privateKey.ExportHex()

	manager, err := NewPasetoTokenManager(privateKeyHex)
	require.NoError(t, err)
	require.NotNil(t, manager)

	parser, err := NewPasetoTokenParser(manager.GetPublicKey())
	require.NoError(t, err)
	require.NotNil(t, parser)

	return manager, parser
}

func TestNewPasetoTokenManager(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
		wantErr    bool
	}{
		{
			name:       "Valid private key",
			privateKey: paseto.NewV4AsymmetricSecretKey().ExportHex(),
			wantErr:    false,
		},
		{
			name:       "Invalid private key",
			privateKey: "invalid-key",
			wantErr:    true,
		},
		{
			name:       "Empty private key",
			privateKey: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewPasetoTokenManager(tt.privateKey)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotEmpty(t, service.GetPublicKey())
			}
		})
	}
}

func TestPasetoTokenSvc_GenerateToken(t *testing.T) {
	service := setupTestManager(t)

	tests := []struct {
		name    string
		claims  TokenClaims
		wantErr bool
	}{
		{
			name: "Valid claims",
			claims: TokenClaims{
				Sub:       "user123",
				UserId:    "456",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "Empty subject",
			claims: TokenClaims{
				UserId:    "456",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "Expired token",
			claims: TokenClaims{
				Sub:       "user123",
				UserId:    "456",
				IssuedAt:  time.Now().Add(-2 * time.Hour),
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			wantErr: false, // Generation should succeed even if expired
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.claims)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestPasetoTokenSvc_ValidateToken(t *testing.T) {
	service := setupTestManager(t)

	validClaims := TokenClaims{
		Sub:       "user123",
		UserId:    "456",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	expiredClaims := TokenClaims{
		Sub:       "user123",
		UserId:    "456",
		IssuedAt:  time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	validToken, err := service.GenerateToken(validClaims)
	require.NoError(t, err)

	expiredToken, err := service.GenerateToken(expiredClaims)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "Expired token",
			token:   expiredToken,
			wantErr: true,
		},
		{
			name:    "Invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, validClaims.Sub, claims.Sub)
				assert.Equal(t, validClaims.UserId, claims.UserId)
			}
		})
	}
}

func TestPasetoTokenSvc_GetPublicKey(t *testing.T) {
	service := setupTestManager(t)

	publicKey := service.GetPublicKey()
	assert.NotEmpty(t, publicKey)

	// Verify that the public key is valid by creating a verifier
	verifier, err := paseto.NewV4AsymmetricPublicKeyFromHex(publicKey)
	assert.NoError(t, err)
	assert.NotNil(t, verifier)
}

func TestTokenValidationAcrossInstances(t *testing.T) {
	// Create first service instance
	service1 := setupTestManager(t)

	// Create second service instance with only public key
	service2, err := NewPasetoTokenManager(service1.privateKey.ExportHex())
	require.NoError(t, err)

	claims := TokenClaims{
		Sub:       "user123",
		UserId:    "456",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Generate token with first service
	token, err := service1.GenerateToken(claims)
	require.NoError(t, err)

	// Validate token with second service
	validatedClaims, err := service2.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, claims.Sub, validatedClaims.Sub)
	assert.Equal(t, claims.UserId, validatedClaims.UserId)
}

// Parse token
func TestNewPasetoTokenParser(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		wantErr   bool
	}{
		{
			name:      "Valid public key",
			publicKey: paseto.NewV4AsymmetricSecretKey().Public().ExportHex(),
			wantErr:   false,
		},
		{
			name:      "Invalid public key",
			publicKey: "invalid-key",
			wantErr:   true,
		},
		{
			name:      "Empty public key",
			publicKey: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewPasetoTokenParser(tt.publicKey)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, parser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parser)
			}
		})
	}
}

func TestPasetoTokenParser_ParseToken(t *testing.T) {
	manager, parser := setupTestManagerAndParser(t)

	claims := TokenClaims{
		Sub:       "user123",
		UserId:    "456",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	token, err := manager.GenerateToken(claims)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "Invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := parser.ParseToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
