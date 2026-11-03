package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) registerRatingRoutes(r chi.Router) {
	// Public aggregate rating endpoints.
	r.Get("/api/v1/artists/{artistID}/rating", h.getArtistRatingStats)
	r.Get("/api/v1/releases/{releaseID}/rating", h.getReleaseRatingStats)

	// Personal rating endpoints.
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Put("/api/v1/me/ratings/artists/{artistID}", h.upsertArtistRating)
		r.Delete("/api/v1/me/ratings/artists/{artistID}", h.deleteArtistRating)
		r.Get("/api/v1/me/ratings/artists/{artistID}", h.getArtistRating)

		r.Put("/api/v1/me/ratings/releases/{releaseID}", h.upsertReleaseRating)
		r.Delete("/api/v1/me/ratings/releases/{releaseID}", h.deleteReleaseRating)
		r.Get("/api/v1/me/ratings/releases/{releaseID}", h.getReleaseRating)
	})
}

type ratingRequest struct {
	Rating int16 `json:"rating"`
}

func parseRatingID(w http.ResponseWriter, r *http.Request, paramName, entityName string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, paramName), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid " + entityName + " id",
		})
		return 0, false
	}

	return id, true
}

func parseRatingRequest(w http.ResponseWriter, r *http.Request) (int16, bool) {
	var req ratingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return 0, false
	}

	if req.Rating < 1 || req.Rating > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "rating must be between 1 and 10",
		})
		return 0, false
	}

	return req.Rating, true
}

func (h *Handler) upsertArtistRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	artistID, ok := parseRatingID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	rating, ok := parseRatingRequest(w, r)
	if !ok {
		return
	}

	if _, err := h.queries.GetArtist(r.Context(), artistID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "artist not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get artist"})
		return
	}

	result, err := h.queries.UpsertArtistRating(r.Context(), sqldb.UpsertArtistRatingParams{
		UserID: userID, ArtistID: artistID, Rating: rating,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save artist rating"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteArtistRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	artistID, ok := parseRatingID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	rows, err := h.queries.DeleteArtistRating(r.Context(), sqldb.DeleteArtistRatingParams{
		UserID: userID, ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete artist rating"})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "artist rating not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getArtistRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	artistID, ok := parseRatingID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	result, err := h.queries.GetArtistRating(r.Context(), sqldb.GetArtistRatingParams{
		UserID: userID, ArtistID: artistID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "artist rating not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get artist rating"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) upsertReleaseRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	releaseID, ok := parseRatingID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	rating, ok := parseRatingRequest(w, r)
	if !ok {
		return
	}

	if _, err := h.queries.GetRelease(r.Context(), releaseID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "release not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get release"})
		return
	}

	result, err := h.queries.UpsertReleaseRating(r.Context(), sqldb.UpsertReleaseRatingParams{
		UserID: userID, ReleaseID: releaseID, Rating: rating,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save release rating"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteReleaseRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	releaseID, ok := parseRatingID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	rows, err := h.queries.DeleteReleaseRating(r.Context(), sqldb.DeleteReleaseRatingParams{
		UserID: userID, ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete release rating"})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "release rating not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getReleaseRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	releaseID, ok := parseRatingID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	result, err := h.queries.GetReleaseRating(r.Context(), sqldb.GetReleaseRatingParams{
		UserID: userID, ReleaseID: releaseID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "release rating not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get release rating"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
