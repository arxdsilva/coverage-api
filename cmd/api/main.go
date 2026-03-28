package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arxdsilva/coverage-api/internal/adapters/auth"
	httpadapter "github.com/arxdsilva/coverage-api/internal/adapters/http"
	"github.com/arxdsilva/coverage-api/internal/adapters/postgres"
	"github.com/arxdsilva/coverage-api/internal/application"
	"github.com/arxdsilva/coverage-api/internal/platform/clock"
	"github.com/arxdsilva/coverage-api/internal/platform/config"
	"github.com/arxdsilva/coverage-api/internal/platform/idgen"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	projectRepo := postgres.NewProjectRepository(pool)
	runRepo := postgres.NewCoverageRunRepository(pool)
	packageRepo := postgres.NewPackageCoverageRepository(pool)
	txManager := postgres.NewTxManager(pool)
	authenticator := auth.NewEnvAPIKeyAuthenticator(cfg.APIKeySecret)

	clockAdapter := clock.NewSystemClock()
	idGenerator := idgen.NewUUIDGenerator()

	ingestUC := application.NewIngestCoverageRunUseCase(projectRepo, runRepo, packageRepo, txManager, idGenerator, clockAdapter)
	getProjectUC := application.NewGetProjectUseCase(projectRepo)
	listRunsUC := application.NewListCoverageRunsUseCase(runRepo)
	latestComparisonUC := application.NewGetLatestComparisonUseCase(projectRepo, runRepo, packageRepo)

	handler := httpadapter.NewHandler(ingestUC, getProjectUC, listRunsUC, latestComparisonUC)
	router := httpadapter.NewRouter(handler, authenticator, cfg.APIKeyHeader)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("coverage-api listening on %s", cfg.ServerAddr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
