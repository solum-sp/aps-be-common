package token

import "time"

type TokenClaims struct {
	Sub       string    `json:"sub"`
	SessionId string    `json:"sessionId"`
	UserId    string    `json:"userId"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

type ITokenManager interface {
	GenerateToken(data TokenClaims) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
}

type ITokenParser interface {
	ParseToken(token string) (*TokenClaims, error)
}
