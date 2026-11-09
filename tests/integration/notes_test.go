package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"music-platform/internal/database"
	apihttp "music-platform/internal/http"
)

func TestNotesIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration tests")
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" || !strings.HasSuffix(dbName, "_test") {
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

	user1Token := registerAndLogin(t, server.URL, "notes_user_1")
	user2Token := registerAndLogin(t, server.URL, "notes_user_2")

	var artistID int64
	err = db.QueryRow(ctx, `
		INSERT INTO artists (name, normalized_name, sort_name)
		VALUES (
			'Integration Notes Artist',
			'integration notes artist',
			'Integration Notes Artist'
		)
		RETURNING id
	`).Scan(&artistID)
	if err != nil {
		t.Fatalf("create test artist: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			`DELETE FROM artists WHERE id = $1`,
			artistID,
		)
		if err != nil {
			t.Errorf("delete test artist: %v", err)
		}
	})

	var releaseID int64
	err = db.QueryRow(ctx, `
		INSERT INTO releases (title, release_type)
		VALUES ('Integration Notes Release', 'album')
		RETURNING id
	`).Scan(&releaseID)
	if err != nil {
		t.Fatalf("create test release: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			`DELETE FROM releases WHERE id = $1`,
			releaseID,
		)
		if err != nil {
			t.Errorf("delete test release: %v", err)
		}
	})

	artistNoteURL := fmt.Sprintf(
		"/api/v1/me/notes/artists/%d",
		artistID,
	)
	releaseNoteURL := fmt.Sprintf(
		"/api/v1/me/notes/releases/%d",
		releaseID,
	)

	t.Run("artist_notes", func(t *testing.T) {
		t.Run("requires_authentication", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				"",
			)

			assertNotesStatus(t, response, http.StatusUnauthorized)
		})

		t.Run("create_and_get_note", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPost,
				artistNoteURL,
				map[string]string{"content": "My personal artist note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusCreated)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "My personal artist note")
		})

		t.Run("duplicate_creation_returns_conflict", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPost,
				artistNoteURL,
				map[string]string{"content": "Another note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusConflict)

			// The original note must remain unchanged.
			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "My personal artist note")
		})

		t.Run("another_user_cannot_access_note", func(t *testing.T) {
			methods := []string{
				http.MethodGet,
				http.MethodPut,
				http.MethodDelete,
			}

			for _, method := range methods {
				t.Run(strings.ToLower(method), func(t *testing.T) {
					response := requestJSON(
						t,
						server.URL,
						method,
						artistNoteURL,
						map[string]string{"content": "Attempt to modify"},
						user2Token,
					)

					assertNotesStatus(t, response, http.StatusNotFound)
				})
			}

			// User 1's note must remain unchanged.
			response := requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "My personal artist note")
		})

		t.Run("update_note", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPut,
				artistNoteURL,
				map[string]string{"content": "Updated artist note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNoContent)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "Updated artist note")
		})

		t.Run("delete_note", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodDelete,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNoContent)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNotFound)

			response = requestJSON(
				t,
				server.URL,
				http.MethodDelete,
				artistNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNotFound)
		})

		t.Run("rejects_invalid_content", func(t *testing.T) {
			invalidContents := []struct {
				name    string
				content string
			}{
				{
					name:    "empty",
					content: "",
				},
				{
					name:    "whitespace_only",
					content: "   \n\t ",
				},
				{
					name:    "too_long",
					content: strings.Repeat("a", 5001),
				},
			}

			for _, testCase := range invalidContents {
				t.Run(testCase.name, func(t *testing.T) {
					response := requestJSON(
						t,
						server.URL,
						http.MethodPost,
						artistNoteURL,
						map[string]string{"content": testCase.content},
						user1Token,
					)

					assertNotesStatus(t, response, http.StatusBadRequest)
				})
			}
		})

		t.Run("rejects_malformed_json", func(t *testing.T) {
			response := requestNotesRaw(
				t,
				server.URL,
				http.MethodPost,
				artistNoteURL,
				`{"content":`,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusBadRequest)
		})

		t.Run("rejects_unknown_json_fields", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPost,
				artistNoteURL,
				map[string]string{
					"content": "Valid content",
					"extra":   "Unexpected field",
				},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusBadRequest)
		})
	})

	t.Run("release_notes", func(t *testing.T) {
		t.Run("note_lifecycle", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPost,
				releaseNoteURL,
				map[string]string{"content": "My release note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusCreated)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				releaseNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "My release note")

			response = requestJSON(
				t,
				server.URL,
				http.MethodPut,
				releaseNoteURL,
				map[string]string{"content": "Updated release note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNoContent)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				releaseNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "Updated release note")

			response = requestJSON(
				t,
				server.URL,
				http.MethodDelete,
				releaseNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNoContent)

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				releaseNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusNotFound)
		})

		t.Run("another_user_cannot_access_note", func(t *testing.T) {
			response := requestJSON(
				t,
				server.URL,
				http.MethodPost,
				releaseNoteURL,
				map[string]string{"content": "Private release note"},
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusCreated)

			for _, method := range []string{
				http.MethodGet,
				http.MethodPut,
				http.MethodDelete,
			} {
				t.Run(strings.ToLower(method), func(t *testing.T) {
					response := requestJSON(
						t,
						server.URL,
						method,
						releaseNoteURL,
						map[string]string{"content": "Unauthorized change"},
						user2Token,
					)

					assertNotesStatus(t, response, http.StatusNotFound)
				})
			}

			response = requestJSON(
				t,
				server.URL,
				http.MethodGet,
				releaseNoteURL,
				nil,
				user1Token,
			)

			assertNotesStatus(t, response, http.StatusOK)
			assertNoteContent(t, response.Body, "Private release note")
		})
	})

	t.Run("nonexistent_entities_return_not_found", func(t *testing.T) {
		artistURL := "/api/v1/me/notes/artists/9223372036854775807"
		releaseURL := "/api/v1/me/notes/releases/9223372036854775807"

		response := requestJSON(
			t,
			server.URL,
			http.MethodPost,
			artistURL,
			map[string]string{"content": "Note"},
			user1Token,
		)
		assertNotesStatus(t, response, http.StatusNotFound)

		response = requestJSON(
			t,
			server.URL,
			http.MethodPost,
			releaseURL,
			map[string]string{"content": "Note"},
			user1Token,
		)
		assertNotesStatus(t, response, http.StatusNotFound)
	})
}

func assertNotesStatus(
	t *testing.T,
	response httpResponse,
	expectedStatus int,
) {
	t.Helper()

	if response.StatusCode != expectedStatus {
		t.Fatalf(
			"expected status %d, got %d; body: %s",
			expectedStatus,
			response.StatusCode,
			response.Body,
		)
	}
}

func assertNoteContent(
	t *testing.T,
	body []byte,
	expectedContent string,
) {
	t.Helper()

	var note struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(body, &note); err != nil {
		t.Fatalf("decode note response: %v; body: %s", err, body)
	}

	if note.Content != expectedContent {
		t.Errorf(
			"expected note content %q, got %q",
			expectedContent,
			note.Content,
		)
	}
}

func requestNotesRaw(
	t *testing.T,
	baseURL string,
	method string,
	path string,
	body string,
	token string,
) httpResponse {
	t.Helper()

	req, err := http.NewRequest(
		method,
		baseURL+path,
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("send %s %s: %v", method, path, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	return httpResponse{
		StatusCode: response.StatusCode,
		Body:       responseBody,
	}
}
