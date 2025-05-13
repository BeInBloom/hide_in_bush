package di

import (
	"log/slog"

	"github.com/BeInBloom/hide_in_bush/internal/handlers"
	"github.com/BeInBloom/hide_in_bush/internal/logger"
	"github.com/BeInBloom/hide_in_bush/internal/middlewares"
	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/router"
	authservice "github.com/BeInBloom/hide_in_bush/internal/services/auth_service"
	orderservice "github.com/BeInBloom/hide_in_bush/internal/services/order_service"
	userservice "github.com/BeInBloom/hide_in_bush/internal/services/user_service"
	withdrawalservice "github.com/BeInBloom/hide_in_bush/internal/services/withdrawal_service"
	psqlstorage "github.com/BeInBloom/hide_in_bush/internal/storage/psql"
	"github.com/go-chi/chi"
)

type routerBuilder interface {
	Build() chi.Router
}

type container struct {
	router            routerBuilder
	cfg               models.Config
	lg                *slog.Logger
	handlers          *handlers.Handlers
	middlewares       *middlewares.Mw
	userService       *userservice.UserService
	orderService      *orderservice.OrderService
	authService       *authservice.AuthService
	withdrawalService *withdrawalservice.WithdrawalService
	db                *psqlstorage.PqsqlStorage
}

func New(cfg models.Config) *container {
	return &container{
		cfg: cfg,
	}
}

func (c *container) DB() *psqlstorage.PqsqlStorage {
	if c.db == nil {
		c.db = psqlstorage.New(
			c.Config().Server.DSN,
		)
	}

	return c.db
}

func (c *container) UserService() *userservice.UserService {
	if c.userService == nil {
		c.userService = userservice.New(
			c.DB(),
		)
	}

	return c.userService
}

func (c *container) OrderService() *orderservice.OrderService {
	if c.orderService == nil {
		c.orderService = orderservice.New(
			c.DB(),
		)
	}

	return c.orderService
}

func (c *container) AuthService() *authservice.AuthService {
	if c.authService == nil {
		c.authService = authservice.New()
	}

	return c.authService
}

func (c *container) WithdrawalService() *withdrawalservice.WithdrawalService {
	if c.withdrawalService == nil {
		c.withdrawalService = withdrawalservice.New(
			c.Config().Server.Address,
			c.DB(),
		)
	}

	return c.withdrawalService
}

func (c *container) Handlers() *handlers.Handlers {
	if c.handlers == nil {
		c.handlers = handlers.New(
			c.UserService(),
			c.AuthService(),
			c.OrderService(),
			c.WithdrawalService(),
		)
	}

	return c.handlers
}

func (c *container) Middleware() *middlewares.Mw {
	if c.middlewares == nil {
		c.middlewares = middlewares.New(
			c.Logger(),
			c.AuthService(),
		)
	}

	return c.middlewares
}

func (c *container) Router() chi.Router {
	if c.router == nil {
		c.router = router.New(
			c.Handlers(),
			c.Middleware(),
		)
	}

	return c.router.Build()
}

func (c *container) Env() string {
	return c.cfg.Env
}

func (c *container) Config() models.Config {
	return c.cfg
}

func (c *container) Address() string {
	return c.cfg.Server.Address
}

func (c *container) Logger() *slog.Logger {
	if c.lg == nil {
		c.lg = logger.New(c.cfg.Env)
	}

	return c.lg
}
