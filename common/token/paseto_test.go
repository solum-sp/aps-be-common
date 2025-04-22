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
				Sub:          "user123",
				UserId:       "456",
				IssuedAt:     time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour),
				IsSuperAdmin: true,
				CustomClaims: map[string]interface{}{
					"custom": "claim",
				},
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
		Sub:          "user123",
		UserId:       "456",
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
		IsSuperAdmin: true,
		CustomClaims: map[string]interface{}{
			"custom": "claim",
		},
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
	assert.Equal(t, claims.IsSuperAdmin, validatedClaims.IsSuperAdmin)
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

// TestPasetoTokenSvc_GenerateCustomToken tests the GenerateCustomToken function.
func TestPasetoTokenSvc_GenerateCustomToken(t *testing.T) {
	service := setupTestManager(t)

	tests := []struct {
		name       string
		data       any
		expireTime time.Time
		wantErr    bool
	}{
		{
			name: "Valid custom data",
			data: map[string]interface{}{
				"userId": "12345",
				"roles":  []string{"admin", "user"},
			},
			expireTime: time.Now().Add(time.Hour),
			wantErr:    false,
		},
		{
			name:       "Empty custom data",
			data:       nil,
			expireTime: time.Now().Add(time.Hour),
			wantErr:    false, // nil data is valid (serializes to "null")
		},
		{
			name:       "Unserializable data",
			data:       func() {}, // Functions are not JSON-serializable
			expireTime: time.Now().Add(time.Hour),
			wantErr:    true,
		},
		{
			name: "Expiration in the past",
			data: map[string]interface{}{
				"userId": "12345",
			},
			expireTime: time.Now().Add(-1 * time.Hour),
			wantErr:    false, // Generation should succeed; expiration checked during validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateCustomToken(tt.data, tt.expireTime)
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

// TestPasetoTokenSvc_ValidateCustomToken tests the ValidateCustomToken function.
func TestPasetoTokenSvc_ValidateCustomToken(t *testing.T) {
	service := setupTestManager(t)

	validData := map[string]interface{}{
		"userId": "12345",
		"roles":  []string{"admin", "user"},
	}
	validExpireTime := time.Now().Add(time.Hour)
	validToken, err := service.GenerateCustomToken(validData, validExpireTime)
	require.NoError(t, err)

	expiredData := map[string]interface{}{
		"userId": "67890",
	}
	expiredExpireTime := time.Now().Add(-1 * time.Hour)
	expiredToken, err := service.GenerateCustomToken(expiredData, expiredExpireTime)
	require.NoError(t, err)

	// Create a token with invalid data claim (manually manipulate token for testing)
	invalidDataToken := createTokenWithInvalidData(t, service)

	tests := []struct {
		name     string
		token    string
		wantErr  bool
		wantData any
	}{
		{
			name:     "Valid token",
			token:    validToken,
			wantErr:  false,
			wantData: validData,
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
		{
			name:    "Token with invalid data claim",
			token:   invalidDataToken,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := service.ValidateCustomToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				// Compare the deserialized data with the expected data
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}

// createTokenWithInvalidData creates a token with a corrupted "data" claim for testing.
func createTokenWithInvalidData(t *testing.T, service *PasetoTokenManager) string {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(time.Hour))
	// Set invalid JSON in the "data" claim
	token.Set("data", "{invalid-json")

	// Sign the token with the private key
	signed := token.V4Sign(service.privateKey, nil)
	return signed
}
