package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var validReleaseTypes = map[string]bool{
	"album":       true,
	"ep":          true,
	"single":      true,
	"compilation": true,
	"mixtape":     true,
	"live":        true,
	"other":       true,
}

var validArtistRoles = map[string]bool{
	"primary":  true,
	"featured": true,
	"producer": true,
}

type releaseRequest struct {
	Title       string               `json:"title"`
	ReleaseType string               `json:"release_type"`
	ReleaseDate *string              `json:"release_date"`
	Description *string              `json:"description"`
	CoverURL    *string              `json:"cover_url"`
	Artists     []releaseArtistInput `json:"artists"`
}

type releaseArtistInput struct {
	ArtistID int64  `json:"artist_id"`
	Role     string `json:"role"`
	Position *int32 `json:"position"`
}

type releaseWithArtists struct {
	sqldb.Release
	Artists []sqldb.GetReleaseArtistsRow `json:"artists"`
}

func (h *Handler) registerReleaseRoutes(r chi.Router) {
	// Releases
	r.Post("/api/v1/releases", h.createRelease)
	r.Get("/api/v1/releases", h.searchReleases)
	r.Get("/api/v1/releases/{id}", h.getRelease)
	r.Put("/api/v1/releases/{id}", h.updateRelease)
	r.Delete("/api/v1/releases/{id}", h.deleteRelease)

	// Release artists
	r.Get("/api/v1/releases/{id}/artists", h.listReleaseArtists)
	r.Post("/api/v1/releases/{id}/artists", h.addReleaseArtist)
	r.Put("/api/v1/releases/{id}/artists/{artist_id}", h.updateReleaseArtist)
	r.Delete("/api/v1/releases/{id}/artists/{artist_id}", h.removeReleaseArtist)
}

func (h *Handler) createRelease(w http.ResponseWriter, r *http.Request) {
	var req releaseRequest

	if err := decodeReleaseRequest(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	if err := validateReleaseRequest(&req, true); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	releaseDate, err := parseReleaseDate(req.ReleaseDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "release_date must use YYYY-MM-DD format",
		})
		return
	}

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to begin transaction",
		})
		return
	}
	defer func() {
		_ = tx.Rollback(r.Context())
	}()

	queries := h.queries.WithTx(tx)

	release, err := queries.CreateRelease(
		r.Context(),
		sqldb.CreateReleaseParams{
			Title:       strings.TrimSpace(req.Title),
			ReleaseType: req.ReleaseType,
			ReleaseDate: releaseDate,
			Description: nullableText(req.Description),
			CoverUrl:    nullableText(req.CoverURL),
		},
	)
	if err != nil {
		writeDatabaseError(w, err, "failed to create release")
		return
	}

	for i, artist := range req.Artists {
		role := artist.Role
		if role == "" {
			role = "primary"
		}

		position := int32(i)
		if artist.Position != nil {
			position = *artist.Position
		}

		err = queries.AddReleaseArtist(
			r.Context(),
			sqldb.AddReleaseArtistParams{
				ReleaseID: release.ID,
				ArtistID:  artist.ArtistID,
				Role:      role,
				Position:  position,
			},
		)
		if err != nil {
			writeDatabaseError(w, err, "failed to add release artist")
			return
		}
	}

	artists, err := queries.GetReleaseArtists(r.Context(), release.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.GetReleaseArtistsRow{}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to commit transaction",
		})
		return
	}

	writeJSON(w, http.StatusCreated, releaseWithArtists{
		Release: release,
		Artists: artists,
	})
}

func (h *Handler) getRelease(w http.ResponseWriter, r *http.Request) {
	id, ok := parseReleaseID(w, r)
	if !ok {
		return
	}

	release, err := h.queries.GetRelease(r.Context(), id)
	if err != nil {
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

	response, err := h.loadReleaseWithArtists(r.Context(), release)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release artists",
		})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) searchReleases(w http.ResponseWriter, r *http.Request) {
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

	releases, err := h.queries.SearchReleases(
		r.Context(),
		sqldb.SearchReleasesParams{
			Query:      query,
			PageLimit:  limit,
			PageOffset: offset,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to search releases",
		})
		return
	}

	if releases == nil {
		releases = []sqldb.Release{}
	}

	writeJSON(w, http.StatusOK, releases)
}

