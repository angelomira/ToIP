package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Validator interface {
	Sign(userID int64, email, role, tokenType string, ttl time.Duration) (string, error)
	Parse(tokenStr, expectedType string) (jwt.MapClaims, error)
}

type HS256 struct {
	secret []byte
}

func NewHS256(secret []byte) *HS256 {
	return &HS256{secret: secret}
}

func (h *HS256) Sign(userID int64, email, role, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"type":  tokenType, // "access" или "refresh"
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
		"iss":   "pz10-auth",
		"aud":   "pz10-clients",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.secret)
}

func (h *HS256) Parse(tokenStr, expectedType string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return h.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithAudience("pz10-clients"),
		jwt.WithIssuer("pz10-auth"),
	)

	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Проверяем тип токена (защита от использования refresh вместо access)
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}