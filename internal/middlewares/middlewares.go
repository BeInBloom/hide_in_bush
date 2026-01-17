package middlewares

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/BeInBloom/hide_in_bush/internal/models"
)

type contextKey string

// UserIDKey используется для извлечения userID из контекста запроса
const UserIDKey contextKey = "userID"

type chiMiddleware = func(next http.Handler) http.Handler

type (
	authService interface {
		VerifyToken(token string) error
		ParseToken(token string) (string, error)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errResponse := models.ErrorResponse{
		Status: "error",
		Errors: message,
	}
	json.NewEncoder(w).Encode(errResponse)
}

// GetUserIDFromContext извлекает userID из контекста запроса.
// Возвращает пустую строку, если userID не найден.
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
