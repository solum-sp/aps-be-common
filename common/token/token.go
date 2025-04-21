package token

import "time"

type TokenClaims struct {
	Sub          string                 `json:"sub"`
	SessionId    string                 `json:"sessionId"`
	UserId       string                 `json:"userId"`
	ExpiresAt    time.Time              `json:"exp"`
	IssuedAt     time.Time              `json:"iat"`
	IsSuperAdmin bool                   `json:"isAdmin"`
	CustomClaims map[string]interface{} `json:"customClaims"`
}

type ITokenManager interface {
	GenerateToken(data TokenClaims) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
	GenerateCustomToken(data any, expAt time.Time) (string, error)
	ValidateCustomToken(token string) (any, error)
}

type ITokenParser interface {
	ParseToken(token string) (*TokenClaims, error)
}
