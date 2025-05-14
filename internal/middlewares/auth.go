package middlewares

import (
	"net/http"
)

func (m *Mw) Auth() chiMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if err := m.authService.VerifyToken(token); err != nil {
				m.handleJSONError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
