package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	jsonvalidator "github.com/BeInBloom/hide_in_bush/internal/validator/json_validator"
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
		UploadOrder(order models.Order) error
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
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.handleJSON(w, http.StatusBadRequest, "body error")
			return
		}
		defer r.Body.Close()

		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			h.handleJSON(w, http.StatusBadRequest, "bad request: invalid JSON format")
			return
		}

		id, err := h.userService.Register(credentials)
		if err != nil {
			if errors.Is(err, storage.ErrUserAlreadyExists) {
				h.handleJSON(w, http.StatusConflict, "user already exists")
				return
			}

			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		token, err := h.authService.GenerateToken(id)
		if err != nil {
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("Authorization", token)

		w.WriteHeader(http.StatusOK)

		successResponse := models.RegisterResponse{
			Status: "success",
			Token:  token,
		}

		json.NewEncoder(w).Encode(successResponse)
	}
}

func (h *Handlers) LoginUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.handleJSON(w, http.StatusBadRequest, "body error")
			return
		}
		defer r.Body.Close()

		var credentials models.UserCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			h.handleJSON(w, http.StatusBadRequest, "bad request: invalid JSON format")
			return
		}

		userID, err := h.userService.ValidateCredentials(credentials)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				h.handleJSON(w, http.StatusNotFound, "user not found")
				return
			}

			h.handleJSON(w, http.StatusBadRequest, "invalid credentials")
			return
		}

		token, err := h.authService.GenerateToken(userID)
		if err != nil {
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("Authorization", token)

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
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.handleJSON(w, http.StatusBadRequest, "body error")
			return
		}
		defer r.Body.Close()

		orderString := strings.TrimSpace(string(body))

		orderModel := models.Order{
			ID:       orderString,
			UserID:   userID,
			Status:   "NEW",
			Uploaded: time.Now(),
		}

		err = h.orderService.UploadOrder(orderModel)
		if err != nil {
			if errors.Is(err, storage.ErrOrderAlreadyRegistered) {
				h.handleJSON(w, http.StatusOK, "order already exists")
				return
			}

			if errors.Is(err, storage.ErrOrderRegisteredToOtherUser) {
				h.handleJSON(w, http.StatusConflict, "order belongs to another user")
				return
			}

			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
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
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		orders, err := h.orderService.GetUserOrders(userID)
		if err != nil {
			if errors.Is(err, storage.ErrNoOrders) {
				h.handleJSON(w, http.StatusNoContent, "no orders found")
				return
			}

			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if len(orders) == 0 {
			h.handleJSON(w, http.StatusNoContent, "no orders found")
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(orders)
	}
}

func (h *Handlers) GetUserBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token := r.Header.Get("Authorization")

		userID, err := h.authService.ParseToken(token)
		if err != nil {
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		balance, err := h.userService.UserBalance(userID)
		if err != nil {
			// Чисто теоретически, это возможно, но не должно происходить
			if errors.Is(err, storage.ErrUserNotFound) {
				h.handleJSON(w, http.StatusNotFound, "user not found")
				return
			}

			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(balance)
	}
}

func (h *Handlers) WithdrawPointsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.handleJSON(w, http.StatusBadRequest, "body error")
			return
		}
		defer r.Body.Close()

		var withdrawalRequest models.WithdrawalRequest
		if err := json.Unmarshal(body, &withdrawalRequest); err != nil {
			h.handleJSON(w, http.StatusBadRequest, "bad request: invalid JSON format")
			return
		}

		wd := models.Withdrawal{
			Order:       withdrawalRequest.Order,
			Sum:         withdrawalRequest.Sum.InexactFloat64(),
			ProcessedAt: time.Now(),
		}
		if err := h.withdrawalService.PostWithdraw(wd); err != nil {
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
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
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
			return
		}

		withdrawals, err := h.withdrawalService.GetUserWithdrawals(userID)
		if err != nil {
			h.handleJSON(w, http.StatusInternalServerError, "internal server error")
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

func (h *Handlers) handleJSON(w http.ResponseWriter, status int, errors ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResponse := models.ErrorResponse{
		Status: "error",
		Errors: errors,
	}

	json.NewEncoder(w).Encode(errResponse)
}
