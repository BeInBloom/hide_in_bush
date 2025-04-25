package di

import (
	"log/slog"
	"net/http"

	"github.com/BeInBloom/hide_in_bush/internal/handlers"
	"github.com/BeInBloom/hide_in_bush/internal/logger"
	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/router"
	"github.com/go-chi/chi"
)

type routerBuilder interface {
	Build() chi.Router
}

type chiMiddleware = func(next http.Handler) http.Handler

// Хз что с этой портянкой делать
type (
	handlerBuilder interface {
		RegisterUserHandler() http.HandlerFunc
		LoginUserHandler() http.HandlerFunc
		UploadOrderHandler() http.HandlerFunc
		GetUserOrdersHandler() http.HandlerFunc
		GetUserBalanceHandler() http.HandlerFunc
		WithdrawPointsHandler() http.HandlerFunc
		GetWithdrawalsHandler() http.HandlerFunc
	}

	middlewaresBuilder interface {
		Auth() chiMiddleware
		Logger() chiMiddleware
	}

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

type container struct {
	router      routerBuilder
	cfg         models.Config
	lg          *slog.Logger
	handlers    *handlers.Handlers
	middlewares middlewaresBuilder
}

func New(cfg models.Config) *container {
	return &container{
		cfg: cfg,
	}
}

func (c *container) Handlers() *handlers.Handlers {
	if c.handlers == nil {
		panic("implement me")
		// c.handlers = handlers.New(
		// 	c.UserService(),
		// 	c.AuthService(),
		// 	c.OrderService(),
		// 	c.WithdrawService(),
		// )
	}

	return c.handlers
}

func (c *container) Middleware() middlewaresBuilder {
	if c.middlewares == nil {
		panic("implement me")
	}

	return c.middlewares
}

func (c *container) Router() chi.Router {
	if c.router == nil {
		c.router = router.New(c.Handlers(), c.Middleware())
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
		logger.New(c.cfg.Env)
	}

	return c.lg
}
