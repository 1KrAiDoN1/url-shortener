package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	slogger "url-shortener/internal/lib/logger/slog"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"
	"url-shortener/internal/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()
	log := setupPrettySlog()
	cfg, err := config.MustLoadConfig("config/local.yaml")
	if err != nil {
		log.Error("Ошибка загрузки конфига", slogger.Err(err))
		os.Exit(1)
	}

	fmt.Println(cfg.DB_config_path)
	log.Info("starting url-shortener", slog.String("env", cfg.HTTPServer.Address))
	log.Debug("debug messages are enabled")

	database, err := postgres.NewDatabase(ctx, cfg.DB_config_path)
	if err != nil {
		log.Error("failed to init db", slogger.Err(err))
		os.Exit(1)
	}
	storage := storage.NewStorage(database)
	service := service.NewService(storage)
	handlers := save.NewHandlers(service)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Route("/url", func(r chi.Router) {
		r.Post("/", handlers.New(ctx, log))
		r.Delete("/{alias}", delete.New(ctx, log, storage))

	})
	router.Get("/{alias}", redirect.New(ctx, log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("Starting server", slog.String("address", cfg.Address))
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", slogger.Err(err))
		}
		close(serverErr)
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("server started")

	select {
	case err := <-serverErr:
		log.Error("server error", slogger.Err(err))
		return
	case sig := <-done:
		log.Info("Shutting down...", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("server shutdown failed", slogger.Err(err))
		}
		log.Info("Server gracefully stopped", slog.String("address", cfg.Address))
		return
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
