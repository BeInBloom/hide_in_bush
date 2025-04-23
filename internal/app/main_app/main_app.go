package mainapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type app interface {
	Run() error
	Close() error
}

// Тут это, конечно, не совсем main app
// Это больше сущность для группировки аппов, если нужно будет
type mainApp struct {
	runners []*appRunner
	logger  *slog.Logger
}

func New(logger *slog.Logger, apps ...app) *mainApp {
	runners := make([]*appRunner, len(apps))
	for i, app := range apps {
		runners[i] = newAppRunner(i, app, logger.With("app_index", i))
	}

	return &mainApp{
		runners: runners,
		logger:  logger.With("app_name", "main_app", "app_index", "main"),
	}
}

func (m *mainApp) Run() error {
	m.logger.Info("Starting all applications")

	errCh := make(chan error, len(m.runners))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for _, runner := range m.runners {
		wg.Add(1)
		go runner.Run(ctx, errCh, cancel, &wg)
	}

	go func() {
		wg.Wait()
		m.logger.Debug("All application goroutines completed")
		close(errCh)
	}()

	for err := range errCh {
		m.logger.Error("Returning application error", "error", err)
		return err
	}

	m.logger.Info("All applications started successfully")
	return nil
}

func (m *mainApp) Close() error {
	m.logger.Info("Closing all applications")

	var errs error

	for _, runner := range m.runners {
		if err := runner.Close(); err != nil {
			errs = errors.Join(
				errs, fmt.Errorf("app %d close error: %w", runner.index, err))
			m.logger.Error("Error closing application",
				"app_index", runner.index,
				"error", err)
		}
	}

	if errs != nil {
		m.logger.Error("Close completed with errors", "error", errs)
	} else {
		m.logger.Info("All applications closed successfully")
	}

	return errs
}
