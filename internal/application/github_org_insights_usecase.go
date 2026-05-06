package application

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

type GitHubReviewer struct {
	Login                      string    `json:"login"`
	DisplayName                string    `json:"displayName,omitempty"`
	TotalReviews               int       `json:"totalReviews"`
	Approvals                  int       `json:"approvals"`
	ChangeRequests             int       `json:"changeRequests"`
	Comments                   int       `json:"comments"`
	UniquePullRequestsReviewed int       `json:"uniquePullRequestsReviewed"`
	RepositoriesReviewed       int       `json:"reposReviewed"`
	LatestReviewAt             time.Time `json:"latestReviewAt"`
}

type GitHubLeaderboardSummary struct {
	RepositoriesScanned int `json:"repositoriesScanned"`
	PullRequestsScanned int `json:"pullRequestsConsidered"`
	TotalReviewEvents   int `json:"totalReviewEvents"`
}

type GitHubReviewerLeaderboard struct {
	Org       string                   `json:"org"`
	From      time.Time                `json:"from"`
	To        time.Time                `json:"to"`
	Summary   GitHubLeaderboardSummary `json:"summary"`
	Reviewers []GitHubReviewer         `json:"reviewers"`
}

type GitHubLeaderboardQuery struct {
	Org          string
	From         time.Time
	To           time.Time
	Limit        int
	Repositories []string
}

type GitHubHangingPullRequest struct {
	Repository         string    `json:"repository"`
	Number             int       `json:"number"`
	Title              string    `json:"title"`
	URL                string    `json:"url"`
	Author             string    `json:"author"`
	Draft              bool      `json:"draft"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	LastActivityAt     time.Time `json:"lastActivityAt"`
	AgeHours           int       `json:"ageHours"`
	IdleHours          int       `json:"idleHours"`
	ReviewState        string    `json:"reviewState"`
	MergeableState     string    `json:"mergeableState"`
	RequestedReviewers []string  `json:"requestedReviewers"`
	Labels             []string  `json:"labels"`
	Reasons            []string  `json:"reasons"`
}

type GitHubHangingSummary struct {
	RepositoriesScanned int `json:"repositoriesScanned"`
	OpenPRsScanned      int `json:"openPullRequestsConsidered"`
	HangingPRs          int `json:"hangingPullRequests"`
}

type GitHubHangingPullRequests struct {
	Org         string                     `json:"org"`
	GeneratedAt time.Time                  `json:"generatedAt"`
	Summary     GitHubHangingSummary       `json:"summary"`
	Items       []GitHubHangingPullRequest `json:"items"`
}

type GitHubHangingQuery struct {
	Org           string
	Limit         int
	MinIdleHours  int
	MinOpenHours  int
	Repositories  []string
	Author        string
	IncludeDrafts bool
	Sort          string
}

type GitHubOrgInsightsService interface {
	GetReviewersLeaderboard(ctx context.Context, query GitHubLeaderboardQuery) (GitHubReviewerLeaderboard, error)
	GetHangingPullRequests(ctx context.Context, query GitHubHangingQuery) (GitHubHangingPullRequests, error)
}

type GitHubReviewerSnapshot struct {
	SnapshotID  string
	Org         string
	WindowDays  int
	WindowFrom  time.Time
	WindowTo    time.Time
	GeneratedAt time.Time
	Summary     GitHubLeaderboardSummary
	Reviewers   []GitHubReviewer
}

type GitHubHangingSnapshot struct {
	SnapshotID  string
	Org         string
	GeneratedAt time.Time
	Summary     GitHubHangingSummary
	Items       []GitHubHangingPullRequest
}

type ListGitHubReviewersLeaderboardInput struct {
	Org        string
	From       *time.Time
	To         *time.Time
	WindowDays int
	Limit      int
	Repos      []string
}

type ListGitHubReviewersLeaderboardOutput struct {
	Org    string `json:"org"`
	Window struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"window"`
	Summary   GitHubLeaderboardSummary `json:"summary"`
	Reviewers []GitHubReviewer         `json:"reviewers"`
}

type ListGitHubReviewersLeaderboardUseCase struct {
	repository GitHubOrgInsightsRepository
	clock      Clock
}

func NewListGitHubReviewersLeaderboardUseCase(repository GitHubOrgInsightsRepository, clock Clock) *ListGitHubReviewersLeaderboardUseCase {
	return &ListGitHubReviewersLeaderboardUseCase{repository: repository, clock: clock}
}

