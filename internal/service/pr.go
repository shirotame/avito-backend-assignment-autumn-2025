package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestService struct {
	logger   *slog.Logger
	pool     *pgxpool.Pool
	prRepo   repository.BasePullRequestRepository
	userRepo repository.BaseUserRepository
}

func NewPullRequestService(
	baseLogger *slog.Logger,
	pool *pgxpool.Pool,
	prRepo repository.BasePullRequestRepository,
	userRepo repository.BaseUserRepository,
) BasePullRequestService {
	logger := baseLogger.With("module", "prservice")
	return &PullRequestService{
		logger:   logger,
		pool:     pool,
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *PullRequestService) CreatePullRequest(
	ctx context.Context,
	dto entity.PullRequestCreateDTO,
) (*entity.PullRequestResponseDTO, error) {
	exists, err := s.prRepo.GetPullRequestById(ctx, s.pool, dto.PullRequestId)
	if err != nil && !errors.Is(err, errs.ErrBaseNotFound) {
		return nil, err
	}
	if exists != nil {
		return nil, errs.ErrPullRequestAlreadyExists
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errs.ErrInternal("error begin transaction", err)
	}
	defer tx.Rollback(ctx)

	err = s.prRepo.AddPullRequest(ctx, tx, &entity.PullRequest{
		Id:              dto.PullRequestId,
		PullRequestName: dto.PullRequestName,
		AuthorId:        dto.AuthorId,
		Status:          entity.StatusOpen,
	})
	if err != nil {
		return nil, err
	}

	author, err := s.userRepo.GetById(ctx, tx, dto.AuthorId)
	if err != nil {
		return nil, err
	}

	availableUsers, err := s.userRepo.GetActiveByTeamName(ctx, tx, author.TeamName)
	if err != nil {
		return nil, err
	}

	assigned := make([]string, 0, 2)
	validCount := 0
	for _, user := range availableUsers {
		if user.Id == dto.AuthorId {
			continue
		}

		validCount++

		if len(assigned) < 2 {
			assigned = append(assigned, user.Id)
		} else {
			j := rand.IntN(validCount)
			if j < 2 {
				assigned[j] = user.Id
			}
		}
	}

	for _, aId := range assigned {
		err := s.prRepo.AddReviewerToPullRequest(ctx, tx, dto.PullRequestId, aId)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.ErrInternal("error commit transaction", err)
	}

	return &entity.PullRequestResponseDTO{
		PullRequest: entity.PullRequestDTO{
			PullRequestId:     dto.PullRequestId,
			PullRequestName:   dto.PullRequestName,
			AuthorId:          dto.AuthorId,
			Status:            entity.StatusOpen,
			AssignedReviewers: assigned,
		},
	}, nil
}

func (s *PullRequestService) MergePullRequest(
	ctx context.Context,
	dto entity.MergePullRequestDTO,
) (*entity.PullRequestResponseDTO, error) {
	exists, err := s.prRepo.GetPullRequestById(ctx, s.pool, dto.PullRequestId)
	if err != nil {
		return nil, err
	}

	assigned, err := s.userRepo.GetReviewersByPrId(ctx, s.pool, dto.PullRequestId)
	if err != nil {
		return nil, err
	}
	assignedIds := make([]string, len(assigned))
	for i, u := range assigned {
		assignedIds[i] = u.Id
	}

	if exists.Status == entity.StatusMerged {
		fmtTime := exists.UpdatedAt.Format(time.RFC3339)

		return &entity.PullRequestResponseDTO{
			PullRequest: entity.PullRequestDTO{
				PullRequestId:     exists.Id,
				PullRequestName:   exists.PullRequestName,
				AuthorId:          exists.AuthorId,
				Status:            exists.Status,
				AssignedReviewers: assignedIds,
				MergedAt:          &fmtTime,
			},
		}, nil
	}

	currTimeAsStr := time.Now().Format(time.RFC3339)
	err = s.prRepo.UpdatePullRequestStatus(ctx, s.pool, exists.Id, entity.StatusMerged)
	if err != nil {
		return nil, err
	}

	return &entity.PullRequestResponseDTO{
		PullRequest: entity.PullRequestDTO{
			PullRequestId:     exists.Id,
			PullRequestName:   exists.PullRequestName,
			AuthorId:          exists.AuthorId,
			Status:            entity.StatusMerged,
			AssignedReviewers: assignedIds,
			MergedAt:          &currTimeAsStr,
		},
	}, err
}

func (s *PullRequestService) ReassignPullRequest(
	ctx context.Context,
	dto entity.ReassignPullRequestDTO,
) (*entity.PullRequestResponseDTO, error) {
	exists, err := s.prRepo.GetPullRequestById(ctx, s.pool, dto.PullRequestId)
	if err != nil {
		return nil, err
	}
	if exists.Status == entity.StatusMerged {
		return nil, errs.ErrReassignOnMergedPR
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errs.ErrInternal("error begin transaction", err)
	}
	defer tx.Rollback(ctx)

	err = s.prRepo.RemoveReviewerFromPullRequest(ctx, tx, exists.Id, dto.OldReviewerId)
	if err != nil {
		if errors.Is(err, errs.ErrBaseNotFound) {
			return nil, errs.ErrUserNotAssigned
		}
		return nil, err
	}

	oldReviewer, err := s.userRepo.GetById(ctx, tx, dto.OldReviewerId)
	if err != nil {
		return nil, err
	}
	activeUsers, err := s.userRepo.GetActiveByTeamName(ctx, tx, oldReviewer.TeamName)
	if err != nil {
		return nil, err
	}
	if len(activeUsers) == 0 {
		return nil, errs.ErrNoActiveUsers
	}

	var newAssignedIdPtr string
	for range len(activeUsers) {
		rndId := rand.IntN(len(activeUsers))
		currId := activeUsers[rndId].Id

		if currId == dto.OldReviewerId && len(activeUsers) == 1 {
			return nil, errs.ErrNoActiveUsers
		}
		if currId != dto.OldReviewerId && currId != exists.AuthorId {
			err := s.prRepo.AddReviewerToPullRequest(ctx, tx, dto.PullRequestId, currId)
			if err != nil {
				return nil, err
			}
			newAssignedIdPtr = currId
			break
		}
	}

	if newAssignedIdPtr == "" {
		return nil, errs.ErrNoActiveUsers
	}

	assigned, err := s.userRepo.GetReviewersByPrId(ctx, tx, dto.PullRequestId)
	if err != nil {
		return nil, err
	}
	assignedIds := make([]string, len(assigned))
	for i, u := range assigned {
		assignedIds[i] = u.Id
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.ErrInternal("error commit transaction", err)
	}

	return &entity.PullRequestResponseDTO{
		PullRequest: entity.PullRequestDTO{
			PullRequestId:     exists.Id,
			PullRequestName:   exists.PullRequestName,
			AuthorId:          exists.AuthorId,
			Status:            exists.Status,
			AssignedReviewers: assignedIds,
		},
		ReplacedBy: &newAssignedIdPtr,
	}, nil
}
