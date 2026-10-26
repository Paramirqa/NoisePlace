package http

import (
	"net/http"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db        *pgxpool.Pool
	queries   *sqldb.Queries
	jwtSecret string
}

func NewRouter(db *pgxpool.Pool, jwtSecret string) http.Handler {
	h := &Handler{
		db:        db,
		queries:   sqldb.New(db),
		jwtSecret: jwtSecret,
	}

	r := chi.NewRouter()

	h.registerHealthRoutes(r)
	h.registerAuthRoutes(r)
	h.registerArtistRoutes(r)
	h.registerReleaseRoutes(r)

	return r
}
