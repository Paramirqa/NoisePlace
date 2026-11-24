package http

import (
	"net/http"

	sqldb "music-platform/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		MaxAge: 300,
	}))

	// Health endpoints
	h.registerHealthRoutes(r)

	// Authentication endpoints
	h.registerAuthRoutes(r)

	// User endpoints
	h.registerUserRoutes(r)

	// Library endpoints
	h.registerLibraryRoutes(r)

	// Artist endpoints
	h.registerArtistRoutes(r)

	// Release endpoints
	h.registerReleaseRoutes(r)

	// Rating endpoints
	h.registerRatingRoutes(r)

	// Notes endpoints
	h.registerNotesRoutes(r)

	return r
}
