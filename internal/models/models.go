package models

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi"
	"github.com/shopspring/decimal"
)

type (
	Config struct {
		Env    string `yaml:"env" json:"env" env:"ENV" env-default:"local"`
		Server Server `yaml:"server" json:"server"`
	}

	Server struct {
		Address              string `yaml:"address" json:"address" env:"ADDRESS"`
		DSN                  string `yaml:"dsn" json:"dsn" env:"DSN"`
		AccrualSystemAddress string `yaml:"accrual_system_address" json:"accrual_system_address" env:"ACCRUAL_SYSTEM_ADDRESS"`
	}
)

type (
	UserCredentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

type (
	RegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	WithdrawalRequest struct {
		Order string `json:"order"`
		Sum   int64  `json:"sum"`
	}
)

type (
	RegisterResponse struct {
		Status string   `json:"status"`
		Token  string   `json:"token,omitempty"`
		Errors []string `json:"errors,omitempty"`
	}

	LoginResponse struct {
		Status string   `json:"status"`
		Token  string   `json:"token,omitempty"`
		Errors []string `json:"errors,omitempty"`
	}

	OrdersGetResponse struct {
		Status string   `json:"status"`
		Orders []Order  `json:"orders,omitempty"`
		Errors []string `json:"errors,omitempty"`
	}

	OrdersPostResponse struct {
		Status string   `json:"status"`
		Errors []string `json:"errors,omitempty"`
	}

	UserBalanceResponse struct {
		Status  string   `json:"status"`
		Balance Balance  `json:"balance"`
		Errors  []string `json:"errors,omitempty"`
	}

	WithdrawalsPointsResponse struct {
		Status string   `json:"status"`
		Errors []string `json:"errors,omitempty"`
	}

	GetWithdrawalsResponse struct {
		Status      string       `json:"status"`
		Withdrawals []Withdrawal `json:"withdrawals,omitempty"`
		Errors      []string     `json:"errors,omitempty"`
	}

	ErrorResponse struct {
		Status string   `json:"status"`
		Errors []string `json:"errors,omitempty"`
	}
)

type (
	User struct {
		//В целом, не понятно, что там за юзер по заданию,
		//По этому пусть будет такая заглушка
		ID        string    `json:"id"`
		Login     string    `json:"login"`
		Password  string    `json:"password"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Orders    []Order   `json:"orders"`
		Balance   Balance   `json:"balance"`
	}

	Order struct {
		ID       string          `json:"number"`
		UserID   string          `json:"user_id"`
		Status   string          `json:"status"`
		Accrual  decimal.Decimal `json:"accrual,omitempty"`
		Uploaded string          `json:"uploaded_at"`
	}

	Balance struct {
		UserID         string          `json:"user_id"`
		CurrentBalance decimal.Decimal `json:"current_balance"`
		Withdrawn      decimal.Decimal `json:"withdrawn"`
	}

	Withdrawal struct {
		Order       string          `json:"order"`
		Sum         decimal.Decimal `json:"sum"`
		ProcessedAt time.Time       `json:"processed_at"`
	}
)

type (
	ServerDeps struct {
		Logger *slog.Logger
		Addr   string
		Router chi.Router
	}

	LogConfig struct {
	}
)
