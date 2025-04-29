package main

import (
	"os"
	"os/signal"
	"syscall"

	mainapp "github.com/BeInBloom/hide_in_bush/internal/app/main_app"
	"github.com/BeInBloom/hide_in_bush/internal/app/server"
	"github.com/BeInBloom/hide_in_bush/internal/config"
	"github.com/BeInBloom/hide_in_bush/internal/di"
	"github.com/BeInBloom/hide_in_bush/internal/models"
)

// Сам проект переписывался чуть ли не с 0 неоднократно
// Подчистить и доделать времени уже особо нету
func main() {
	di := di.New(config.MustConfig())

	lg := di.Logger()

	mainApp := mainapp.New(
		di.Logger(),
		server.New(
			models.ServerDeps{
				Logger: di.Logger(),
				Addr:   di.Address(),
				Router: di.Router(),
			},
		),
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
