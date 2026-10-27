package http

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"music-platform/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

const middlewareTestSecret = "test-secret-for-middleware-tests-32chars"

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing authorization header",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "authorization token is required",
		},
		{
			name:           "unsupported authorization scheme",
			authHeader:     "Basic abc123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header",
		},
		{
			name:           "missing bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header",
		},
		{
			name:           "malformed token",
			authHeader:     "Bearer invalid.token.value",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired access token",
		},
		{
			name:           "token signed with wrong secret",
			authHeader:     "Bearer " + generateMiddlewareTestToken(t, "42", time.Now().Add(time.Minute), "another-test-secret-that-is-at-least-32-characters"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired access token",
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + generateMiddlewareTestToken(t, "42", time.Now().Add(-time.Minute), middlewareTestSecret),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired access token",
		},
		{
			name:           "invalid subject",
			authHeader:     "Bearer " + generateMiddlewareTestToken(t, "not-a-user-id", time.Now().Add(time.Minute), middlewareTestSecret),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid access token subject",
		},
		{
			name:           "zero user ID",
			authHeader:     "Bearer " + generateMiddlewareTestToken(t, "0", time.Now().Add(time.Minute), middlewareTestSecret),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid access token subject",
		},
		{
			name:           "negative user ID",
			authHeader:     "Bearer " + generateMiddlewareTestToken(t, "-1", time.Now().Add(time.Minute), middlewareTestSecret),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid access token subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := (&Handler{
				jwtSecret: middlewareTestSecret,
			}).authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("next handler must not be called for an invalid token")
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("unexpected status code: got %d, want %d; body: %s",
					rec.Code, tt.expectedStatus, rec.Body.String())
			}

			if !strings.Contains(rec.Body.String(), tt.expectedBody) {
				t.Errorf("unexpected response body: got %q, want it to contain %q",
					rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	const expectedUserID int64 = 42

	token, err := auth.GenerateAccessToken(expectedUserID, middlewareTestSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() returned an error: %v", err)
	}

	var nextHandlerCalled bool

	handler := (&Handler{
		jwtSecret: middlewareTestSecret,
	}).authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true

		userID, ok := currentUserID(r)
		if !ok {
			t.Error("user ID was not found in request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if userID != expectedUserID {
			t.Errorf("unexpected user ID: got %d, want %d", userID, expectedUserID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !nextHandlerCalled {
		t.Fatal("next handler was not called for a valid token")
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("unexpected status code: got %d, want %d; body: %s",
			rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestAuthMiddlewareAcceptsCaseInsensitiveBearerScheme(t *testing.T) {
	token, err := auth.GenerateAccessToken(42, middlewareTestSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() returned an error: %v", err)
	}

	var nextHandlerCalled bool

	handler := (&Handler{
		jwtSecret: middlewareTestSecret,
	}).authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer "+token)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !nextHandlerCalled {
		t.Fatal("next handler was not called for a valid Bearer token")
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("unexpected status code: got %d, want %d",
			rec.Code, http.StatusNoContent)
	}
}

func generateMiddlewareTestToken(
	t *testing.T,
	subject string,
	expiresAt time.Time,
	secret string,
) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	return signedToken
}

func TestAuthMiddlewareParsesInt64UserID(t *testing.T) {
	expectedUserID := int64(2147483648)

	token, err := auth.GenerateAccessToken(expectedUserID, middlewareTestSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() returned an error: %v", err)
	}

	handler := (&Handler{
		jwtSecret: middlewareTestSecret,
	}).authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := currentUserID(r)
		if !ok {
			t.Error("user ID was not found in request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if userID != expectedUserID {
			t.Errorf("unexpected user ID: got %d, want %d",
				userID, expectedUserID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("unexpected status code: got %d, want %d",
			rec.Code, http.StatusNoContent)
	}

	if _, err := strconv.ParseInt("2147483648", 10, 64); err != nil {
		t.Fatalf("test value does not fit into int64: %v", err)
	}
}
