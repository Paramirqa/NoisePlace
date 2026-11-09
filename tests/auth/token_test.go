package auth_test

import (
	"testing"
	"time"

	"music-platform/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-for-jwt-unit-tests-32chars"

func TestGenerateAccessToken(t *testing.T) {
	userID := int64(42)

	tokenString, err := auth.GenerateAccessToken(userID, testJWTSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() returned an error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("GenerateAccessToken() returned an empty token")
	}

	claims := &auth.AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				t.Fatalf("unexpected signing method: %v", token.Method.Alg())
			}

			return []byte(testJWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	if !token.Valid {
		t.Fatal("generated token is invalid")
	}

	if claims.Subject != "42" {
		t.Errorf("unexpected subject: got %q, want %q", claims.Subject, "42")
	}

	if claims.IssuedAt == nil {
		t.Fatal("issued-at claim is missing")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expiration claim is missing")
	}

	if !claims.ExpiresAt.Time.After(claims.IssuedAt.Time) {
		t.Error("expiration must be later than issue time")
	}

	if got := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); got != auth.AccessTokenTTL {
		t.Errorf("unexpected token lifetime: got %v, want %v", got, auth.AccessTokenTTL)
	}
}

func TestGenerateAccessTokenRejectsNonPositiveUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
	}{
		{
			name:   "zero user ID",
			userID: 0,
		},
		{
			name:   "negative user ID",
			userID: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateAccessToken(tt.userID, testJWTSecret)

			if err == nil {
				t.Fatal("expected an error for non-positive user ID")
			}

			if token != "" {
				t.Fatal("expected an empty token for invalid user ID")
			}
		})
	}
}

func TestGenerateAccessTokenRejectsShortSecret(t *testing.T) {
	token, err := auth.GenerateAccessToken(42, "short-secret")

	if err == nil {
		t.Fatal("expected an error for a secret shorter than 32 characters")
	}

	if token != "" {
		t.Fatal("expected an empty token for an invalid secret")
	}
}

func TestGenerateAccessTokenRejectsWrongSecret(t *testing.T) {
	tokenString, err := auth.GenerateAccessToken(42, testJWTSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() returned an error: %v", err)
	}

	wrongSecret := "another-test-secret-that-is-at-least-32-characters"

	claims := &auth.AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(wrongSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)

	if err == nil && token.Valid {
		t.Fatal("token should not be valid when verified with the wrong secret")
	}
}

func TestAccessTokenTTL(t *testing.T) {
	if auth.AccessTokenTTL != 24 * time.Hour {
		t.Errorf("unexpected AccessTokenTTL: got %v, want %v", auth.AccessTokenTTL, 24 * time.Hour)
	}
}
