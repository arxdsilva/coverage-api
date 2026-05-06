package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arxdsilva/opencoverage/internal/application"
)

type cacheEntry struct {
	createdAt time.Time
	value     any
}

type RateLimitedError struct {
	message string
}

func (e *RateLimitedError) Error() string     { return e.message }
func (e *RateLimitedError) RateLimited() bool { return true }

type OrgInsightsService struct {
	baseURL  string
	token    string
	client   *http.Client
	maxRepos int
	ttl      time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

func NewOrgInsightsService(baseURL string, token string, maxRepos int, ttl time.Duration) *OrgInsightsService {
	cleanBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if cleanBaseURL == "" {
		cleanBaseURL = "https://api.github.com"
	}
	if maxRepos <= 0 {
		maxRepos = 200
	}
	if ttl < 0 {
		ttl = 0
	}

	return &OrgInsightsService{
		baseURL:  cleanBaseURL,
		token:    strings.TrimSpace(token),
		client:   &http.Client{Timeout: 20 * time.Second},
		maxRepos: maxRepos,
		ttl:      ttl,
		cache:    map[string]cacheEntry{},
	}
}

func (s *OrgInsightsService) GetReviewersLeaderboard(ctx context.Context, query application.GitHubLeaderboardQuery) (application.GitHubReviewerLeaderboard, error) {
	cacheKey := s.leaderboardCacheKey(query)
	if cached, ok := s.getCached(cacheKey); ok {
		return cached.(application.GitHubReviewerLeaderboard), nil
	}

	repositories, err := s.resolveRepositories(ctx, query.Org, query.Repositories)
	if err != nil {
		return application.GitHubReviewerLeaderboard{}, err
	}

	type reviewerAcc struct {
		reviewer     application.GitHubReviewer
		prs          map[string]struct{}
		repositories map[string]struct{}
	}

	reviewers := map[string]*reviewerAcc{}
	totalPullRequests := 0
	totalEvents := 0

	for _, repo := range repositories {
		pulls, err := s.listPullsUpdatedInRange(ctx, query.Org, repo, query.From, query.To, "all")
		if err != nil {
			return application.GitHubReviewerLeaderboard{}, err
		}
		totalPullRequests += len(pulls)

		for _, pull := range pulls {
			reviews, err := s.listReviews(ctx, query.Org, repo, pull.Number)
			if err != nil {
				return application.GitHubReviewerLeaderboard{}, err
			}

			for _, review := range reviews {
				if review.User.Login == "" || review.SubmittedAt.IsZero() {
					continue
				}
				if review.SubmittedAt.Before(query.From) || review.SubmittedAt.After(query.To) {
					continue
				}
				state := strings.ToUpper(strings.TrimSpace(review.State))
				if state != "APPROVED" && state != "CHANGES_REQUESTED" && state != "COMMENTED" {
					continue
				}

				totalEvents++
				acc := reviewers[review.User.Login]
				if acc == nil {
					acc = &reviewerAcc{
						reviewer: application.GitHubReviewer{
							Login:       review.User.Login,
							DisplayName: strings.TrimSpace(review.User.Name),
						},
						prs:          map[string]struct{}{},
						repositories: map[string]struct{}{},
					}
					reviewers[review.User.Login] = acc
				}

				acc.reviewer.TotalReviews++
				switch state {
				case "APPROVED":
					acc.reviewer.Approvals++
				case "CHANGES_REQUESTED":
					acc.reviewer.ChangeRequests++
				case "COMMENTED":
					acc.reviewer.Comments++
				}

				prKey := fmt.Sprintf("%s#%d", repo, pull.Number)
				acc.prs[prKey] = struct{}{}
				acc.repositories[repo] = struct{}{}
				if review.SubmittedAt.After(acc.reviewer.LatestReviewAt) {
					acc.reviewer.LatestReviewAt = review.SubmittedAt
				}
			}
		}
	}

	result := make([]application.GitHubReviewer, 0, len(reviewers))
	for _, acc := range reviewers {
		acc.reviewer.UniquePullRequestsReviewed = len(acc.prs)
		acc.reviewer.RepositoriesReviewed = len(acc.repositories)
		result = append(result, acc.reviewer)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalReviews != result[j].TotalReviews {
			return result[i].TotalReviews > result[j].TotalReviews
		}
		if result[i].Approvals != result[j].Approvals {
			return result[i].Approvals > result[j].Approvals
		}
		if !result[i].LatestReviewAt.Equal(result[j].LatestReviewAt) {
			return result[i].LatestReviewAt.After(result[j].LatestReviewAt)
		}
		return result[i].Login < result[j].Login
	})

	if len(result) > query.Limit {
		result = result[:query.Limit]
	}

	out := application.GitHubReviewerLeaderboard{
		Org:  query.Org,
		From: query.From,
		To:   query.To,
		Summary: application.GitHubLeaderboardSummary{
			RepositoriesScanned: len(repositories),
			PullRequestsScanned: totalPullRequests,
			TotalReviewEvents:   totalEvents,
		},
		Reviewers: result,
	}

	s.setCached(cacheKey, out)
	return out, nil
}

