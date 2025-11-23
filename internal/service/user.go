package service

import (
	"context"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	logger   *slog.Logger
	pool     *pgxpool.Pool
	userRepo repository.BaseUserRepository
	prRepo   repository.BasePullRequestRepository
}

func NewUserService(
	baseLogger *slog.Logger,
	pool *pgxpool.Pool,
	userRepo repository.BaseUserRepository,
	prRepo repository.BasePullRequestRepository,
) BaseUserService {
	logger := baseLogger.With("module", "userservice")
	return &UserService{
		logger:   logger,
		pool:     pool,
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetIsActive(
	ctx context.Context,
	dto entity.SetUserIsActiveDTO,
) (*entity.UserDTO, error) {
	exists, err := s.userRepo.GetById(ctx, s.pool, dto.UserId)
	if err != nil {
		s.logger.Debug("failed to SetIsActive: GetById failed", "err", err)
		return nil, err
	}

	err = s.userRepo.UpdateUser(ctx, s.pool, exists.Id, &entity.UserUpdate{
		IsActive: &dto.IsActive,
	})
	if err != nil {
		s.logger.Debug("failed to SetIsActive: UpdateUser failed", "err", err)
		return nil, err
	}

	return &entity.UserDTO{
		UserId:   exists.Id,
		Username: exists.Username,
		IsActive: dto.IsActive,
	}, nil
}

func (s *UserService) GetReview(
	ctx context.Context,
	userId string,
) (*entity.UserPullRequestsDTO, error) {
	exists, err := s.userRepo.GetById(ctx, s.pool, userId)
	if err != nil {
		s.logger.Debug("failed to GetReview: GetById failed", "err", err)
		return nil, err
	}

	prs, err := s.prRepo.GetPullRequestsByReviewerId(ctx, s.pool, exists.Id)
	if err != nil {
		s.logger.Debug("failed to GetReview: GetPullRequestsByReviewerId failed", "err", err)
		return nil, err
	}

	prsDTO := make([]entity.PullRequestDTO, len(prs))
	for i, pr := range prs {
		prsDTO[i] = entity.PullRequestDTO{
			PullRequestId:   pr.Id,
			PullRequestName: pr.PullRequestName,
			AuthorId:        pr.AuthorId,
			Status:          pr.Status,
		}
	}
	return &entity.UserPullRequestsDTO{
		UserId:       userId,
		PullRequests: prsDTO,
	}, nil
}
