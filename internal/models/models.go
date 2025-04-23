package models

import (
	"log/slog"

	"github.com/go-chi/chi"
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

type ServerDeps struct {
	Logger *slog.Logger
	Addr   string
	Router chi.Router
}
