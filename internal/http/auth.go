package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"music-platform/internal/auth"
	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	emailPattern    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func (h *Handler) registerAuthRoutes(r chi.Router) {

	r.Post("/api/v1/auth/register", h.registerUser)

}

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
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

	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain a single JSON object",
		})
		return

	}

	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !usernamePattern.MatchString(username) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "username must contain 3-32 letters, digits or underscores",
		})
		return
	}

	if len(email) > 254 || !emailPattern.MatchString(email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid email",
		})
		return
	}

	if !utf8.ValidString(req.Password) ||
		len(req.Password) < 12 ||
		len(req.Password) > 128 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "password must be between 12 and 128 bytes",
		})
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to process password",
		})
		return
	}

	user, err := h.queries.CreateUser(r.Context(), sqldb.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "username or email already exists",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to register user",
		})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

var _ = pgx.ErrNoRows
