package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arxdsilva/opencoverage/internal/adapters/auth"
	httpadapter "github.com/arxdsilva/opencoverage/internal/adapters/http"
	"github.com/arxdsilva/opencoverage/internal/adapters/postgres"
	"github.com/arxdsilva/opencoverage/internal/application"
	"github.com/arxdsilva/opencoverage/internal/platform/clock"
	"github.com/arxdsilva/opencoverage/internal/platform/config"
	"github.com/arxdsilva/opencoverage/internal/platform/idgen"
	"github.com/arxdsilva/opencoverage/internal/platform/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("startup_failed", "stage", "load_config", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("startup_failed", "stage", "validate_config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("startup_failed", "stage", "create_db_pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := migrations.Up(ctx, cfg.DatabaseURL, cfg.MigrationsDir); err != nil {
		slog.Error("startup_failed", "stage", "run_migrations", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("startup_failed", "stage", "ping_db", "error", err)
		os.Exit(1)
	}

	projectRepo := postgres.NewProjectRepository(pool)
	runRepo := postgres.NewCoverageRunRepository(pool)
	packageRepo := postgres.NewPackageCoverageRepository(pool)
	integrationRunRepo := postgres.NewIntegrationTestRunRepository(pool)
	integrationSpecRepo := postgres.NewIntegrationSpecResultRepository(pool)
	txManager := postgres.NewTxManager(pool)
	authenticator := auth.NewEnvAPIKeyAuthenticator(cfg.APIKeySecret)

	clockAdapter := clock.NewSystemClock()
	idGenerator := idgen.NewUUIDGenerator()

	ingestUC := application.NewIngestCoverageRunUseCase(projectRepo, runRepo, packageRepo, txManager, idGenerator, clockAdapter)
	ingestIntegrationUC := application.NewIngestIntegrationRunUseCase(projectRepo, integrationRunRepo, integrationSpecRepo, txManager, idGenerator, clockAdapter)
	listProjectsUC := application.NewListProjectsUseCase(projectRepo)
	getProjectUC := application.NewGetProjectUseCase(projectRepo)
	listRunsUC := application.NewListCoverageRunsUseCase(runRepo)
	listIntegrationRunsUC := application.NewListIntegrationRunsUseCase(integrationRunRepo)
	latestComparisonUC := application.NewGetLatestComparisonUseCase(projectRepo, runRepo, packageRepo)
	latestIntegrationComparisonUC := application.NewGetLatestIntegrationComparisonUseCase(projectRepo, integrationRunRepo, integrationSpecRepo)
	getIntegrationRunUC := application.NewGetIntegrationRunUseCase(integrationRunRepo, integrationSpecRepo)
	getIntegrationHeatmapUC := application.NewGetIntegrationHeatmapUseCase(integrationRunRepo)
	listBranchesUC := application.NewListBranchesUseCase(runRepo)
	listContributorsUC := application.NewListContributorsUseCase(projectRepo, runRepo)

	handler := httpadapter.NewHandler(
		ingestUC,
		ingestIntegrationUC,
		listProjectsUC,
		getProjectUC,
		listRunsUC,
		listIntegrationRunsUC,
		latestComparisonUC,
		latestIntegrationComparisonUC,
		getIntegrationRunUC,
		getIntegrationHeatmapUC,
		listBranchesUC,
		listContributorsUC,
	)
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
		slog.Info("server_starting", "addr", cfg.ServerAddr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("shutdown_signal_received", "signal", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server_failed", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown_failed", "error", err)
	}
	slog.Info("server_stopped")
}