func (uc *ListGitHubReviewersLeaderboardUseCase) Execute(ctx context.Context, in ListGitHubReviewersLeaderboardInput) (ListGitHubReviewersLeaderboardOutput, error) {
	org := strings.TrimSpace(in.Org)
	if org == "" {
		return ListGitHubReviewersLeaderboardOutput{}, NewInvalidArgument("org is required", map[string]any{"field": "org"})
	}

	windowDays := in.WindowDays
	if windowDays == 0 {
		windowDays = 30
	}
	if windowDays < 1 || windowDays > 365 {
		return ListGitHubReviewersLeaderboardOutput{}, NewInvalidArgument("windowDays must be between 1 and 365", map[string]any{"field": "windowDays"})
	}

	limit := in.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return ListGitHubReviewersLeaderboardOutput{}, NewInvalidArgument("limit must be between 1 and 100", map[string]any{"field": "limit"})
	}

	to := uc.clock.Now().UTC()
	from := to.AddDate(0, 0, -windowDays)
	if in.From != nil && in.To != nil {
		from = in.From.UTC()
		to = in.To.UTC()
	}
	if from.After(to) {
		return ListGitHubReviewersLeaderboardOutput{}, NewInvalidArgument("from must be before or equal to to", map[string]any{"field": "from"})
	}

	result, err := uc.repository.GetLatestReviewersSnapshot(ctx, org, windowDays)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ListGitHubReviewersLeaderboardOutput{}, NewNotFound("reviewers leaderboard snapshot not found", map[string]any{"org": org, "windowDays": windowDays})
		}
		return ListGitHubReviewersLeaderboardOutput{}, NewInternal("failed to load reviewers leaderboard snapshot", err)
	}

	if len(result.Reviewers) > limit {
		result.Reviewers = result.Reviewers[:limit]
	}

	out := ListGitHubReviewersLeaderboardOutput{
		Org:       result.Org,
		Summary:   result.Summary,
		Reviewers: result.Reviewers,
	}
	out.Window.From = result.WindowFrom
	out.Window.To = result.WindowTo
	return out, nil
}

type ListGitHubHangingPullRequestsInput struct {
	Org           string
	Limit         int
	MinIdleHours  int
	MinOpenHours  int
	Repos         []string
	Author        string
	IncludeDrafts bool
	Sort          string
}

type ListGitHubHangingPullRequestsOutput struct {
	Org         string                     `json:"org"`
	GeneratedAt time.Time                  `json:"generatedAt"`
	Criteria    map[string]any             `json:"criteria"`
	Summary     GitHubHangingSummary       `json:"summary"`
	Items       []GitHubHangingPullRequest `json:"items"`
}

type ListGitHubHangingPullRequestsUseCase struct {
	repository GitHubOrgInsightsRepository
	clock      Clock
}

func NewListGitHubHangingPullRequestsUseCase(repository GitHubOrgInsightsRepository, clock Clock) *ListGitHubHangingPullRequestsUseCase {
	return &ListGitHubHangingPullRequestsUseCase{repository: repository, clock: clock}
}

func (uc *ListGitHubHangingPullRequestsUseCase) Execute(ctx context.Context, in ListGitHubHangingPullRequestsInput) (ListGitHubHangingPullRequestsOutput, error) {
	org := strings.TrimSpace(in.Org)
	if org == "" {
		return ListGitHubHangingPullRequestsOutput{}, NewInvalidArgument("org is required", map[string]any{"field": "org"})
	}

	limit := in.Limit
	if limit == 0 {
		limit = 50
	}
	if limit < 1 || limit > 200 {
		return ListGitHubHangingPullRequestsOutput{}, NewInvalidArgument("limit must be between 1 and 200", map[string]any{"field": "limit"})
	}

	minIdleHours := in.MinIdleHours
	if minIdleHours == 0 {
		minIdleHours = 48
	}
	if minIdleHours < 1 {
		return ListGitHubHangingPullRequestsOutput{}, NewInvalidArgument("minIdleHours must be >= 1", map[string]any{"field": "minIdleHours"})
	}

	minOpenHours := in.MinOpenHours
	if minOpenHours == 0 {
		minOpenHours = 72
	}
	if minOpenHours < 1 {
		return ListGitHubHangingPullRequestsOutput{}, NewInvalidArgument("minOpenHours must be >= 1", map[string]any{"field": "minOpenHours"})
	}

	sortBy := in.Sort
	if sortBy == "" {
		sortBy = "staleness_desc"
	}
	if sortBy != "staleness_desc" && sortBy != "open_time_desc" {
		return ListGitHubHangingPullRequestsOutput{}, NewInvalidArgument("sort must be staleness_desc or open_time_desc", map[string]any{"field": "sort"})
	}

	result, err := uc.repository.GetLatestHangingSnapshot(ctx, org)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ListGitHubHangingPullRequestsOutput{}, NewNotFound("hanging pull requests snapshot not found", map[string]any{"org": org})
		}
		return ListGitHubHangingPullRequestsOutput{}, NewInternal("failed to load hanging pull requests snapshot", err)
	}

	items := make([]GitHubHangingPullRequest, 0, len(result.Items))
	repoFilters := uniqueNonEmpty(in.Repos)
	repoFilterSet := map[string]struct{}{}
	for _, repo := range repoFilters {
		repoFilterSet[strings.ToLower(repo)] = struct{}{}
	}
	author := strings.TrimSpace(in.Author)

	for _, item := range result.Items {
		if len(repoFilterSet) > 0 {
			if _, ok := repoFilterSet[strings.ToLower(item.Repository)]; !ok {
				continue
			}
		}
		if !in.IncludeDrafts && item.Draft {
			continue
		}
		if author != "" && !strings.EqualFold(item.Author, author) {
			continue
		}
		if item.IdleHours < minIdleHours || item.AgeHours < minOpenHours {
			continue
		}
		items = append(items, item)
	}
	if sortBy == "open_time_desc" {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].AgeHours > items[j].AgeHours
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].IdleHours > items[j].IdleHours
		})
	}

	if len(items) > limit {
		items = items[:limit]
	}

	result.Summary.HangingPRs = len(items)

	return ListGitHubHangingPullRequestsOutput{
		Org:         result.Org,
		GeneratedAt: result.GeneratedAt,
		Criteria: map[string]any{
			"minIdleHours":  minIdleHours,
			"minOpenHours":  minOpenHours,
			"includeDrafts": in.IncludeDrafts,
		},
		Summary: result.Summary,
		Items:   items,
	}, nil
}

