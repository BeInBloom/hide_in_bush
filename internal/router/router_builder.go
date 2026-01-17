package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/validator"
	jsonvalidator "github.com/BeInBloom/hide_in_bush/internal/validator/json_validator"
	lunavallidator "github.com/BeInBloom/hide_in_bush/internal/validator/luna_vallidator"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

const (
	timeout       = 30 * time.Second
	throttleLimit = 100
)

const (
	registerScheme = `{
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

	loginScheme = `{
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

	withdrawScheme = `{
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
				"type": "number",
				"minimum": 0.01
			}
		}
	}`
)

type (
	chiMiddleware = func(next http.Handler) http.Handler
)

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
		BodyValidator(stauts int, v validator.Validator[[]byte]) chiMiddleware
	}
)

type routerBuilder struct {
	handlers    handlerBuilder
	middlewares middlewaresBuilder
	router      chi.Router
	once        sync.Once
}

func New(
	handlers handlerBuilder, middlewares middlewaresBuilder,
) *routerBuilder {
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
	var (
		loginJSONValidator     = jsonvalidator.New(loginScheme)
		registerJSONValidator  = jsonvalidator.New(registerScheme)
		withdrawJSONValidator  = jsonvalidator.New(withdrawScheme)
		algorithmLunaValidator = lunavallidator.New()
	)

	rb.router.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.With(middleware.AllowContentType("application/json"), rb.middlewares.BodyValidator(http.StatusBadRequest, registerJSONValidator)).
				Post("/register", rb.handlers.RegisterUserHandler())
			r.With(middleware.AllowContentType("application/json"), rb.middlewares.BodyValidator(http.StatusBadRequest, loginJSONValidator)).
				Post("/login", rb.handlers.LoginUserHandler())

			r.Group(func(r chi.Router) {
				r.Use(rb.middlewares.Auth())

				r.With(middleware.AllowContentType("text/plain"), rb.middlewares.BodyValidator(http.StatusUnprocessableEntity, algorithmLunaValidator)).
					Post("/orders", rb.handlers.UploadOrderHandler())
				r.Get("/orders", rb.handlers.GetUserOrdersHandler())

				r.Get("/balance", rb.handlers.GetUserBalanceHandler())
				r.With(middleware.AllowContentType("application/json"), rb.middlewares.BodyValidator(http.StatusBadRequest, withdrawJSONValidator)).
					Post("/balance/withdraw", rb.handlers.WithdrawPointsHandler())

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

	rb.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{
			"Accept", "Authorization",
			"Content-Type", "Content-Encoding",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}
