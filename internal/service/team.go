package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamService struct {
	logger   *slog.Logger
	pool     *pgxpool.Pool
	teamRepo repository.BaseTeamRepository
	userRepo repository.BaseUserRepository
}

func NewTeamService(
	baseLogger *slog.Logger,
	pool *pgxpool.Pool,
	userRepo repository.BaseUserRepository,
	teamRepo repository.BaseTeamRepository,
) BaseTeamService {
	logger := baseLogger.With("module", "teamservice")
	return &TeamService{
		logger:   logger,
		pool:     pool,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *TeamService) AddTeam(
	ctx context.Context,
	dto entity.TeamDTO,
) (*entity.ResponseTeamDTO, error) {
	exists, err := s.teamRepo.GetTeam(ctx, s.pool, dto.TeamName)
	if err != nil && !errors.Is(err, errs.ErrBaseNotFound) {
		s.logger.Debug("failed to AddTeam: error in GetTeam", "dto", dto, "err", err)
		return nil, err
	}
	if exists != nil {
		s.logger.Debug(
			"failed GetTeam: team with this name already exists",
			"dto.teamName",
			dto.TeamName,
		)
		return nil, errs.ErrTeamAlreadyExists
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errs.ErrInternal("error begin transaction", err)
	}
	defer tx.Rollback(ctx)

	err = s.teamRepo.AddTeam(ctx, tx, &entity.Team{TeamName: dto.TeamName})
	if err != nil {
		s.logger.Debug("failed to AddTeam: error in AddTeam", "dto", dto, "err", err)
		return nil, err
	}

	users := make([]entity.User, len(dto.Members))
	for i, user := range dto.Members {
		users[i] = entity.User{
			Id:       user.UserId,
			Username: user.Username,
			TeamName: dto.TeamName,
			IsActive: user.IsActive,
		}
	}
	err = s.userRepo.AddUsers(ctx, tx, users)
	if err != nil {
		s.logger.Debug("failed to AddTeam: error in AddUsers", "dto", dto, "err", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		s.logger.Debug("failed to AddTeam: failed commiting result", "dto", dto, "err", err)
		return nil, err
	}
	return &entity.ResponseTeamDTO{
		Team: entity.TeamDTO{
			TeamName: dto.TeamName,
			Members:  dto.Members,
		},
	}, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*entity.TeamDTO, error) {
	exists, err := s.teamRepo.GetTeam(ctx, s.pool, teamName)
	if err != nil {
		s.logger.Debug("failed to GetTeam: error in GetTeam", "teamName", teamName, "err", err)
		return nil, err
	}

	users, err := s.userRepo.GetByTeamName(ctx, s.pool, teamName)
	if err != nil {
		s.logger.Debug(
			"failed to GetTeam: error in GetByTeamName",
			"teamName",
			teamName,
			"err",
			err,
		)
		return nil, err
	}
	usersDTO := make([]entity.UserDTO, len(users))
	for i, user := range users {
		usersDTO[i] = entity.UserDTO{
			UserId:   user.Id,
			Username: user.Username,
			IsActive: user.IsActive,
		}
	}

	return &entity.TeamDTO{
		TeamName: exists.TeamName,
		Members:  usersDTO,
	}, nil
}
