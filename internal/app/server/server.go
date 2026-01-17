package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
)

const (
	shutdownTimeout = 10 * time.Second
)

type server struct {
	httpServer *http.Server
	done       chan struct{}
	logger     *slog.Logger
}

func New(deps models.ServerDeps) *server {
	logger := deps.Logger.With(
		"app_name", "http_server", "addr", deps.Addr,
	)

	s := &http.Server{
		Addr:    deps.Addr,
		Handler: deps.Router,
	}

	server := &server{
		logger:     logger,
		httpServer: s,
		done:       make(chan struct{}),
	}

	return server
}

func (s *server) Run() error {
	s.logger.Info("Starting HTTP server")

	errCh := make(chan error, 1)
	go func() {
		err := s.httpServer.ListenAndServe()
		errCh <- err
	}()

	s.logger.Info("HTTP server started successfully")

	<-s.done

	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("HTTP server error", "error", err)
		return err
	}

	return nil
}

func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.logger.Info("Closing HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down HTTP server", "error", err)
		return err
	}

	s.logger.Info("HTTP server closed successfully")

	s.done <- struct{}{}

	return nil
}
