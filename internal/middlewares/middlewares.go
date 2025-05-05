package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/BeInBloom/hide_in_bush/internal/models"
)

type chiMiddleware = func(next http.Handler) http.Handler

type (
	authService interface {
		VerifyToken(token string) error
	}
)

type Mw struct {
	authService authService
	logger      *slog.Logger
}

func New(lg *slog.Logger, as authService) *Mw {
	return &Mw{
		authService: as,
		logger:      lg,
	}
}

func (m *Mw) handleJSONError(w http.ResponseWriter, status int, message ...string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	errResponse := models.ErrorResponse{
		Status: "error",
		Errors: message,
	}
	json.NewEncoder(w).Encode(errResponse)
}
