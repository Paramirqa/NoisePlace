package http

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type ratingStatsResponse struct {
	AverageRating float64 `json:"average_rating"`
	RatingsCount  int64   `json:"ratings_count"`
}

func (h *Handler) getArtistRatingStats(w http.ResponseWriter, r *http.Request) {
	artistID, ok := parseRatingID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	// Check that the artist exists.
	if _, err := h.queries.GetArtist(r.Context(), artistID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "artist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get artist",
		})
		return
	}

	stats, err := h.queries.GetArtistRatingStats(r.Context(), artistID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get artist rating stats",
		})
		return
	}

	average, err := stats.AverageRating.Float64Value()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to convert average rating",
		})
		return
	}

	writeJSON(w, http.StatusOK, ratingStatsResponse{
		AverageRating: average.Float64,
		RatingsCount:  stats.RatingsCount,
	})
}

func (h *Handler) getReleaseRatingStats(w http.ResponseWriter, r *http.Request) {
	releaseID, ok := parseRatingID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	// Check that the release exists.
	if _, err := h.queries.GetRelease(r.Context(), releaseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "release not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release",
		})
		return
	}

	stats, err := h.queries.GetReleaseRatingStats(r.Context(), releaseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release rating stats",
		})
		return
	}

	average, err := stats.AverageRating.Float64Value()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to convert average rating",
		})
		return
	}

	writeJSON(w, http.StatusOK, ratingStatsResponse{
		AverageRating: average.Float64,
		RatingsCount:  stats.RatingsCount,
	})
}
