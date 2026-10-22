package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	queries := sqldb.New(db)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})

	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "not_ready"})
			return
		}

		writeJSON(w, http.StatusOK, healthResponse{Status: "ready"})
	})

	r.Post("/api/v1/artists", func(w http.ResponseWriter, r *http.Request) {
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

		if strings.TrimSpace(req.Name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "name is required",
			})
			return
		}

		normalizedName := strings.ToLower(strings.TrimSpace(req.NormalizedName))
		if normalizedName == "" {
			normalizedName = strings.ToLower(strings.TrimSpace(req.Name))
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

		artist, err := queries.CreateArtist(r.Context(), sqldb.CreateArtistParams{
			Name:           strings.TrimSpace(req.Name),
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
	})

	r.Get("/api/v1/artists/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid artist id",
			})
			return
		}

		artist, err := queries.GetArtist(r.Context(), id)
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
	})

	r.Get("/api/v1/artists", func(w http.ResponseWriter, r *http.Request) {
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

		artists, err := queries.SearchArtists(r.Context(), sqldb.SearchArtistsParams{
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

		writeJSON(w, http.StatusOK, artists)
	})

	r.Put("/api/v1/artists/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		artist, err := queries.UpdateArtist(r.Context(), sqldb.UpdateArtistParams{
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
	})

	return r
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