func (s *OrgInsightsService) GetHangingPullRequests(ctx context.Context, query application.GitHubHangingQuery) (application.GitHubHangingPullRequests, error) {
	cacheKey := s.hangingCacheKey(query)
	if cached, ok := s.getCached(cacheKey); ok {
		return cached.(application.GitHubHangingPullRequests), nil
	}

	repositories, err := s.resolveRepositories(ctx, query.Org, query.Repositories)
	if err != nil {
		return application.GitHubHangingPullRequests{}, err
	}

	now := time.Now().UTC()
	items := make([]application.GitHubHangingPullRequest, 0)
	considered := 0

	for _, repo := range repositories {
		pulls, err := s.listOpenPulls(ctx, query.Org, repo)
		if err != nil {
			return application.GitHubHangingPullRequests{}, err
		}

		for _, pull := range pulls {
			considered++
			if !query.IncludeDrafts && pull.Draft {
				continue
			}
			if query.Author != "" && !strings.EqualFold(pull.User.Login, query.Author) {
				continue
			}

			ageHours := int(now.Sub(pull.CreatedAt).Hours())
			idleHours := int(now.Sub(pull.UpdatedAt).Hours())
			if ageHours < query.MinOpenHours || idleHours < query.MinIdleHours {
				continue
			}

			reviews, err := s.listReviews(ctx, query.Org, repo, pull.Number)
			if err != nil {
				return application.GitHubHangingPullRequests{}, err
			}

			reviewState := latestReviewState(reviews)
			reasons := make([]string, 0, 2)
			if len(reviews) == 0 {
				reasons = append(reasons, "awaiting-review")
			}
			if reviewState == "changes_requested" {
				if len(pull.RequestedReviewers) == 0 {
					reasons = append(reasons, "awaiting-author")
				} else {
					reasons = append(reasons, "changes-requested")
				}
			}

			mergeableState := ""
			detail, err := s.getPullDetail(ctx, query.Org, repo, pull.Number)
			if err == nil {
				mergeableState = strings.TrimSpace(detail.MergeableState)
				if strings.EqualFold(mergeableState, "dirty") {
					reasons = append(reasons, "merge-conflict")
				}
			}

			if len(reasons) == 0 {
				continue
			}

			item := application.GitHubHangingPullRequest{
				Repository:         repo,
				Number:             pull.Number,
				Title:              pull.Title,
				URL:                pull.HTMLURL,
				Author:             pull.User.Login,
				Draft:              pull.Draft,
				CreatedAt:          pull.CreatedAt,
				UpdatedAt:          pull.UpdatedAt,
				LastActivityAt:     pull.UpdatedAt,
				AgeHours:           ageHours,
				IdleHours:          idleHours,
				ReviewState:        reviewState,
				MergeableState:     strings.ToLower(mergeableState),
				RequestedReviewers: mapUserLogins(pull.RequestedReviewers),
				Labels:             mapLabelNames(pull.Labels),
				Reasons:            reasons,
			}
			items = append(items, item)
		}
	}

	if query.Sort == "open_time_desc" {
		sort.SliceStable(items, func(i, j int) bool { return items[i].AgeHours > items[j].AgeHours })
	} else {
		sort.SliceStable(items, func(i, j int) bool { return items[i].IdleHours > items[j].IdleHours })
	}

	if len(items) > query.Limit {
		items = items[:query.Limit]
	}

	out := application.GitHubHangingPullRequests{
		Org:         query.Org,
		GeneratedAt: now,
		Summary: application.GitHubHangingSummary{
			RepositoriesScanned: len(repositories),
			OpenPRsScanned:      considered,
			HangingPRs:          len(items),
		},
		Items: items,
	}

	s.setCached(cacheKey, out)
	return out, nil
}

