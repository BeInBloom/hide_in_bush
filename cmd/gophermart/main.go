package main

import (
	"os"
	"os/signal"
	"syscall"

	mainapp "github.com/BeInBloom/hide_in_bush/internal/app/main_app"
	"github.com/BeInBloom/hide_in_bush/internal/app/server"
	"github.com/BeInBloom/hide_in_bush/internal/config"
	"github.com/BeInBloom/hide_in_bush/internal/logger"
)

func main() {
	cfg := config.MustConfig()
	lg := logger.New(cfg.Env)

	mainApp := mainapp.New(
		lg,
		server.New(cfg.Address),
	)

	errChn := make(chan error, 1)
	go func() {
		err := mainApp.Run()
		errChn <- err
	}()

	sysCalls := make(chan os.Signal, 1)
	signal.Notify(sysCalls, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChn:
		lg.Error("Application error", "error", err)
	case <-sysCalls:
		lg.Info("Received system signal, shutting down")
	}

	if err := mainApp.Close(); err != nil {
		lg.Error("Application close error", "error", err)
		os.Exit(1)
	}

	lg.Info("Application closed")
}
