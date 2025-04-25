package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	orderservice "github.com/BeInBloom/hide_in_bush/internal/services/order_service"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	jsonvalidator "github.com/BeInBloom/hide_in_bush/internal/validator/json_validator"
	"github.com/shopspring/decimal"
)

type validator interface {
	Validate(data []byte) (bool, error)
	Report() []string
}

var _ validator = (*jsonvalidator.Validator)(nil)

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

type Handlers struct {
	userService       userService
	authService       authService
	orderService      orderService
	withdrawalService withdrawalService
}

func New(
	userService userService,
	authService authService,
	orderService orderService,
	withdrawalService withdrawalService,
) *Handlers {
	return &Handlers{
		userService:       userService,
		authService:       authService,
		orderService:      orderService,
		withdrawalService: withdrawalService,
	}
}

func (h *Handlers) RegisterUserHandler() http.HandlerFunc {
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
	validator := jsonvalidator.New(userCredentialsSchema)

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

		if ok, err := validator.Validate(body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		} else if !ok {
			errors := validator.Report()

			errResponse := models.RegisterResponse{
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

			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
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

func (h *Handlers) LoginUserHandler() http.HandlerFunc {
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
	validator := jsonvalidator.New(userCredentialsSchema)

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

		if ok, err := validator.Validate(body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		} else if !ok {
			errors := validator.Report()

			errResponse := models.RegisterResponse{
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

func (h *Handlers) UploadOrderHandler() http.HandlerFunc {
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

func (h *Handlers) GetUserOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.Header.Get("Authorization")

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

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(orders)
	}
}

func (h *Handlers) GetUserBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Get("Authorization")

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

func (h *Handlers) WithdrawPointsHandler() http.HandlerFunc {
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
	validator := jsonvalidator.New(withdrawalSchema)
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

		if ok, err := validator.Validate(body); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		} else if !ok {
			errors := validator.Report()

			errResponse := models.RegisterResponse{
				Status: "error",
				Errors: errors,
			}

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errResponse)
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
			Order:       withdrawalRequest.Order,
			Sum:         decimal.NewFromInt(withdrawalRequest.Sum),
			ProcessedAt: time.Now(),
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

func (h *Handlers) GetWithdrawalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.Header.Get("Authorization")

		userID, err := h.authService.ParseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		withdrawals, err := h.withdrawalService.GetUserWithdrawals(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			errResponse := models.WithdrawalsPointsResponse{
				Status: "error",
				Errors: []string{"internal server error"},
			}

			json.NewEncoder(w).Encode(errResponse)
			return
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(withdrawals)
	}
}
