package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"github.com/xeipuuv/gojsonschema"
)

type userService interface {
	Register(user models.UserCredentials) (string, error)
	Login(user models.UserCredentials) (string, error)
}

type handlers struct {
	userService
}

// newHandlers creates a new instance of handlers
func NewHandlers() *handlers {
	return &handlers{}
}

func (h *handlers) RegisterUserHandler() http.HandlerFunc {
	const userCredentialsSchema = `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["login", "password"],
		"additionalProperties": false,
		"properties": {
			"login": {
				"type": "string",
				"minLength": 3,
				"maxLength": 50,
				"pattern": "^[a-zA-Z0-9_-]+$"
			},
			"password": {
				"type": "string",
				"minLength": 4,
				"maxLength": 100
			}
		}
	}`
	schemaLoader := gojsonschema.NewStringLoader(userCredentialsSchema)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		documentLoader := gojsonschema.NewBytesLoader(body)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusBadRequest)
			return
		}

		if !result.Valid() {
			var errors []string
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			errorResponse := struct {
				Errors []string `json:"errors"`
			}{errors}

			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		token, err := h.userService.Register(credentials)
		if err != nil {
			if err == storage.ErrUserAlreadyExists {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"token":  token,
		},
		)
	}
}

func (h *handlers) LoginUserHandler() http.HandlerFunc {
	const userCredentialsSchema = `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["login", "password"],
		"additionalProperties": false,
		"properties": {
			"login": {
				"type": "string",
				"minLength": 3,
				"maxLength": 50,
				"pattern": "^[a-zA-Z0-9_-]+$"
			},
			"password": {
				"type": "string",
				"minLength": 4,
				"maxLength": 100
			}
		}
	}`
	schemaLoader := gojsonschema.NewStringLoader(userCredentialsSchema)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Валидируем данные с помощью JSON Schema
		documentLoader := gojsonschema.NewBytesLoader(body)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Проверяем результат валидации
		if !result.Valid() {
			// Собираем ошибки валидации
			var errors []string
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": errors,
			})
			return
		}

		// Десериализуем данные пользователя
		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			http.Error(w, "Bad request: invalid JSON format", http.StatusBadRequest)
			return
		}

		// Вызываем сервис для аутентификации
		token, err := h.userService.Login(credentials)
		if err != nil {
			// Проверяем тип ошибки для точного определения статуса
			if err == storage.ErrInvalidCredentials {
				http.Error(w, "Invalid login/password", http.StatusUnauthorized)
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Возвращаем токен в ответе
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}

// uploadOrderHandler returns a handler for order upload
func (h *handlers) UploadOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}

// getUserOrdersHandler returns a handler for fetching user orders
func (h *handlers) GetUserOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}

// getUserBalanceHandler returns a handler for fetching user balance
func (h *handlers) GetUserBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}

// withdrawPointsHandler returns a handler for withdrawing points
func (h *handlers) WithdrawPointsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}

// getWithdrawalsHandler returns a handler for fetching withdrawal history
func (h *handlers) GetWithdrawalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}
