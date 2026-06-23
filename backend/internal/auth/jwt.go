package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type TokenManager struct {
	secret string
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: secret,
		ttl:    ttl,
	}
}

func (manager *TokenManager) IssueToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(manager.ttl)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID.String(),
		Email:  email,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(manager.secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, expiresAt, nil
}

func (manager *TokenManager) ParseToken(tokenString string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(manager.secret), nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token claims")
	}

	return *claims, nil
}

func (claims Claims) Principal() (Principal, error) {
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Principal{}, fmt.Errorf("invalid user id in token: %w", err)
	}

	return Principal{
		UserID: userID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
