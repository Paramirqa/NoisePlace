package http

import (
	"errors"
	"net/http"
	"strconv"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) registerLibraryRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Post("/api/v1/me/library/artists/{artistID}", h.addArtistToLibrary)
		r.Delete("/api/v1/me/library/artists/{artistID}", h.removeArtistFromLibrary)

		r.Post("/api/v1/me/library/releases/{releaseID}", h.addReleaseToLibrary)
		r.Delete("/api/v1/me/library/releases/{releaseID}", h.removeReleaseFromLibrary)
		r.Get("/api/v1/me/library/artists", h.listLibraryArtists)
		r.Get("/api/v1/me/library/releases", h.listLibraryReleases)
	})
}

func (h *Handler) addArtistToLibrary(w http.ResponseWriter, r *http.Request) {
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

	err = h.queries.AddArtistToLibrary(r.Context(), sqldb.AddArtistToLibraryParams{
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to add artist to library",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeArtistFromLibrary(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.queries.RemoveArtistFromLibrary(r.Context(), sqldb.RemoveArtistFromLibraryParams{
		UserID:   userID,
		ArtistID: artistID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to remove artist from library",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "artist is not in library",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addReleaseToLibrary(w http.ResponseWriter, r *http.Request) {
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

	err = h.queries.AddReleaseToLibrary(r.Context(), sqldb.AddReleaseToLibraryParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to add release to library",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeReleaseFromLibrary(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.queries.RemoveReleaseFromLibrary(r.Context(), sqldb.RemoveReleaseFromLibraryParams{
		UserID:    userID,
		ReleaseID: releaseID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to remove release from library",
		})
		return
	}

	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "release is not in library",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseLibraryID(
	w http.ResponseWriter,
	r *http.Request,
	paramName string,
	entityName string,
) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, paramName), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid " + entityName + " id",
		})
		return 0, false
	}

	return id, true
}

func (h *Handler) listLibraryArtists(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	artists, err := h.queries.ListLibraryArtists(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list library artists",
		})
		return
	}

	if artists == nil {
		artists = []sqldb.ListLibraryArtistsRow{}
	}

	writeJSON(w, http.StatusOK, artists)
}

func (h *Handler) listLibraryReleases(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
		return
	}

	releases, err := h.queries.ListLibraryReleases(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list library releases",
		})
		return
	}

	if releases == nil {
		releases = []sqldb.ListLibraryReleasesRow{}
	}

	writeJSON(w, http.StatusOK, releases)
}
