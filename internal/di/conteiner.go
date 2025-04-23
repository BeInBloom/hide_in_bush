package di

import (
	"log/slog"

	"github.com/BeInBloom/hide_in_bush/internal/logger"
	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/go-chi/chi"
)

type routerBuilder interface {
	Build() chi.Router
}

type container struct {
	cfg    models.Config
	lg     *slog.Logger
	router routerBuilder
}

func New(cfg models.Config) *container {
	return &container{
		cfg: cfg,
	}
}

func (c *container) Router() chi.Router {
	if c.router == nil {

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
