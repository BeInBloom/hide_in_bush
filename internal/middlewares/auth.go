package middlewares

import (
	"context"
	"net/http"
)

// Auth middleware проверяет JWT токен и добавляет userID в контекст запроса.
// После успешной авторизации userID можно получить через GetUserIDFromContext(ctx).
func (m *Mw) Auth() chiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if err := m.authService.VerifyToken(token); err != nil {
				m.handleJSONError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// Извлекаем userID из токена и добавляем в контекст
			userID, err := m.authService.ParseToken(token)
			if err != nil {
				m.handleJSONError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