func (h *Handler) updateRelease(w http.ResponseWriter, r *http.Request) {
	id, ok := parseReleaseID(w, r)
	if !ok {
		return
	}

	var req releaseRequest

	if err := decodeReleaseRequest(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	if err := validateReleaseRequest(&req, false); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	releaseDate, err := parseReleaseDate(req.ReleaseDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "release_date must use YYYY-MM-DD format",
		})
		return
	}

	release, err := h.queries.UpdateRelease(
		r.Context(),
		sqldb.UpdateReleaseParams{
			ID:          id,
			Title:       strings.TrimSpace(req.Title),
			ReleaseType: req.ReleaseType,
			ReleaseDate: releaseDate,
			Description: nullableText(req.Description),
			CoverUrl:    nullableText(req.CoverURL),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "release not found",
			})
			return
		}

		writeDatabaseError(w, err, "failed to update release")
		return
	}

	response, err := h.loadReleaseWithArtists(r.Context(), release)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release artists",
		})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) loadReleaseWithArtists(
	ctx context.Context,
	release sqldb.Release,
) (releaseWithArtists, error) {
	artists, err := h.queries.GetReleaseArtists(ctx, release.ID)
	if err != nil {
		return releaseWithArtists{}, err
	}

	if artists == nil {
		artists = []sqldb.GetReleaseArtistsRow{}
	}

	return releaseWithArtists{
		Release: release,
		Artists: artists,
	}, nil
}

func parseReleaseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return 0, false
	}

	return id, true
}

func decodeReleaseRequest(
	w http.ResponseWriter,
	r *http.Request,
	req *releaseRequest,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(req); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return err
		}

		return errors.New("request body must contain exactly one JSON value")
	}

	return nil
}

func validateReleaseRequest(req *releaseRequest, creating bool) error {
	req.Title = strings.TrimSpace(req.Title)
	req.ReleaseType = strings.TrimSpace(req.ReleaseType)

	if req.Title == "" {
		return errors.New("title is required")
	}

	if !validReleaseTypes[req.ReleaseType] {
		return errors.New(
			"release_type must be one of: album, ep, single, compilation, mixtape, live, other",
		)
	}

	if creating {
		if len(req.Artists) == 0 {
			return errors.New("at least one artist is required")
		}

		hasPrimary := false

		for i := range req.Artists {
			artist := &req.Artists[i]

			if artist.ArtistID <= 0 {
				return errors.New("artist_id must be a positive integer")
			}

			if artist.Role == "" {
				artist.Role = "primary"
			}

			if !validArtistRoles[artist.Role] {
				return errors.New(
					"artist role must be primary, featured, or producer",
				)
			}

			if artist.Position != nil && *artist.Position < 0 {
				return errors.New("artist position must be non-negative")
			}

			if artist.Role == "primary" {
				hasPrimary = true
			}
		}

		if !hasPrimary {
			return errors.New("at least one primary artist is required")
		}
	}

	normalizeOptionalField(&req.Description)
	normalizeOptionalField(&req.CoverURL)

	return nil
}

func normalizeOptionalField(value **string) {
	if *value == nil {
		return
	}

	trimmed := strings.TrimSpace(**value)
	*value = &trimmed
}

func parseReleaseDate(value *string) (pgtype.Date, error) {
	if value == nil {
		return pgtype.Date{}, nil
	}

	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
	if err != nil {
		return pgtype.Date{}, err
	}

	return pgtype.Date{
		Time:  parsed,
		Valid: true,
	}, nil
}

func writeDatabaseError(w http.ResponseWriter, err error, message string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "referenced artist or release does not exist",
			})
			return
		case "23505": // unique_violation
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "conflicting record already exists",
			})
			return
		case "23514": // check_violation
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "request violates a database constraint",
			})
			return
		}
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": message,
	})
}

// GET /api/v1/releases/{id}/artists
func (h *Handler) listReleaseArtists(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || releaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return
	}

	// Проверяем существование релиза.
	_, err = h.queries.GetRelease(r.Context(), releaseID)
	if err != nil {
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

	artists, err := h.queries.GetReleaseArtists(r.Context(), releaseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.GetReleaseArtistsRow{}
	}

	writeJSON(w, http.StatusOK, artists)
}

