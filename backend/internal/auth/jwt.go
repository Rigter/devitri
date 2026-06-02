package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TokenTTL is the time-to-live for JWT tokens (24 hours)
	TokenTTL = 24 * time.Hour
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token has expired
	ErrExpiredToken = errors.New("expired token")
)

// JWTManager handles JWT token generation and verification
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWTManager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{secretKey: secretKey, tokenDuration: tokenDuration}
}

// Generate generates a new JWT token
func (manager *JWTManager) Generate() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(manager.tokenDuration).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Verify verifies a JWT token and returns the claims
func (manager *JWTManager) Verify(accessToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}
		return []byte(manager.secretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > int64(exp) {
		return nil, ErrExpiredToken
	}

	return token, nil
}