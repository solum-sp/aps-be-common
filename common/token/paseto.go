package token

import (
	"encoding/json"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
)

type PasetoTokenManager struct {
	privateKey paseto.V4AsymmetricSecretKey
	publicKey  paseto.V4AsymmetricPublicKey
}

var _ ITokenManager = (*PasetoTokenManager)(nil)

func NewPasetoTokenManager(privateKey string) (*PasetoTokenManager, error) {
	privateK, err := paseto.NewV4AsymmetricSecretKeyFromHex(privateKey)
	if err != nil {
		return nil, err
	}

	return &PasetoTokenManager{
		privateKey: privateK,
		publicKey:  privateK.Public(),
	}, nil
}

func (m *PasetoTokenManager) GenerateToken(data TokenClaims) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(data.IssuedAt)
	token.SetExpiration(data.ExpiresAt)
	token.SetSubject(data.Sub)
	token.Set("userId", data.UserId)
	token.Set("sessionId", data.SessionId)
	token.Set("isAdmin", data.IsSuperAdmin)
	token.Set("customClaims", data.CustomClaims)

	// Sign the token with the private key
	signed := token.V4Sign(m.privateKey, nil)
	return signed, nil
}

func (m *PasetoTokenManager) GenerateCustomToken(data any, expireTime time.Time) (string, error) {
	// Serialize the custom data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize custom data: %w", err)
	}

	// Create a new PASETO token
	token := paseto.NewToken()

	// Set standard claims
	token.SetIssuedAt(time.Now())
	token.SetExpiration(expireTime)

	// Store the serialized data in a claim named "data"
	token.Set("data", string(dataBytes))

	// Sign the token with the private key
	signed := token.V4Sign(m.privateKey, nil)
	return signed, nil
}

// ValidateCustomToken with normalization for roles slice
func (m *PasetoTokenManager) ValidateCustomToken(token string) (any, error) {
	// Create a new parser
	parser := paseto.NewParser()

	// Parse and verify the token with the public key
	parsedToken, err := parser.ParseV4Public(m.publicKey, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom token: %w", err)
	}

	// Extract the "data" claim
	var dataStr string
	if err := parsedToken.Get("data", &dataStr); err != nil {
		return nil, fmt.Errorf("failed to get custom data: %w", err)
	}

	// Deserialize the JSON data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize custom data: %w", err)
	}

	// Normalize the "roles" field if it exists
	if roles, ok := data["roles"]; ok {
		rolesSlice, ok := roles.([]interface{})
		if !ok {
			return nil, fmt.Errorf("roles field is not a slice")
		}
		stringRoles := make([]string, len(rolesSlice))
		for i, role := range rolesSlice {
			strRole, ok := role.(string)
			if !ok {
				return nil, fmt.Errorf("role at index %d is not a string", i)
			}
			stringRoles[i] = strRole
		}
		data["roles"] = stringRoles
	}

	return data, nil
}

func (m *PasetoTokenManager) ValidateToken(t string) (*TokenClaims, error) {
	parser := paseto.NewParser()

	// Parse and verify the token
	token, err := parser.ParseV4Public(m.publicKey, t, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, err := parseToClaims(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims, nil
}

func (m *PasetoTokenManager) GetPublicKey() string {
	return m.publicKey.ExportHex()
}

type PasetoTokenParser struct {
	publicKey paseto.V4AsymmetricPublicKey
	parser    paseto.Parser
}

var _ ITokenParser = (*PasetoTokenParser)(nil)

func NewPasetoTokenParser(publicKey string) (*PasetoTokenParser, error) {
	publicK, err := paseto.NewV4AsymmetricPublicKeyFromHex(publicKey)
	if err != nil {
		return nil, err
	}

	return &PasetoTokenParser{
		publicKey: publicK,
		parser:    paseto.NewParser(),
	}, nil
}

func (p *PasetoTokenParser) ParseToken(t string) (*TokenClaims, error) {
	token, err := p.parser.ParseV4Public(p.publicKey, t, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, err := parseToClaims(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims, nil
}

func parseToClaims(token *paseto.Token) (*TokenClaims, error) {
	claims := &TokenClaims{}

	sub, err := token.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}

	claims.Sub = sub

	var userId string
	if err := token.Get("userId", &userId); err != nil {
		return nil, fmt.Errorf("failed to get userId: %w", err)
	}

	claims.UserId = userId

	var sessionId string
	if err := token.Get("sessionId", &sessionId); err != nil {
		return nil, fmt.Errorf("failed to get sessionId: %w", err)
	}

	claims.SessionId = sessionId

	issuedAt, err := token.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("failed to get issuedAt: %w", err)
	}

	claims.IssuedAt = issuedAt

	expiresAt, err := token.GetExpiration()
	if err != nil {
		return nil, fmt.Errorf("failed to get expiresAt: %w", err)
	}

	claims.ExpiresAt = expiresAt

	var isAdmin bool
	if err := token.Get("isAdmin", &isAdmin); err != nil {
		return nil, fmt.Errorf("failed to get isAdmin: %w", err)
	}

	claims.IsSuperAdmin = isAdmin

	var customClaims map[string]interface{}
	if err := token.Get("customClaims", &customClaims); err != nil {
		return nil, fmt.Errorf("failed to get customClaims: %w", err)
	}

	claims.CustomClaims = customClaims

	return claims, nil
}
