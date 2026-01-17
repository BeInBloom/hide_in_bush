package mainapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type appRunner struct {
	index    int
	instance app
	logger   *slog.Logger

	running atomic.Bool
	closed  atomic.Bool
}

func newAppRunner(index int, instance app, logger *slog.Logger) *appRunner {
	return &appRunner{
		//не знаю пока, зачем храню индекс, он используется только в 1 месте, но пусть будет
		index:    index,
		instance: instance,
		logger:   logger,
	}
}

func (r *appRunner) Run(ctx context.Context, errCh chan<- error, cancelFunc context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	//Имеет ли смысл проверять на то, что приложение уже было закрыто?
	//Просто если убрать эту проверку, то я смогу запускать его, когда потребуется
	if r.running.Load() || r.closed.Load() {
		r.logger.Warn("Attempt to run already running or closed application")
		return
	}

	r.running.Store(true)

	r.logger.Debug("Starting application")

	appErrCh := make(chan error, 1)
	go func() {
		err := r.instance.Run()
		appErrCh <- err
	}()

	select {
	case err := <-appErrCh:
		r.running.Store(false)

		if err != nil {
			r.logger.Error("Application failed",
				"error", err)
			errCh <- fmt.Errorf("app %d failed: %w", r.index, err)
			cancelFunc()
		} else {
			r.logger.Info("Application completed successfully")
		}
	case <-ctx.Done():
		r.running.Store(false)
		r.handleShutdown()
	}
}

func (r *appRunner) handleShutdown() {
	if r.closed.Load() {
		return
	}

	r.logger.Warn("Shutting down application due to context cancellation")

	if err := r.Close(); err != nil {
		r.logger.Error("Error closing application during shutdown", "error", err)
	}
}

func (r *appRunner) IsRunning() bool {
	return r.running.Load()
}

func (r *appRunner) IsClosed() bool {
	return r.closed.Load()
}

func (r *appRunner) Close() error {
	if !r.closed.CompareAndSwap(false, true) {
		r.logger.Debug("Application already closed")
		return nil
	}

	r.logger.Debug("Closing application")

	err := r.instance.Close()
	if err != nil {
		r.logger.Error("Error closing application",
			"error", err)
	} else {
		r.logger.Debug("Application closed successfully")
	}

	return err
}
