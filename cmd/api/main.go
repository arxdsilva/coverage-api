package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arxdsilva/opencoverage/internal/adapters/auth"
	githubadapter "github.com/arxdsilva/opencoverage/internal/adapters/github"
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
	gitHubInsightsRepo := postgres.NewGitHubOrgInsightsRepository(pool)
	listGitHubReviewersUC := application.NewListGitHubReviewersLeaderboardUseCase(gitHubInsightsRepo, clockAdapter)
	listGitHubHangingPRsUC := application.NewListGitHubHangingPullRequestsUseCase(gitHubInsightsRepo, clockAdapter)
	gitHubInsightsService := githubadapter.NewOrgInsightsService(
		cfg.GitHubAPIBaseURL,
		cfg.GitHubToken,
		cfg.GitHubInsightsMaxRepos,
		time.Duration(cfg.GitHubInsightsCacheTTLSeconds)*time.Second,
	)
	syncGitHubInsightsUC := application.NewSyncGitHubOrgInsightsUseCase(
		gitHubInsightsService,
		gitHubInsightsRepo,
		txManager,
		idGenerator,
		clockAdapter,
	)

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
		listGitHubReviewersUC,
		listGitHubHangingPRsUC,
	)
	router := httpadapter.NewRouter(handler, authenticator, cfg.APIKeyHeader)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()

	errCh := make(chan error, 2)
	go func() {
		slog.Info("server_starting", "addr", cfg.ServerAddr)
		errCh <- server.ListenAndServe()
	}()
	go runGitHubInsightsWorker(runCtx, syncGitHubInsightsUC, cfg.GitHubOrgs, cfg.GitHubInsightsWindowDays, cfg.GitHubInsightsSyncInterval)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	shutdownRequested := false
	select {
	case sig := <-sigCh:
		slog.Info("shutdown_signal_received", "signal", sig.String())
		shutdownRequested = true
		cancelRun()
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server_failed", "error", err)
			cancelRun()
			os.Exit(1)
		}
		shutdownRequested = true
		cancelRun()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if shutdownRequested {
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown_failed", "error", err)
		}
	}
	slog.Info("server_stopped")
}

func runGitHubInsightsWorker(
	ctx context.Context,
	syncUC *application.SyncGitHubOrgInsightsUseCase,
	orgs []string,
	windowDays []int,
	interval time.Duration,
) {
	runOnce := func(ctx context.Context) {
		for _, org := range orgs {
			if err := syncUC.Execute(ctx, application.SyncGitHubOrgInsightsInput{Org: org, WindowDays: windowDays}); err != nil {
				if errors.Is(err, context.Canceled) {
					slog.Info("github_insights_sync_canceled", "org", org)
					return
				}
				slog.Error("github_insights_sync_failed", "org", org, "error", err)
				continue
			}
			slog.Info("github_insights_sync_completed", "org", org, "window_days", windowDays)
		}
	}

	slog.Info("github_insights_worker_started", "orgs_count", len(orgs), "window_days", windowDays, "interval", interval.String())
	runOnce(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("github_insights_worker_stopped")
			return
		case <-ticker.C:
			runOnce(ctx)
		}
	}
}
