package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 24 * time.Hour

type AccessTokenClaims struct {
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, secret string) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("user ID must be positive")
	}

	if len(secret) < 32 {
		return "", fmt.Errorf("JWT secret must contain at least 32 characters")
	}
	now := time.Now()
	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signedToken, nil

}
