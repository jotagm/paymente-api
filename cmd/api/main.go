package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/jotagm/payment-api/internal/handler"
	"github.com/jotagm/payment-api/internal/middleware"
	"github.com/jotagm/payment-api/internal/repository"
	"github.com/jotagm/payment-api/internal/service"
	"github.com/jotagm/payment-api/observability"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); otelEndpoint != "" {
		tp, err := observability.NewTracerProvider(ctx, otelEndpoint, "payment-api")
		if err != nil {
			slog.Error("tracer provider", "error", err)
			os.Exit(1)
		}
		defer tp.Shutdown(ctx)
	} else {
		slog.Info("OTEL_EXPORTER_OTLP_ENDPOINT not set, tracing disabled")
	}

	db, err := sqlx.Connect("postgres", mustEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("database connect", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	accountRepo := repository.NewAccountRepository(db)
	transferRepo := repository.NewTransferRepository(db)

	transferSvc := service.NewTransferService(accountRepo, transferRepo)
	accountSvc := service.NewAccountService(accountRepo)

	jwtSecret := mustEnv("JWT_SECRET")

	authHandler := handler.NewAuthHandler(jwtSecret)
	transferHandler := handler.NewTransferHandler(transferSvc)
	accountHandler := handler.NewAccountHandler(accountSvc)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(observability.MetricsMiddleware)

	r.Handle("/metrics", promhttp.Handler())

	r.Post("/auth/token", authHandler.Token)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))
		r.Post("/transfers", transferHandler.Create)
		r.Get("/transfers/{id}", transferHandler.GetByID)
		r.Get("/accounts/{id}/balance", accountHandler.GetBalance)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	slog.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, otelhttp.NewHandler(r, "payment-api")); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *sqlx.DB) error {
	migration := `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";

		CREATE TABLE IF NOT EXISTS accounts (
			id         TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
			owner_name TEXT          NOT NULL,
			balance    NUMERIC(15,2) NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS transfers (
			id           TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
			from_account TEXT          NOT NULL REFERENCES accounts(id),
			to_account   TEXT          NOT NULL REFERENCES accounts(id),
			amount       NUMERIC(15,2) NOT NULL,
			status       TEXT          NOT NULL DEFAULT 'completed',
			created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
		);

		INSERT INTO accounts (id, owner_name, balance) VALUES
			('acc-001', 'Alice',   10000.00),
			('acc-002', 'Bob',      5000.00),
			('acc-003', 'Charlie',  2500.00)
		ON CONFLICT DO NOTHING;
	`
	_, err := db.Exec(migration)
	if err != nil {
		return err
	}
	slog.Info("migrations applied")
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
