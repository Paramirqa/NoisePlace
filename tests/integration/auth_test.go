package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"music-platform/internal/database"
	apihttp "music-platform/internal/http"
)

const testJWTSecret = "integration-test-jwt-secret-at-least-32-bytes"

func TestAuthIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration tests")
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if !strings.HasSuffix(dbName, "_test") {
		t.Fatalf(
			"refusing to run integration tests against database %q: database name must end in _test",
			dbName,
		)
	}

	ctx := context.Background()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(db.Close)

	server := httptest.NewServer(apihttp.NewRouter(db, testJWTSecret))
	t.Cleanup(server.Close)

	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		t.Fatalf("generate unique test identity: %v", err)
	}

	suffix := hex.EncodeToString(randomBytes)
	username := "it_" + suffix
	email := "it_" + suffix + "@example.com"
	password := "integration-test-password"

	t.Run("register", func(t *testing.T) {
		body := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}

		response := requestJSON(t, server.URL, http.MethodPost,
			"/api/v1/auth/register", body, "")

		if response.StatusCode != http.StatusCreated {
			t.Fatalf("register: expected status %d, got %d; body: %s",
				http.StatusCreated, response.StatusCode, response.Body)
		}

		var user map[string]any
		if err := json.Unmarshal(response.Body, &user); err != nil {
			t.Fatalf("decode registration response: %v", err)
		}

		if user["username"] != username {
			t.Errorf("expected username %q, got %v", username, user["username"])
		}
		if user["email"] != email {
			t.Errorf("expected email %q, got %v", email, user["email"])
		}
		if _, exists := user["password_hash"]; exists {
			t.Error("registration response must not expose password_hash")
		}
	})

	t.Run("duplicate_registration_returns_conflict", func(t *testing.T) {
		body := map[string]string{
			"username": username,
			"email":    email,
			"password": password,
		}

		response := requestJSON(t, server.URL, http.MethodPost,
			"/api/v1/auth/register", body, "")

		if response.StatusCode != http.StatusConflict {
			t.Fatalf("expected status %d for duplicate registration, got %d; body: %s",
				http.StatusConflict, response.StatusCode, response.Body)
		}
	})

	t.Run("invalid_login_returns_unauthorized", func(t *testing.T) {
		body := map[string]string{
			"email":    email,
			"password": "wrong-password",
		}

		response := requestJSON(t, server.URL, http.MethodPost,
			"/api/v1/auth/login", body, "")

		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d; body: %s",
				http.StatusUnauthorized, response.StatusCode, response.Body)
		}
	})

	var accessToken string

	t.Run("login", func(t *testing.T) {
		body := map[string]string{
			"email":    email,
			"password": password,
		}

		response := requestJSON(t, server.URL, http.MethodPost,
			"/api/v1/auth/login", body, "")

		if response.StatusCode != http.StatusOK {
			t.Fatalf("login: expected status %d, got %d; body: %s",
				http.StatusOK, response.StatusCode, response.Body)
		}

		var result struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
		}
		if err := json.Unmarshal(response.Body, &result); err != nil {
			t.Fatalf("decode login response: %v", err)
		}

		if result.AccessToken == "" {
			t.Fatal("login response contains an empty access_token")
		}
		if result.TokenType != "Bearer" {
			t.Errorf("expected token_type Bearer, got %q", result.TokenType)
		}
		if result.ExpiresIn != 900 {
			t.Errorf("expected expires_in 900, got %d", result.ExpiresIn)
		}

		accessToken = result.AccessToken
	})

	t.Run("current_user_requires_authentication", func(t *testing.T) {
		response := requestJSON(t, server.URL, http.MethodGet,
			"/api/v1/users/me", nil, "")

		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status %d without token, got %d; body: %s",
				http.StatusUnauthorized, response.StatusCode, response.Body)
		}
	})

	t.Run("current_user_returns_authenticated_user", func(t *testing.T) {
		if accessToken == "" {
			t.Fatal("login did not provide an access token")
		}

		response := requestJSON(t, server.URL, http.MethodGet,
			"/api/v1/users/me", nil, accessToken)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d; body: %s",
				http.StatusOK, response.StatusCode, response.Body)
		}

		var user map[string]any
		if err := json.Unmarshal(response.Body, &user); err != nil {
			t.Fatalf("decode current user response: %v", err)
		}

		if user["username"] != username {
			t.Errorf("expected username %q, got %v", username, user["username"])
		}
		if user["email"] != email {
			t.Errorf("expected email %q, got %v", email, user["email"])
		}
		if _, exists := user["password_hash"]; exists {
			t.Error("current user response must not expose password_hash")
		}
	})
}

type httpResponse struct {
	StatusCode int
	Body       []byte
}

func requestJSON(
	t *testing.T,
	baseURL string,
	method string,
	path string,
	body any,
	token string,
) httpResponse {
	t.Helper()

	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, baseURL+path, &requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("send %s %s: %v", method, path, err)
	}
	defer response.Body.Close()

	var responseBody bytes.Buffer
	if _, err := responseBody.ReadFrom(response.Body); err != nil {
		t.Fatalf("read response body: %v", err)
	}

	return httpResponse{
		StatusCode: response.StatusCode,
		Body:       responseBody.Bytes(),
	}
}
