package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/repository"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresPullRequestRepository struct {
	logger *slog.Logger
}

func NewPostgresPullRequestRepository(baseLogger *slog.Logger) repository.BasePullRequestRepository {
	logger := baseLogger.With("module", "prrepo")
	return &PostgresPullRequestRepository{
		logger: logger,
	}
}

func (p *PostgresPullRequestRepository) GetPullRequestsByReviewerId(ctx context.Context, db repository.Querier, reviewerId string) ([]entity.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, updated_at FROM pull_requests pr
		JOIN pull_requests_users pr_u ON pr.id = pr_u.pr_id
        WHERE pr_u.user_id = $1
	`
	var prs []entity.PullRequest

	rows, err := db.Query(ctx, query, reviewerId)
	if err != nil {
		p.logger.Debug("failed to GetPullRequestsByReviewerId", "reviewerId", reviewerId, "err", err)
		return nil, errs.ErrInternal("failed to GetPullRequestsByReviewerId", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pr entity.PullRequest
		err := rows.Scan(
			&pr.Id,
			&pr.PullRequestName,
			&pr.AuthorId,
			&pr.Status,
			&pr.UpdatedAt,
		)
		if err != nil {
			p.logger.Debug("failed to GetPullRequestsByReviewerId: scan error", "reviewerId", reviewerId, "err", err)
			return nil, errs.ErrInternal("failed to GetPullRequestsByReviewerId: scan error", err)
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

func (p *PostgresPullRequestRepository) GetPullRequestById(ctx context.Context, db repository.Querier, prId string) (*entity.PullRequest, error) {
	query := `
		SELECT id, name, author_id, status, updated_at FROM pull_requests 
        WHERE id = $1
	`
	var pr entity.PullRequest
	err := db.QueryRow(ctx, query, prId).Scan(
		&pr.Id,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		&pr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.logger.Debug("failed to GetPullRequestById: not found", "prId", prId)
			return nil, errs.ErrNotFound("pull request", "id", prId)
		}
		p.logger.Debug("failed to GetPullRequestById", "prId", prId, "err", err)
		return nil, errs.ErrInternal("failed to GetPullRequestById", err)
	}
	return &pr, nil
}

func (p *PostgresPullRequestRepository) AddPullRequest(ctx context.Context, db repository.Querier, ent *entity.PullRequest) error {
	query := `
		INSERT INTO pull_requests (id, name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(ctx, query, ent.Id, ent.PullRequestName, ent.AuthorId, ent.Status)
	if err != nil {
		p.logger.Debug("failed to AddPullRequest", "err", err)
		return errs.ErrInternal("failed to AddPullRequest", err)
	}
	return nil
}

func (p *PostgresPullRequestRepository) UpdatePullRequestStatus(ctx context.Context, db repository.Querier, prId string, newStatus string) error {
	query := `
		UPDATE pull_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	ct, err := db.Exec(ctx, query, newStatus, time.Now(), prId)
	if err != nil {
		p.logger.Debug("failed to UpdatePullRequestStatus", "prId", prId, "newStatus", newStatus, "err", err)
		return errs.ErrInternal("failed to UpdatePullRequestStatus", err)
	}
	if ct.RowsAffected() == 0 {
		p.logger.Debug("failed to UpdatePullRequestStatus: not found", "prId", prId, "newStatus", newStatus)
		return errs.ErrNotFound("pull request", "id", prId)
	}
	return nil
}

func (p *PostgresPullRequestRepository) AddReviewerToPullRequest(ctx context.Context, db repository.Querier, prId string, reviewerId string) error {
	query := `
		INSERT INTO pull_requests_users (user_id, pr_id)
		VALUES ($1, $2)
	`
	_, err := db.Exec(ctx, query, reviewerId, prId)
	if err != nil {
		p.logger.Debug("failed to AddReviewerToPullRequest", "prId", prId, "reviewerId", reviewerId, "err", err)
		return errs.ErrInternal("failed to AddReviewerToPullRequest", err)
	}
	return nil
}

func (p *PostgresPullRequestRepository) RemoveReviewerFromPullRequest(ctx context.Context, db repository.Querier, prId string, reviewerId string) error {
	query := `
		DELETE FROM pull_requests_users 
		WHERE user_id = $1 AND pr_id = $2
	`
	ct, err := db.Exec(ctx, query, reviewerId, prId)
	if err != nil {
		p.logger.Debug("failed to RemoveReviewerFromPullRequest", "prId", prId, "reviewerId", reviewerId, "err", err)
		return errs.ErrInternal("failed to RemoveReviewerFromPullRequest", err)
	}
	if ct.RowsAffected() == 0 {
		p.logger.Debug("failed to RemoveReviewerFromPullRequest: not found", "prId", prId, "reviewerId", reviewerId)
		return errs.ErrNotFound("pull request", "reviewerId and prId", fmt.Sprintf("%s, %s", reviewerId, prId))
	}
	return nil
}