func latestReviewState(reviews []githubReview) string {
	if len(reviews) == 0 {
		return ""
	}
	latest := reviews[0]
	for _, review := range reviews[1:] {
		if review.SubmittedAt.After(latest.SubmittedAt) {
			latest = review
		}
	}
	state := strings.ToLower(strings.TrimSpace(latest.State))
	switch state {
	case "approved", "changes_requested", "commented":
		return state
	default:
		return ""
	}
}

func mapUserLogins(users []githubUser) []string {
	logins := make([]string, 0, len(users))
	for _, user := range users {
		if trimmed := strings.TrimSpace(user.Login); trimmed != "" {
			logins = append(logins, trimmed)
		}
	}
	return logins
}

func mapLabelNames(labels []githubLabel) []string {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		if trimmed := strings.TrimSpace(label.Name); trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

func (s *OrgInsightsService) resolveRepositories(ctx context.Context, org string, requested []string) ([]string, error) {
	if len(requested) > 0 {
		unique := map[string]struct{}{}
		out := make([]string, 0, len(requested))
		for _, repo := range requested {
			name := strings.TrimSpace(repo)
			if name == "" {
				continue
			}
			if _, ok := unique[name]; ok {
				continue
			}
			unique[name] = struct{}{}
			out = append(out, name)
		}
		return out, nil
	}
	return s.listOrgRepos(ctx, org)
}

func (s *OrgInsightsService) listOrgRepos(ctx context.Context, org string) ([]string, error) {
	repos := make([]string, 0)
	page := 1
	for len(repos) < s.maxRepos {
		u := fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=%d", s.baseURL, url.PathEscape(org), page)
		var response []githubRepo
		if err := s.doJSON(ctx, http.MethodGet, u, &response); err != nil {
			return nil, err
		}
		if len(response) == 0 {
			break
		}
		for _, repo := range response {
			if repo.Archived || repo.Disabled {
				continue
			}
			repos = append(repos, repo.Name)
			if len(repos) >= s.maxRepos {
				break
			}
		}
		page++
	}
	return repos, nil
}

func (s *OrgInsightsService) listPullsUpdatedInRange(ctx context.Context, org string, repo string, from time.Time, to time.Time, state string) ([]githubPull, error) {
	items := make([]githubPull, 0)
	page := 1
	for {
		u := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s&sort=updated&direction=desc&per_page=100&page=%d", s.baseURL, url.PathEscape(org), url.PathEscape(repo), url.QueryEscape(state), page)
		var response []githubPull
		if err := s.doJSON(ctx, http.MethodGet, u, &response); err != nil {
			return nil, err
		}
		if len(response) == 0 {
			break
		}

		stop := false
		for _, pull := range response {
			if pull.UpdatedAt.Before(from) {
				stop = true
				break
			}
			if pull.UpdatedAt.After(to) {
				continue
			}
			items = append(items, pull)
		}
		if stop {
			break
		}
		page++
	}
	return items, nil
}

func (s *OrgInsightsService) listOpenPulls(ctx context.Context, org string, repo string) ([]githubPull, error) {
	items := make([]githubPull, 0)
	page := 1
	for {
		u := fmt.Sprintf("%s/repos/%s/%s/pulls?state=open&sort=updated&direction=desc&per_page=100&page=%d", s.baseURL, url.PathEscape(org), url.PathEscape(repo), page)
		var response []githubPull
		if err := s.doJSON(ctx, http.MethodGet, u, &response); err != nil {
			return nil, err
		}
		if len(response) == 0 {
			break
		}
		items = append(items, response...)
		page++
	}
	return items, nil
}

func (s *OrgInsightsService) listReviews(ctx context.Context, org string, repo string, number int) ([]githubReview, error) {
	items := make([]githubReview, 0)
	page := 1
	for {
		u := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews?per_page=100&page=%d", s.baseURL, url.PathEscape(org), url.PathEscape(repo), number, page)
		var response []githubReview
		if err := s.doJSON(ctx, http.MethodGet, u, &response); err != nil {
			return nil, err
		}
		if len(response) == 0 {
			break
		}
		items = append(items, response...)
		page++
	}
	return items, nil
}

func (s *OrgInsightsService) getPullDetail(ctx context.Context, org string, repo string, number int) (githubPullDetail, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", s.baseURL, url.PathEscape(org), url.PathEscape(repo), number)
	var detail githubPullDetail
	err := s.doJSON(ctx, http.MethodGet, u, &detail)
	return detail, err
}

func (s *OrgInsightsService) doJSON(ctx context.Context, method string, rawURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		if resp.Header.Get("X-RateLimit-Remaining") == "0" || resp.StatusCode == http.StatusTooManyRequests {
			return &RateLimitedError{message: "github api rate limit exceeded"}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("github api status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (s *OrgInsightsService) leaderboardCacheKey(query application.GitHubLeaderboardQuery) string {
	parts := []string{
		"leaderboard",
		query.Org,
		query.From.UTC().Format(time.RFC3339),
		query.To.UTC().Format(time.RFC3339),
		strconv.Itoa(query.Limit),
		strings.Join(query.Repositories, ","),
	}
	return strings.Join(parts, "|")
}

func (s *OrgInsightsService) hangingCacheKey(query application.GitHubHangingQuery) string {
	parts := []string{
		"hanging",
		query.Org,
		strconv.Itoa(query.Limit),
		strconv.Itoa(query.MinIdleHours),
		strconv.Itoa(query.MinOpenHours),
		strings.Join(query.Repositories, ","),
		strings.ToLower(strings.TrimSpace(query.Author)),
		strconv.FormatBool(query.IncludeDrafts),
		query.Sort,
	}
	return strings.Join(parts, "|")
}

func (s *OrgInsightsService) getCached(key string) (any, bool) {
	if s.ttl <= 0 {
		return nil, false
	}
	s.mu.RLock()
	entry, ok := s.cache[key]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Since(entry.createdAt) > s.ttl {
		s.mu.Lock()
		delete(s.cache, key)
		s.mu.Unlock()
		return nil, false
	}
	return entry.value, true
}

func (s *OrgInsightsService) setCached(key string, value any) {
	if s.ttl <= 0 {
		return
	}
	s.mu.Lock()
	s.cache[key] = cacheEntry{createdAt: time.Now().UTC(), value: value}
	s.mu.Unlock()
}

type githubRepo struct {
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
	Disabled bool   `json:"disabled"`
}

type githubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type githubLabel struct {
	Name string `json:"name"`
}

type githubPull struct {
	Number             int           `json:"number"`
	Title              string        `json:"title"`
	HTMLURL            string        `json:"html_url"`
	Draft              bool          `json:"draft"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	User               githubUser    `json:"user"`
	RequestedReviewers []githubUser  `json:"requested_reviewers"`
	Labels             []githubLabel `json:"labels"`
}

type githubPullDetail struct {
	MergeableState string `json:"mergeable_state"`
}

type githubReview struct {
	State       string     `json:"state"`
	SubmittedAt time.Time  `json:"submitted_at"`
	User        githubUser `json:"user"`
}
