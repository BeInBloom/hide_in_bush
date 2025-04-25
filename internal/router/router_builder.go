package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	timeout       = 30 * time.Second
	throttleLimit = 100
)

type chiMiddleware = func(next http.Handler) http.Handler

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
)

type routerBuilder struct {
	handlers    handlerBuilder
	middlewares middlewaresBuilder
	router      chi.Router
	once        sync.Once
}

func New(handlers handlerBuilder, middlewares middlewaresBuilder) *routerBuilder {
	return &routerBuilder{
		handlers:    handlers,
		middlewares: middlewares,
		router:      chi.NewRouter(),
	}
}

func (rb *routerBuilder) Build() chi.Router {
	rb.once.Do(func() {
		rb.setMiddlewares()
		rb.setRoutes()
	})
	return rb.router
}

func (rb *routerBuilder) setRoutes() {
	rb.router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", rb.handlers.RegisterUserHandler())
			r.Post("/login", rb.handlers.LoginUserHandler())

			r.Group(func(r chi.Router) {
				r.Use(rb.middlewares.Auth())

				r.Post("/orders", rb.handlers.UploadOrderHandler())
				r.Get("/orders", rb.handlers.GetUserOrdersHandler())

				r.Get("/balance", rb.handlers.GetUserBalanceHandler())
				r.Post("/balance/withdraw", rb.handlers.WithdrawPointsHandler())

				r.Get("/withdrawals", rb.handlers.GetWithdrawalsHandler())
			})
		})
	})
}

func (rb *routerBuilder) setMiddlewares() {
	compressor := middleware.NewCompressor(5, "gzip", "application/json")

	rb.router.Use(middleware.Recoverer)

	rb.router.Use(middleware.RequestID)
	rb.router.Use(middleware.RealIP)
	rb.router.Use(rb.middlewares.Logger())

	rb.router.Use(middleware.Throttle(throttleLimit))
	rb.router.Use(middleware.Timeout(timeout))

	rb.router.Use(compressor.Handler)
}
