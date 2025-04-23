package server

import (
	"log/slog"
	"net/http"
)

type middle struct {
	logger *slog.Logger
}

func (m *middle) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Логика аутентификации
			// 1. Получить токен из заголовка Authorization
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 2. Проверить токен (примерная логика)
			// userID, err := m.userService.ValidateToken(tokenString)
			// if err != nil {
			//     http.Error(w, "Unauthorized", http.StatusUnauthorized)
			//     return
			// }

			// 3. Добавить информацию о пользователе в контекст
			// ctx := context.WithValue(r.Context(), "userID", userID)

			// 4. Передать запрос дальше
			// next.ServeHTTP(w, r.WithContext(ctx))

			// Заглушка
			m.logger.Info("Auth middleware called", "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
