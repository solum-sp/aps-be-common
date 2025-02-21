package token

import (
	"fmt"

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

	// Sign the token with the private key
	signed := token.V4Sign(m.privateKey, nil)
	return signed, nil
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

	return claims, nil
}
