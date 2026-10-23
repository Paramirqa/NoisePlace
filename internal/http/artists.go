package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) registerArtistRoutes(r chi.Router) {
	r.Post("/api/v1/artists", h.createArtist)
	r.Get("/api/v1/artists/{id}", h.getArtist)
	r.Get("/api/v1/artists", h.searchArtists)
	r.Put("/api/v1/artists/{id}", h.updateArtist)
}

func (h *Handler) createArtist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string  `json:"name"`
		NormalizedName string  `json:"normalized_name"`
		SortName       *string `json:"sort_name"`
		Description    *string `json:"description"`
		ImageURL       *string `json:"image_url"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
		return
	}

	normalizedName := strings.ToLower(strings.TrimSpace(req.NormalizedName))
	if normalizedName == "" {
		normalizedName = strings.ToLower(name)
	}

	if req.SortName != nil {
		value := strings.TrimSpace(*req.SortName)
		req.SortName = &value
	}

	if req.Description != nil {
		value := strings.TrimSpace(*req.Description)
		req.Description = &value
	}

	if req.ImageURL != nil {
		value := strings.TrimSpace(*req.ImageURL)
		req.ImageURL = &value
	}

	artist, err := h.queries.CreateArtist(r.Context(), sqldb.CreateArtistParams{
		Name:           name,
		NormalizedName: normalizedName,
		SortName:       nullableText(req.SortName),
		Description:    nullableText(req.Description),
		ImageUrl:       nullableText(req.ImageURL),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create artist",
		})
		return
	}

	writeJSON(w, http.StatusCreated, artist)
}

func (h *Handler) getArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist id",
		})
		return
	}

	artist, err := h.queries.GetArtist(r.Context(), id)
	if err != nil {
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

	writeJSON(w, http.StatusOK, artist)
}

func (h *Handler) searchArtists(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "query parameter is required",
		})
		return
	}

	limit := int32(20)
	offset := int32(0)

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 1 || parsed > 100 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "limit must be between 1 and 100",
			})
			return
		}

		limit = int32(parsed)
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "offset must be a non-negative integer",
			})
			return
		}

		offset = int32(parsed)
	}

	artists, err := h.queries.SearchArtists(r.Context(), sqldb.SearchArtistsParams{
		Query:      query,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to search artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.Artist{}
	}

	writeJSON(w, http.StatusOK, artists)
}

func (h *Handler) updateArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist id",
		})
		return
	}

	var req struct {
		Name           string  `json:"name"`
		NormalizedName string  `json:"normalized_name"`
		SortName       *string `json:"sort_name"`
		Description    *string `json:"description"`
		ImageURL       *string `json:"image_url"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
		return
	}

	normalizedName := strings.ToLower(strings.TrimSpace(req.NormalizedName))
	if normalizedName == "" {
		normalizedName = strings.ToLower(name)
	}

	if req.SortName != nil {
		value := strings.TrimSpace(*req.SortName)
		req.SortName = &value
	}

	if req.Description != nil {
		value := strings.TrimSpace(*req.Description)
		req.Description = &value
	}

	if req.ImageURL != nil {
		value := strings.TrimSpace(*req.ImageURL)
		req.ImageURL = &value
	}

	artist, err := h.queries.UpdateArtist(r.Context(), sqldb.UpdateArtistParams{
		ID:             id,
		Name:           name,
		NormalizedName: normalizedName,
		SortName:       nullableText(req.SortName),
		Description:    nullableText(req.Description),
		ImageUrl:       nullableText(req.ImageURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "artist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update artist",
		})
		return
	}

	writeJSON(w, http.StatusOK, artist)
}
