package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const maxNoteLength = 5000

type noteRequest struct {
	Content string `json:"content"`
}

func (h *Handler) registerNotesRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Post("/api/v1/me/notes/artists/{artistID}", h.createArtistNote)
		r.Get("/api/v1/me/notes/artists/{artistID}", h.getArtistNote)
		r.Put("/api/v1/me/notes/artists/{artistID}", h.updateArtistNote)
		r.Delete("/api/v1/me/notes/artists/{artistID}", h.deleteArtistNote)

		r.Post("/api/v1/me/notes/releases/{releaseID}", h.createReleaseNote)
		r.Get("/api/v1/me/notes/releases/{releaseID}", h.getReleaseNote)
		r.Put("/api/v1/me/notes/releases/{releaseID}", h.updateReleaseNote)
		r.Delete("/api/v1/me/notes/releases/{releaseID}", h.deleteReleaseNote)
	})
}

func decodeNoteContent(w http.ResponseWriter, r *http.Request) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req noteRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return "", false
	}

	// Reject additional JSON values after the first object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request must contain a single JSON object",
		})
		return "", false
	}

	content := strings.TrimSpace(req.Content)
	if content == "" || utf8.RuneCountInString(content) > maxNoteLength {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "content must contain between 1 and 5000 characters",
		})
		return "", false
	}

	if strings.IndexByte(content, 0) >= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "content contains an invalid character",
		})
		return "", false
	}

	return content, true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (h *Handler) createArtistNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	artistID, ok := parseLibraryID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	content, ok := decodeNoteContent(w, r)
	if !ok {
		return
	}

	_, err := h.queries.GetArtist(r.Context(), artistID)
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

	err = h.queries.CreateArtistNote(r.Context(), sqldb.CreateArtistNoteParams{
		UserID:   userID,
		ArtistID: artistID,
		Content:  content,
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "note already exists",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create artist note",
		})
		return
	}

	note, err := h.queries.GetArtistNote(r.Context(), sqldb.GetArtistNoteParams{
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get created artist note",
		})
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

func (h *Handler) getArtistNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	artistID, ok := parseLibraryID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	note, err := h.queries.GetArtistNote(r.Context(), sqldb.GetArtistNoteParams{
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "artist note not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get artist note",
		})
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) updateArtistNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	artistID, ok := parseLibraryID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	content, ok := decodeNoteContent(w, r)
	if !ok {
		return
	}

	rows, err := h.queries.UpdateArtistNote(r.Context(), sqldb.UpdateArtistNoteParams{
		Content:  content,
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update artist note",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "artist note not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteArtistNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	artistID, ok := parseLibraryID(w, r, "artistID", "artist")
	if !ok {
		return
	}

	rows, err := h.queries.DeleteArtistNote(r.Context(), sqldb.DeleteArtistNoteParams{
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete artist note",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "artist note not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createReleaseNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	releaseID, ok := parseLibraryID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	content, ok := decodeNoteContent(w, r)
	if !ok {
		return
	}

	_, err := h.queries.GetRelease(r.Context(), releaseID)
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

	err = h.queries.CreateReleaseNote(r.Context(), sqldb.CreateReleaseNoteParams{
		UserID:    userID,
		ReleaseID: releaseID,
		Content:   content,
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "note already exists",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create release note",
		})
		return
	}

	note, err := h.queries.GetReleaseNote(r.Context(), sqldb.GetReleaseNoteParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get created release note",
		})
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

func (h *Handler) getReleaseNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	releaseID, ok := parseLibraryID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	note, err := h.queries.GetReleaseNote(r.Context(), sqldb.GetReleaseNoteParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "release note not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get release note",
		})
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) updateReleaseNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	releaseID, ok := parseLibraryID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	content, ok := decodeNoteContent(w, r)
	if !ok {
		return
	}

	rows, err := h.queries.UpdateReleaseNote(r.Context(), sqldb.UpdateReleaseNoteParams{
		Content:   content,
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update release note",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "release note not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteReleaseNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	releaseID, ok := parseLibraryID(w, r, "releaseID", "release")
	if !ok {
		return
	}

	rows, err := h.queries.DeleteReleaseNote(r.Context(), sqldb.DeleteReleaseNoteParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete release note",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "release note not found",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
