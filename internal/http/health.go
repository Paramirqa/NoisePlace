package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) registerHealthRoutes(r chi.Router) {
	r.Get("/healthz", h.healthz)
	r.Get("/readyz", h.readyz)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
	})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status: "not_ready",
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ready",
	})
}