type SyncGitHubOrgInsightsInput struct {
	Org        string
	WindowDays []int
}

type SyncGitHubOrgInsightsUseCase struct {
	service      GitHubOrgInsightsService
	repository   GitHubOrgInsightsRepository
	txManager    TransactionManager
	idGenerator  IDGenerator
	clock        Clock
	hangingLimit int
}

func NewSyncGitHubOrgInsightsUseCase(
	service GitHubOrgInsightsService,
	repository GitHubOrgInsightsRepository,
	txManager TransactionManager,
	idGenerator IDGenerator,
	clock Clock,
) *SyncGitHubOrgInsightsUseCase {
	return &SyncGitHubOrgInsightsUseCase{
		service:      service,
		repository:   repository,
		txManager:    txManager,
		idGenerator:  idGenerator,
		clock:        clock,
		hangingLimit: 5000,
	}
}

func (uc *SyncGitHubOrgInsightsUseCase) Execute(ctx context.Context, in SyncGitHubOrgInsightsInput) error {
	org := strings.TrimSpace(in.Org)
	if org == "" {
		return NewInvalidArgument("org is required", map[string]any{"field": "org"})
	}

	windows := in.WindowDays
	if len(windows) == 0 {
		windows = []int{30}
	}

	now := uc.clock.Now().UTC()

	for _, windowDays := range windows {
		if windowDays < 1 || windowDays > 365 {
			return NewInvalidArgument("windowDays must be between 1 and 365", map[string]any{"field": "windowDays"})
		}

		leaderboard, err := uc.service.GetReviewersLeaderboard(ctx, GitHubLeaderboardQuery{
			Org:          org,
			From:         now.AddDate(0, 0, -windowDays),
			To:           now,
			Limit:        5000,
			Repositories: nil,
		})
		if err != nil {
			if IsRateLimited(err) {
				return NewRateLimited("GitHub API rate limit exceeded", map[string]any{"provider": "github"})
			}
			return NewInternal("failed to sync reviewers leaderboard", err)
		}

		snapshot := GitHubReviewerSnapshot{
			SnapshotID:  uc.idGenerator.NewID(),
			Org:         org,
			WindowDays:  windowDays,
			WindowFrom:  leaderboard.From,
			WindowTo:    leaderboard.To,
			GeneratedAt: now,
			Summary:     leaderboard.Summary,
			Reviewers:   leaderboard.Reviewers,
		}

		if err := uc.txManager.WithinTx(ctx, func(txCtx context.Context) error {
			if err := uc.repository.SaveReviewersSnapshot(txCtx, snapshot); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	hanging, err := uc.service.GetHangingPullRequests(ctx, GitHubHangingQuery{
		Org:           org,
		Limit:         uc.hangingLimit,
		MinIdleHours:  0,
		MinOpenHours:  0,
		Repositories:  nil,
		Author:        "",
		IncludeDrafts: true,
		Sort:          "staleness_desc",
	})
	if err != nil {
		if IsRateLimited(err) {
			return NewRateLimited("GitHub API rate limit exceeded", map[string]any{"provider": "github"})
		}
		return NewInternal("failed to sync hanging pull requests", err)
	}

	hangingSnapshot := GitHubHangingSnapshot{
		SnapshotID:  uc.idGenerator.NewID(),
		Org:         org,
		GeneratedAt: now,
		Summary:     hanging.Summary,
		Items:       hanging.Items,
	}

	if err := uc.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if err := uc.repository.SaveHangingSnapshot(txCtx, hangingSnapshot); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func uniqueNonEmpty(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

type rateLimitedTaggedError interface {
	error
	RateLimited() bool
}

func IsRateLimited(err error) bool {
	var tagged rateLimitedTaggedError
	if errors.As(err, &tagged) {
		return tagged.RateLimited()
	}
	return false
}
