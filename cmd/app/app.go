package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/morzik45/stk-registry/pkg/config"
	"github.com/morzik45/stk-registry/pkg/email/receiver"
	"github.com/morzik45/stk-registry/pkg/logging"
	"github.com/morzik45/stk-registry/pkg/postgres"
	"github.com/morzik45/stk-registry/pkg/scheduler"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	router                *gin.Engine
	db                    *postgres.DB
	cfg                   *config.Config
	logger                *zap.Logger
	emailReceiver         *receiver.Receiver
	emailCheckerScheduler *scheduler.ScheduledExecutor
	emailSenderScheduler  *scheduler.ScheduledExecutor
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	var err error

	app := &App{
		router: gin.Default(),
		cfg:    cfg,
	}

	app.logger, err = logging.NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	defer func(logger *zap.Logger) {
		syncErr := logger.Sync()
		if syncErr != nil {
			fmt.Println("Error while syncing logger:", err)
		}
	}(app.logger)

	app.db, err = postgres.NewDB(ctx, app.cfg, app.logger)
	if err != nil {
		return nil, err
	}

	app.emailReceiver, err = receiver.NewReceiver(app.db, app.cfg, app.logger)
	if err != nil {
		return nil, err
	}

	app.initFrontend()
	app.initBackend()

	return app, nil
}

func (app *App) Run(ctx context.Context) {

	app.RunPeriodicTasks()

	// Запускаем приложение.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.cfg.Web.LocalPort),
		Handler: app.router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("listen and serve http", zap.Error(err))
		}
	}()

	// корректный выход из приложения
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	app.logger.Info("Server Started")
	<-done
	app.logger.Info("Server Stopped")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		app.Close(ctx)
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		app.logger.Fatal("Server Shutdown Failed", zap.Error(err))
	}
	app.logger.Info("Server Exited Properly")
}

func (app *App) Close(ctx context.Context) {
	if app.emailCheckerScheduler != nil {
		app.emailCheckerScheduler.Stop()
	}
	if app.emailSenderScheduler != nil {
		app.emailSenderScheduler.Stop()
	}
	err := app.db.Close()
	if err != nil {
		app.logger.Error("failed to close postgres client", zap.Error(err))
	}
}

func (app *App) emailChecker() {

}
