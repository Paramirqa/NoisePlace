package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"music-platform/internal/database"
	apihttp "music-platform/internal/http"
)

func TestRatingsIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping integration tests")
	}

	ctx := context.Background()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(db.Close)

	server := httptest.NewServer(apihttp.NewRouter(db, testJWTSecret))
	t.Cleanup(server.Close)

	// Create two independent users.
	user1Token := registerAndLogin(t, server.URL, "rating_user_1")
	user2Token := registerAndLogin(t, server.URL, "rating_user_2")

	// Create a test artist.
	var artistID int64
	err = db.QueryRow(ctx, `
		INSERT INTO artists (name, normalized_name, sort_name)
		VALUES ('Integration Rating Artist', 'integration rating artist', 'Integration Rating Artist')
		RETURNING id
	`).Scan(&artistID)
	if err != nil {
		t.Fatalf("create test artist: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(context.Background(),
			`DELETE FROM artists WHERE id = $1`, artistID)
		if err != nil {
			t.Errorf("delete test artist: %v", err)
		}
	})

	artistRatingURL := fmt.Sprintf("/api/v1/me/ratings/artists/%d", artistID)
	artistStatsURL := fmt.Sprintf("/api/v1/artists/%d/rating", artistID)

	t.Run("user_cannot_access_another_users_rating", func(t *testing.T) {
		// User 1 creates a rating.
		response := requestJSON(t, server.URL, http.MethodPut,
			artistRatingURL, map[string]int{"rating": 8}, user1Token)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("user 1 rating: expected 200, got %d; body: %s",
				response.StatusCode, response.Body)
		}

		// User 2 must not see user 1's rating.
		response = requestJSON(t, server.URL, http.MethodGet,
			artistRatingURL, nil, user2Token)

		if response.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 when user 2 requests user 1 rating, got %d",
				response.StatusCode)
		}

		// User 2 must not delete user 1's rating.
		response = requestJSON(t, server.URL, http.MethodDelete,
			artistRatingURL, nil, user2Token)

		if response.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 when user 2 deletes user 1 rating, got %d",
				response.StatusCode)
		}

		// User 1's rating must still exist.
		response = requestJSON(t, server.URL, http.MethodGet,
			artistRatingURL, nil, user1Token)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("user 1 rating should still exist, got %d; body: %s",
				response.StatusCode, response.Body)
		}
	})

	t.Run("updating_rating_recalculates_average_without_increasing_count", func(t *testing.T) {
		// User 2 adds rating 10.
		response := requestJSON(t, server.URL, http.MethodPut,
			artistRatingURL, map[string]int{"rating": 10}, user2Token)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("user 2 rating: expected 200, got %d; body: %s",
				response.StatusCode, response.Body)
		}

		assertArtistStats(t, server.URL, artistStatsURL, 9, 2)

		// User 1 changes rating from 8 to 6.
		response = requestJSON(t, server.URL, http.MethodPut,
			artistRatingURL, map[string]int{"rating": 6}, user1Token)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("update rating: expected 200, got %d; body: %s",
				response.StatusCode, response.Body)
		}

		assertArtistStats(t, server.URL, artistStatsURL, 8, 2)
	})

	t.Run("deleting_rating_recalculates_average_and_count", func(t *testing.T) {
		response := requestJSON(t, server.URL, http.MethodDelete,
			artistRatingURL, nil, user1Token)

		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("delete rating: expected 204, got %d; body: %s",
				response.StatusCode, response.Body)
		}

		assertArtistStats(t, server.URL, artistStatsURL, 10, 1)

		// Deleted rating must no longer be available to user 1.
		response = requestJSON(t, server.URL, http.MethodGet,
			artistRatingURL, nil, user1Token)

		if response.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 after deleting rating, got %d",
				response.StatusCode)
		}
	})
}

func registerAndLogin(t *testing.T, baseURL, suffix string) string {
	t.Helper()

	identity := fmt.Sprintf("%s_%d", suffix, os.Getpid())
	email := identity + "@example.com"

	registerResponse := requestJSON(t, baseURL, http.MethodPost,
		"/api/v1/auth/register",
		map[string]string{
			"username": identity,
			"email":    email,
			"password": "integration-test-password",
		}, "")

	if registerResponse.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: expected 201, got %d; body: %s",
			identity, registerResponse.StatusCode, registerResponse.Body)
	}

	loginResponse := requestJSON(t, baseURL, http.MethodPost,
		"/api/v1/auth/login",
		map[string]string{
			"email":    email,
			"password": "integration-test-password",
		}, "")

	if loginResponse.StatusCode != http.StatusOK {
		t.Fatalf("login %s: expected 200, got %d; body: %s",
			identity, loginResponse.StatusCode, loginResponse.Body)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(loginResponse.Body, &result); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	if result.AccessToken == "" {
		t.Fatalf("login %s returned empty access token", identity)
	}

	return result.AccessToken
}

func assertArtistStats(
	t *testing.T,
	baseURL string,
	path string,
	expectedAverage float64,
	expectedCount int64,
) {
	t.Helper()

	response := requestJSON(t, baseURL, http.MethodGet, path, nil, "")

	if response.StatusCode != http.StatusOK {
		t.Fatalf("get artist stats: expected 200, got %d; body: %s",
			response.StatusCode, response.Body)
	}

	var stats struct {
		AverageRating float64 `json:"average_rating"`
		RatingsCount  int64   `json:"ratings_count"`
	}

	if err := json.Unmarshal(response.Body, &stats); err != nil {
		t.Fatalf("decode artist stats: %v", err)
	}

	if stats.AverageRating != expectedAverage {
		t.Errorf("expected average %.2f, got %.2f",
			expectedAverage, stats.AverageRating)
	}

	if stats.RatingsCount != expectedCount {
		t.Errorf("expected ratings count %d, got %d",
			expectedCount, stats.RatingsCount)
	}
}
