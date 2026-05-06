package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/arxdsilva/opencoverage/internal/application"
	"github.com/arxdsilva/opencoverage/internal/domain"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GitHubOrgInsightsRepository struct {
	pool *pgxpool.Pool
}

func NewGitHubOrgInsightsRepository(pool *pgxpool.Pool) *GitHubOrgInsightsRepository {
	return &GitHubOrgInsightsRepository{pool: pool}
}

func (r *GitHubOrgInsightsRepository) SaveReviewersSnapshot(ctx context.Context, snapshot application.GitHubReviewerSnapshot) error {
	q := getQuerier(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO github_reviewer_snapshots (
			id, org_name, window_days, window_from, window_to,
			repositories_scanned, pull_requests_considered, total_review_events,
			generated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		snapshot.SnapshotID,
		snapshot.Org,
		snapshot.WindowDays,
		snapshot.WindowFrom,
		snapshot.WindowTo,
		snapshot.Summary.RepositoriesScanned,
		snapshot.Summary.PullRequestsScanned,
		snapshot.Summary.TotalReviewEvents,
		snapshot.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("insert github reviewer snapshot: %w", err)
	}

	for _, reviewer := range snapshot.Reviewers {
		entryID := uuid.NewString()
		_, err := q.Exec(ctx, `
			INSERT INTO github_reviewer_entries (
				id, snapshot_id, login, display_name,
				total_reviews, approvals, change_requests, comments,
				unique_pull_requests_reviewed, repos_reviewed, latest_review_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			entryID,
			snapshot.SnapshotID,
			reviewer.Login,
			reviewer.DisplayName,
			reviewer.TotalReviews,
			reviewer.Approvals,
			reviewer.ChangeRequests,
			reviewer.Comments,
			reviewer.UniquePullRequestsReviewed,
			reviewer.RepositoriesReviewed,
			reviewer.LatestReviewAt,
		)
		if err != nil {
			return fmt.Errorf("insert github reviewer entry: %w", err)
		}
	}

	return nil
}

func (r *GitHubOrgInsightsRepository) GetLatestReviewersSnapshot(ctx context.Context, org string, windowDays int) (application.GitHubReviewerSnapshot, error) {
	q := getQuerier(ctx, r.pool)

	var snapshot application.GitHubReviewerSnapshot
	err := q.QueryRow(ctx, `
		SELECT id, org_name, window_days, window_from, window_to,
			repositories_scanned, pull_requests_considered, total_review_events,
			generated_at
		FROM github_reviewer_snapshots
		WHERE org_name = $1 AND window_days = $2
		ORDER BY generated_at DESC, created_at DESC
		LIMIT 1
	`, org, windowDays).Scan(
		&snapshot.SnapshotID,
		&snapshot.Org,
		&snapshot.WindowDays,
		&snapshot.WindowFrom,
		&snapshot.WindowTo,
		&snapshot.Summary.RepositoriesScanned,
		&snapshot.Summary.PullRequestsScanned,
		&snapshot.Summary.TotalReviewEvents,
		&snapshot.GeneratedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return application.GitHubReviewerSnapshot{}, domain.ErrNotFound
		}
		return application.GitHubReviewerSnapshot{}, fmt.Errorf("query latest github reviewer snapshot: %w", err)
	}

	rows, err := q.Query(ctx, `
		SELECT login, COALESCE(display_name, ''), total_reviews, approvals,
			change_requests, comments, unique_pull_requests_reviewed,
			repos_reviewed, latest_review_at
		FROM github_reviewer_entries
		WHERE snapshot_id = $1
		ORDER BY total_reviews DESC, approvals DESC, latest_review_at DESC, login ASC
	`, snapshot.SnapshotID)
	if err != nil {
		return application.GitHubReviewerSnapshot{}, fmt.Errorf("query github reviewer entries: %w", err)
	}
	defer rows.Close()

	reviewers := make([]application.GitHubReviewer, 0)
	for rows.Next() {
		var reviewer application.GitHubReviewer
		if err := rows.Scan(
			&reviewer.Login,
			&reviewer.DisplayName,
			&reviewer.TotalReviews,
			&reviewer.Approvals,
			&reviewer.ChangeRequests,
			&reviewer.Comments,
			&reviewer.UniquePullRequestsReviewed,
			&reviewer.RepositoriesReviewed,
			&reviewer.LatestReviewAt,
		); err != nil {
			return application.GitHubReviewerSnapshot{}, fmt.Errorf("scan github reviewer entry: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}
	if err := rows.Err(); err != nil {
		return application.GitHubReviewerSnapshot{}, fmt.Errorf("iterate github reviewer entries: %w", err)
	}

	snapshot.Reviewers = reviewers
	return snapshot, nil
}

func (r *GitHubOrgInsightsRepository) SaveHangingSnapshot(ctx context.Context, snapshot application.GitHubHangingSnapshot) error {
	q := getQuerier(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO github_hanging_snapshots (
			id, org_name, repositories_scanned, open_pull_requests_considered,
			hanging_pull_requests, generated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		snapshot.SnapshotID,
		snapshot.Org,
		snapshot.Summary.RepositoriesScanned,
		snapshot.Summary.OpenPRsScanned,
		snapshot.Summary.HangingPRs,
		snapshot.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("insert github hanging snapshot: %w", err)
	}

	for _, item := range snapshot.Items {
		itemID := uuid.NewString()
		_, err := q.Exec(ctx, `
			INSERT INTO github_hanging_items (
				id, snapshot_id, repository_name, pr_number, title, url,
				author, draft, created_at_pr, updated_at_pr, last_activity_at,
				age_hours, idle_hours, review_state, mergeable_state,
				requested_reviewers, labels, reasons
			)
			VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11,
				$12, $13, $14, $15,
				$16, $17, $18
			)
		`,
			itemID,
			snapshot.SnapshotID,
			item.Repository,
			item.Number,
			item.Title,
			item.URL,
			item.Author,
			item.Draft,
			item.CreatedAt,
			item.UpdatedAt,
			item.LastActivityAt,
			item.AgeHours,
			item.IdleHours,
			item.ReviewState,
			item.MergeableState,
			item.RequestedReviewers,
			item.Labels,
			item.Reasons,
		)
		if err != nil {
			return fmt.Errorf("insert github hanging item: %w", err)
		}
	}

	return nil
}

func (r *GitHubOrgInsightsRepository) GetLatestHangingSnapshot(ctx context.Context, org string) (application.GitHubHangingSnapshot, error) {
	q := getQuerier(ctx, r.pool)

	var snapshot application.GitHubHangingSnapshot
	err := q.QueryRow(ctx, `
		SELECT id, org_name, repositories_scanned, open_pull_requests_considered,
			hanging_pull_requests, generated_at
		FROM github_hanging_snapshots
		WHERE org_name = $1
		ORDER BY generated_at DESC, created_at DESC
		LIMIT 1
	`, org).Scan(
		&snapshot.SnapshotID,
		&snapshot.Org,
		&snapshot.Summary.RepositoriesScanned,
		&snapshot.Summary.OpenPRsScanned,
		&snapshot.Summary.HangingPRs,
		&snapshot.GeneratedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return application.GitHubHangingSnapshot{}, domain.ErrNotFound
		}
		return application.GitHubHangingSnapshot{}, fmt.Errorf("query latest github hanging snapshot: %w", err)
	}

	rows, err := q.Query(ctx, `
		SELECT repository_name, pr_number, title, url,
			author, draft, created_at_pr, updated_at_pr, last_activity_at,
			age_hours, idle_hours, review_state, mergeable_state,
			requested_reviewers, labels, reasons
		FROM github_hanging_items
		WHERE snapshot_id = $1
	`, snapshot.SnapshotID)
	if err != nil {
		return application.GitHubHangingSnapshot{}, fmt.Errorf("query github hanging items: %w", err)
	}
	defer rows.Close()

	items := make([]application.GitHubHangingPullRequest, 0)
	for rows.Next() {
		var item application.GitHubHangingPullRequest
		if err := rows.Scan(
			&item.Repository,
			&item.Number,
			&item.Title,
			&item.URL,
			&item.Author,
			&item.Draft,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.LastActivityAt,
			&item.AgeHours,
			&item.IdleHours,
			&item.ReviewState,
			&item.MergeableState,
			&item.RequestedReviewers,
			&item.Labels,
			&item.Reasons,
		); err != nil {
			return application.GitHubHangingSnapshot{}, fmt.Errorf("scan github hanging item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return application.GitHubHangingSnapshot{}, fmt.Errorf("iterate github hanging items: %w", err)
	}

	snapshot.Items = items
	return snapshot, nil
}