// POST /api/v1/releases/{id}/artists
func (h *Handler) addReleaseArtist(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || releaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return
	}

	var req releaseArtistInput

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Проверяем, что в теле содержится только один JSON-объект.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain a single JSON object",
		})
		return
	}

	if req.ArtistID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "artist_id must be positive",
		})
		return
	}

	// По умолчанию исполнитель считается основным.
	if req.Role == "" {
		req.Role = "primary"
	}

	if !validArtistRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist role",
		})
		return
	}

	position := int32(0)
	if req.Position != nil {
		position = *req.Position
	}

	if position < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "position must be non-negative",
		})
		return
	}

	// Проверяем существование релиза.
	_, err = h.queries.GetRelease(r.Context(), releaseID)
	if err != nil {
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

	// Проверяем существование исполнителя.
	_, err = h.queries.GetArtist(r.Context(), req.ArtistID)
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

	// Создаём связь между релизом и исполнителем.
	err = h.queries.AddReleaseArtist(
		r.Context(),
		sqldb.AddReleaseArtistParams{
			ReleaseID: releaseID,
			ArtistID:  req.ArtistID,
			Role:      req.Role,
			Position:  position,
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				writeJSON(w, http.StatusConflict, map[string]string{
					"error": "artist is already linked to this release",
				})
				return
			case "23503":
				writeJSON(w, http.StatusNotFound, map[string]string{
					"error": "release or artist not found",
				})
				return
			case "23514":
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid release artist data",
				})
				return
			}
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to add release artist",
		})
		return
	}

	// Возвращаем актуальный список исполнителей релиза.
	artists, err := h.queries.GetReleaseArtists(r.Context(), releaseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "artist linked, but failed to load release artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.GetReleaseArtistsRow{}
	}

	writeJSON(w, http.StatusCreated, artists)
}

// PUT /api/v1/releases/{id}/artists/{artist_id}
func (h *Handler) updateReleaseArtist(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || releaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return
	}

	artistID, err := strconv.ParseInt(chi.URLParam(r, "artist_id"), 10, 64)
	if err != nil || artistID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist id",
		})
		return
	}

	var req releaseArtistInput

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain a single JSON object",
		})
		return
	}

	if req.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "role is required",
		})
		return
	}

	if !validArtistRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist role",
		})
		return
	}

	if req.Position == nil || *req.Position < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "position must be a non-negative integer",
		})
		return
	}

	// Проверяем существование релиза.
	_, err = h.queries.GetRelease(r.Context(), releaseID)
	if err != nil {
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

	// Проверяем существование исполнителя.
	_, err = h.queries.GetArtist(r.Context(), artistID)
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

	// Обновляем связь исполнителя с релизом.
	rowsAffected, err := h.queries.UpdateReleaseArtist(
		r.Context(),
		sqldb.UpdateReleaseArtistParams{
			Role:      req.Role,
			Position:  *req.Position,
			ReleaseID: releaseID,
			ArtistID:  artistID,
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23514":
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid release artist data",
				})
				return
			}
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update release artist",
		})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "artist is not linked to this release",
		})
		return
	}

	// Возвращаем актуальный список исполнителей релиза.
	artists, err := h.queries.GetReleaseArtists(r.Context(), releaseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "artist updated, but failed to load release artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.GetReleaseArtistsRow{}
	}

	writeJSON(w, http.StatusOK, artists)
}

// DELETE /api/v1/releases/{id}/artists/{artist_id}
func (h *Handler) removeReleaseArtist(w http.ResponseWriter, r *http.Request) {
	releaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || releaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return
	}

	artistID, err := strconv.ParseInt(chi.URLParam(r, "artist_id"), 10, 64)
	if err != nil || artistID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid artist id",
		})
		return
	}

	// Проверяем существование релиза.
	_, err = h.queries.GetRelease(r.Context(), releaseID)
	if err != nil {
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

	// Проверяем существование исполнителя.
	_, err = h.queries.GetArtist(r.Context(), artistID)
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

	// Удаляем связь и получаем количество затронутых строк.
	rowsAffected, err := h.queries.RemoveReleaseArtist(
		r.Context(),
		sqldb.RemoveReleaseArtistParams{
			ReleaseID: releaseID,
			ArtistID:  artistID,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to remove release artist",
		})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "artist is not linked to this release",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/releases/{id}
func (h *Handler) deleteRelease(w http.ResponseWriter, r *http.Request) {
	releaseID, ok := parseReleaseID(w, r)
	if !ok {
		return
	}

	rowsAffected, err := h.queries.DeleteRelease(r.Context(), releaseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete release",
		})
		return
	}

	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "release not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
