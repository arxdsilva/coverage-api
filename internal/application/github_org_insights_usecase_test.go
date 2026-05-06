package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

type stubGitHubInsightsService struct {
	leaderboardResult GitHubReviewerLeaderboard
	hangingResult     GitHubHangingPullRequests
	err               error

	lastLeaderboardQuery GitHubLeaderboardQuery
	lastHangingQuery     GitHubHangingQuery
}

func (s *stubGitHubInsightsService) GetReviewersLeaderboard(ctx context.Context, query GitHubLeaderboardQuery) (GitHubReviewerLeaderboard, error) {
	s.lastLeaderboardQuery = query
	if s.err != nil {
		return GitHubReviewerLeaderboard{}, s.err
	}
	return s.leaderboardResult, nil
}

func (s *stubGitHubInsightsService) GetHangingPullRequests(ctx context.Context, query GitHubHangingQuery) (GitHubHangingPullRequests, error) {
	s.lastHangingQuery = query
	if s.err != nil {
		return GitHubHangingPullRequests{}, s.err
	}
	return s.hangingResult, nil
}

type fakeRateLimitedError struct{}

func (fakeRateLimitedError) Error() string     { return "rate limited" }
func (fakeRateLimitedError) RateLimited() bool { return true }

type stubGitHubInsightsRepository struct {
	reviewersSnapshot GitHubReviewerSnapshot
	hangingSnapshot   GitHubHangingSnapshot
	err               error

	lastOrg      string
	lastWindow   int
	savedReviews int
	savedHanging int
}

func (s *stubGitHubInsightsRepository) SaveReviewersSnapshot(ctx context.Context, snapshot GitHubReviewerSnapshot) error {
	s.savedReviews++
	if s.err != nil {
		return s.err
	}
	s.reviewersSnapshot = snapshot
	return nil
}

func (s *stubGitHubInsightsRepository) GetLatestReviewersSnapshot(ctx context.Context, org string, windowDays int) (GitHubReviewerSnapshot, error) {
	s.lastOrg = org
	s.lastWindow = windowDays
	if s.err != nil {
		return GitHubReviewerSnapshot{}, s.err
	}
	return s.reviewersSnapshot, nil
}

func (s *stubGitHubInsightsRepository) SaveHangingSnapshot(ctx context.Context, snapshot GitHubHangingSnapshot) error {
	s.savedHanging++
	if s.err != nil {
		return s.err
	}
	s.hangingSnapshot = snapshot
	return nil
}

func (s *stubGitHubInsightsRepository) GetLatestHangingSnapshot(ctx context.Context, org string) (GitHubHangingSnapshot, error) {
	s.lastOrg = org
	if s.err != nil {
		return GitHubHangingSnapshot{}, s.err
	}
	return s.hangingSnapshot, nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fixedIDGenerator struct{}

func (fixedIDGenerator) NewID() string { return "id-1" }

func TestListGitHubReviewersLeaderboardUseCaseExecute(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	repo := &stubGitHubInsightsRepository{
		reviewersSnapshot: GitHubReviewerSnapshot{
			Org:        "acme",
			WindowFrom: now.AddDate(0, 0, -30),
			WindowTo:   now,
			Summary:    GitHubLeaderboardSummary{RepositoriesScanned: 2, PullRequestsScanned: 10, TotalReviewEvents: 20},
			Reviewers:  []GitHubReviewer{{Login: "alice", TotalReviews: 11}},
		},
	}
	uc := NewListGitHubReviewersLeaderboardUseCase(repo, fixedClock{now: now})

	out, err := uc.Execute(context.Background(), ListGitHubReviewersLeaderboardInput{Org: "acme"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Org != "acme" {
		t.Fatalf("expected org acme, got %s", out.Org)
	}
	if repo.lastWindow != 30 {
		t.Fatalf("expected requested window 30, got %d", repo.lastWindow)
	}
	if repo.lastOrg != "acme" {
		t.Fatalf("expected org acme, got %s", repo.lastOrg)
	}
}

func TestListGitHubReviewersLeaderboardUseCaseNotFound(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	repo := &stubGitHubInsightsRepository{err: domain.ErrNotFound}
	uc := NewListGitHubReviewersLeaderboardUseCase(repo, fixedClock{now: now})

	_, err := uc.Execute(context.Background(), ListGitHubReviewersLeaderboardInput{Org: "acme"})
	if err == nil {
		t.Fatalf("expected error")
	}
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Code != CodeNotFound {
		t.Fatalf("expected code %s, got %s", CodeNotFound, appErr.Code)
	}
}

func TestListGitHubHangingPullRequestsUseCaseDefaults(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	repo := &stubGitHubInsightsRepository{
		hangingSnapshot: GitHubHangingSnapshot{
			Org:         "acme",
			GeneratedAt: now,
			Summary:     GitHubHangingSummary{RepositoriesScanned: 1, OpenPRsScanned: 2, HangingPRs: 1},
			Items:       []GitHubHangingPullRequest{{Repository: "repo", Number: 1, AgeHours: 100, IdleHours: 60}},
		},
	}
	uc := NewListGitHubHangingPullRequestsUseCase(repo, fixedClock{now: now})

	out, err := uc.Execute(context.Background(), ListGitHubHangingPullRequestsInput{Org: "acme"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Criteria["minIdleHours"].(int) != 48 {
		t.Fatalf("expected default minIdleHours 48")
	}
	if out.Criteria["minOpenHours"].(int) != 72 {
		t.Fatalf("expected default minOpenHours 72")
	}
	if out.Summary.HangingPRs != 1 {
		t.Fatalf("expected filtered hanging count 1, got %d", out.Summary.HangingPRs)
	}
}

func TestListGitHubHangingPullRequestsUseCaseValidation(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	repo := &stubGitHubInsightsRepository{}
	uc := NewListGitHubHangingPullRequestsUseCase(repo, fixedClock{now: now})

	_, err := uc.Execute(context.Background(), ListGitHubHangingPullRequestsInput{Org: "", Limit: 10})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Code != CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %s", appErr.Code)
	}
}

func TestSyncGitHubOrgInsightsUseCaseExecute(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	service := &stubGitHubInsightsService{
		leaderboardResult: GitHubReviewerLeaderboard{
			Org:       "acme",
			From:      now.AddDate(0, 0, -30),
			To:        now,
			Summary:   GitHubLeaderboardSummary{RepositoriesScanned: 1, PullRequestsScanned: 3, TotalReviewEvents: 5},
			Reviewers: []GitHubReviewer{{Login: "alice", TotalReviews: 3}},
		},
		hangingResult: GitHubHangingPullRequests{
			Org:         "acme",
			GeneratedAt: now,
			Summary:     GitHubHangingSummary{RepositoriesScanned: 1, OpenPRsScanned: 2, HangingPRs: 1},
			Items:       []GitHubHangingPullRequest{{Repository: "repo", Number: 10, AgeHours: 5, IdleHours: 3}},
		},
	}
	repo := &stubGitHubInsightsRepository{}

	uc := NewSyncGitHubOrgInsightsUseCase(service, repo, stubTxManager{}, fixedIDGenerator{}, fixedClock{now: now})
	err := uc.Execute(context.Background(), SyncGitHubOrgInsightsInput{Org: "acme", WindowDays: []int{30}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.savedReviews != 1 {
		t.Fatalf("expected one reviewers snapshot save, got %d", repo.savedReviews)
	}
	if repo.savedHanging != 1 {
		t.Fatalf("expected one hanging snapshot save, got %d", repo.savedHanging)
	}
}
