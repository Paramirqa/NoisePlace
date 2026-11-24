package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	maxPlaylistTitleLength       = 100
	maxPlaylistDescriptionLength = 5000
)

type createPlaylistRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

func (h *Handler) registerPlaylistRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Post("/api/v1/me/playlists", h.createPlaylist)
		r.Get("/api/v1/me/playlists", h.listPlaylists)
		r.Get("/api/v1/me/playlists/{id}", h.getPlaylist)
		r.Patch("/api/v1/me/playlists/{id}", h.updatePlaylist)
		r.Delete("/api/v1/me/playlists/{id}", h.deletePlaylist)
		r.Get("/api/v1/me/playlists/{id}/releases", h.listPlaylistReleases)
		r.Post("/api/v1/me/playlists/{id}/releases", h.addPlaylistRelease)
		r.Delete("/api/v1/me/playlists/{id}/releases/{release_id}", h.removePlaylistRelease)
	})
}

func parsePlaylistID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid playlist id",
		})
		return 0, false
	}

	return id, true
}

func decodePlaylistJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
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

func validatePlaylistTitle(title string) (string, bool) {
	title = strings.TrimSpace(title)

	if title == "" || utf8.RuneCountInString(title) > maxPlaylistTitleLength {
		return "", false
	}

	return title, true
}

func validatePlaylistDescription(description *string) bool {
	return description == nil ||
		utf8.RuneCountInString(*description) <= maxPlaylistDescriptionLength
}

func (h *Handler) createPlaylist(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	var req createPlaylistRequest
	if err := decodePlaylistJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	title, ok := validatePlaylistTitle(req.Title)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "title must contain between 1 and 100 characters",
		})
		return
	}

	if !validatePlaylistDescription(req.Description) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "description must not exceed 5000 characters",
		})
		return
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		req.Description = &description
	}

	playlist, err := h.queries.CreatePlaylist(r.Context(), sqldb.CreatePlaylistParams{
		UserID:      userID,
		Title:       title,
		Description: nullableText(req.Description),
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create playlist",
		})
		return
	}

	writeJSON(w, http.StatusCreated, playlist)
}

func (h *Handler) listPlaylists(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlists, err := h.queries.ListUserPlaylists(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list playlists",
		})
		return
	}

	if playlists == nil {
		playlists = []sqldb.Playlist{}
	}

	writeJSON(w, http.StatusOK, playlists)
}

func (h *Handler) getPlaylist(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	playlist, err := h.queries.GetUserPlaylist(r.Context(), sqldb.GetUserPlaylistParams{
		ID:     playlistID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "playlist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get playlist",
		})
		return
	}

	writeJSON(w, http.StatusOK, playlist)
}

func (h *Handler) updatePlaylist(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	var fields map[string]json.RawMessage
	if err := decodePlaylistJSON(w, r, &fields); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	if len(fields) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "at least one field must be provided",
		})
		return
	}

	for key := range fields {
		if key != "title" && key != "description" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "only title and description can be updated",
			})
			return
		}
	}

	var titleValue *string
	if raw, exists := fields["title"]; exists {
		var title string
		if err := json.Unmarshal(raw, &title); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "title must be a non-null string",
			})
			return
		}

		normalized, valid := validatePlaylistTitle(title)
		if !valid {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "title must contain between 1 and 100 characters",
			})
			return
		}

		titleValue = &normalized
	}

	var descriptionValue *string
	updateDescription := false

	if raw, exists := fields["description"]; exists {
		updateDescription = true

		if string(raw) != "null" {
			var description string
			if err := json.Unmarshal(raw, &description); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "description must be a string or null",
				})
				return
			}

			if !validatePlaylistDescription(&description) {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "description must not exceed 5000 characters",
				})
				return
			}

			description = strings.TrimSpace(description)
			descriptionValue = &description
		}
	}

	playlist, err := h.queries.UpdatePlaylist(r.Context(), sqldb.UpdatePlaylistParams{
		ID:                playlistID,
		UserID:            userID,
		Title:             nullableText(titleValue),
		Description:       nullableText(descriptionValue),
		UpdateDescription: updateDescription,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "playlist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update playlist",
		})
		return
	}

	writeJSON(w, http.StatusOK, playlist)
}

func (h *Handler) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	rows, err := h.queries.DeletePlaylist(r.Context(), sqldb.DeletePlaylistParams{
		ID:     playlistID,
		UserID: userID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete playlist",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "playlist not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type addPlaylistReleaseRequest struct {
	ReleaseID int64 `json:"release_id"`
}

// GET /api/v1/me/playlists/{id}/releases
func (h *Handler) listPlaylistReleases(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	// Проверяем существование плейлиста и его владельца.
	_, err := h.queries.GetUserPlaylist(r.Context(), sqldb.GetUserPlaylistParams{
		ID:     playlistID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "playlist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get playlist",
		})
		return
	}

	releases, err := h.queries.ListPlaylistReleases(
		r.Context(),
		sqldb.ListPlaylistReleasesParams{
			ID:     playlistID,
			UserID: userID,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list playlist releases",
		})
		return
	}

	if releases == nil {
		releases = []sqldb.ListPlaylistReleasesRow{}
	}

	writeJSON(w, http.StatusOK, releases)
}

// POST /api/v1/me/playlists/{id}/releases
func (h *Handler) addPlaylistRelease(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	var req addPlaylistReleaseRequest
	if err := decodePlaylistJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON request body",
		})
		return
	}

	if req.ReleaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "release_id must be a positive integer",
		})
		return
	}

	ctx := r.Context()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to begin transaction",
		})
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := h.queries.WithTx(tx)

	// Блокируем плейлист для последовательного изменения состава.
	_, err = queries.GetPlaylistForUpdate(ctx, sqldb.GetPlaylistForUpdateParams{
		ID:     playlistID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "playlist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to lock playlist",
		})
		return
	}

	// Проверяем существование релиза.
	_, err = queries.GetRelease(ctx, req.ReleaseID)
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

	playlistRelease, err := queries.AddReleaseToPlaylist(
		ctx,
		sqldb.AddReleaseToPlaylistParams{
			PlaylistID: playlistID,
			ReleaseID:  req.ReleaseID,
			UserID:     userID,
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "release is already in playlist",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to add release to playlist",
		})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to commit transaction",
		})
		return
	}

	writeJSON(w, http.StatusCreated, playlistRelease)
}

// DELETE /api/v1/me/playlists/{id}/releases/{release_id}
func (h *Handler) removePlaylistRelease(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	playlistID, ok := parsePlaylistID(w, r)
	if !ok {
		return
	}

	releaseID, err := strconv.ParseInt(chi.URLParam(r, "release_id"), 10, 64)
	if err != nil || releaseID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid release id",
		})
		return
	}

	ctx := r.Context()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to begin transaction",
		})
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := h.queries.WithTx(tx)

	// Блокируем плейлист перед изменением состава.
	_, err = queries.GetPlaylistForUpdate(ctx, sqldb.GetPlaylistForUpdateParams{
		ID:     playlistID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "playlist not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to lock playlist",
		})
		return
	}

	rows, err := queries.RemoveReleaseFromPlaylist(
		ctx,
		sqldb.RemoveReleaseFromPlaylistParams{
			PlaylistID: playlistID,
			ReleaseID:  releaseID,
			UserID:     userID,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to remove release from playlist",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "release is not in playlist",
		})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to commit transaction",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
