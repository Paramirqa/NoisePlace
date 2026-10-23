package http

import (
	"net/http"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db      *pgxpool.Pool
	queries *sqldb.Queries
}

func NewRouter(db *pgxpool.Pool) http.Handler {
	h := &Handler{
		db:      db,
		queries: sqldb.New(db),
	}

	r := chi.NewRouter()

	h.registerHealthRoutes(r)
	h.registerArtistRoutes(r)

	return r
}
