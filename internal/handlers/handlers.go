package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	orderservice "github.com/BeInBloom/hide_in_bush/internal/services/order_service"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"github.com/xeipuuv/gojsonschema"
)

type (
	userService interface {
		Register(credentials models.UserCredentials) (userID string, err error)
		ValidateCredentials(models.UserCredentials) (userID string, err error)
		UserBalance(userID string) (models.Balance, error)
	}

	withdrawalService interface {
		GetUserWithdrawals(userID string) ([]models.Withdrawal, error)
		PostWithdraw(withdrawwal models.Withdrawal) error
	}

	orderService interface {
		UploadOrder(userID string, order models.Order) error
		GetUserOrders(userID string) ([]models.Order, error)
	}

	authService interface {
		GenerateToken(userID string) (string, error)
		ParseToken(token string) (string, error)
	}
)

type handlers struct {
	userService       userService
	authService       authService
	orderService      orderService
	withdrawalService withdrawalService
}

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
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"Body error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}
		defer r.Body.Close()

		documentLoader := gojsonschema.NewBytesLoader(body)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if !result.Valid() {
			var errors []string
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}

			w.WriteHeader(http.StatusBadRequest)

			errorResponse := models.RegisterResponse{
				Status: "error",
				Errors: errors,
			}

			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"bad request: invalid JSON format"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		userID, err := h.userService.Register(credentials)
		if err != nil {
			if err == storage.ErrUserAlreadyExists {
				w.WriteHeader(http.StatusConflict)

				errResponse := models.RegisterResponse{
					Status: "error",
					Errors: []string{"user already exists"},
				}

				json.NewEncoder(w).Encode(errResponse)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		token, err := h.authService.GenerateToken(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)

		successResponse := models.RegisterResponse{
			Status: "success",
			Token:  token,
		}

		json.NewEncoder(w).Encode(successResponse)
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
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.LoginResponse{
				Status: "error",
				Errors: []string{"body error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}
		defer r.Body.Close()

		documentLoader := gojsonschema.NewBytesLoader(body)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errResponse := models.LoginResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if !result.Valid() {
			var errors []string
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}

			errResponse := models.LoginResponse{
				Status: "error",
				Errors: errors,
			}

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResponse)
			return
		}

		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.LoginResponse{
				Status: "error",
				Errors: []string{"bad request: invalid JSON format"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		userID, err := h.userService.ValidateCredentials(credentials)
		if err != nil {
			if err == storage.ErrInvalidCredentials {
				w.WriteHeader(http.StatusUnauthorized)

				errResponse := models.LoginResponse{
					Status: "error",
					Errors: []string{"invalid credentials"},
				}

				json.NewEncoder(w).Encode(errResponse)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.LoginResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		token, err := h.authService.GenerateToken(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.LoginResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)

		successResponse := models.LoginResponse{
			Status: "success",
			Token:  token,
		}

		json.NewEncoder(w).Encode(successResponse)
	}
}

func (h *handlers) UploadOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.Header.Get("Authorization")

		userID, err := h.authService.ParseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.OrdersPostResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.OrdersPostResponse{
				Status: "error",
				Errors: []string{"body error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}
		defer r.Body.Close()

		orderString := strings.TrimSpace(string(body))

		order, err := strconv.Atoi(orderString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.OrdersPostResponse{
				Status: "error",
				Errors: []string{"invalid order format"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if !isValidLuna(order) {
			w.WriteHeader(http.StatusUnprocessableEntity)

			errResponse := models.OrdersPostResponse{
				Status: "error",
				Errors: []string{"not Luna order number"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		orderModel := models.Order{
			ID:     orderString,
			Status: "NEW",
		}

		err = h.orderService.UploadOrder(userID, orderModel)
		if err != nil {
			if errors.Is(err, orderservice.ErrOrderBelongsToUser) {
				w.WriteHeader(http.StatusOK)

				successResponse := models.OrdersPostResponse{
					Status: "success",
				}

				json.NewEncoder(w).Encode(successResponse)
				return
			}

			if errors.Is(err, orderservice.ErrOrderBelongsToAnotherUser) {
				w.WriteHeader(http.StatusConflict)

				errResponse := models.OrdersPostResponse{
					Status: "error",
					Errors: []string{"order belongs to another user"},
				}

				json.NewEncoder(w).Encode(errResponse)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.OrdersPostResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *handlers) GetUserOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := w.Header().Get("Authorization")

		userID, err := h.authService.ParseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		orders, err := h.orderService.GetUserOrders(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		json.NewEncoder(w).Encode(orders)

		w.WriteHeader(http.StatusOK)
	}
}

func (h *handlers) GetUserBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.Header.Get("Authorization")

		userID, err := h.authService.ParseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.UserBalanceResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		balance, err := h.userService.UserBalance(userID)
		if err != nil {
			// Чисто теоретически, это возможно, но не должно происходить
			if errors.Is(err, storage.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)

				errResponse := models.UserBalanceResponse{
					Status: "error",
					Errors: []string{"user not found"},
				}

				json.NewEncoder(w).Encode(errResponse)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.UserBalanceResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(balance)
	}
}

func (h *handlers) WithdrawPointsHandler() http.HandlerFunc {
	const (
		withdrawalSchema = `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"required": ["order", "sum"],
			"additionalProperties": false,
			"properties": {
				"order": {
					"type": "string",
					"minLength": 1
				},
				"sum": {
					"type": "integer",
					"minimum": 1
				}
			}
		}`
	)
	schemaLoader := gojsonschema.NewStringLoader(withdrawalSchema)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"Body error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}
		defer r.Body.Close()

		documentLoader := gojsonschema.NewBytesLoader(body)
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if !result.Valid() {
			var errors []string
			for _, err := range result.Errors() {
				errors = append(errors, err.String())
			}

			w.WriteHeader(http.StatusBadRequest)

			errorResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: errors,
			}

			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		var withdrawalRequest models.WithdrawalRequest
		if err := json.Unmarshal(body, &withdrawalRequest); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"bad request: invalid JSON format"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		wd := models.Withdrawal{
			Order: withdrawalRequest.Order,
			Sum:   withdrawalRequest.Sum,
		}
		if err := h.withdrawalService.PostWithdraw(wd); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)

		successResponse := models.WithdrawalsPointsResponse{
			Status: "success",
		}

		json.NewEncoder(w).Encode(successResponse)
	}
}

func (h *handlers) GetWithdrawalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("implement me")
	}
}
