package postgres

import (
	"context"
	"errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/entity"
	errs "github.com/shirotame/avito-backend-assignment-autumn-2025/internal/errors"
	"github.com/shirotame/avito-backend-assignment-autumn-2025/internal/repository"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type PostgresTeamRepository struct {
	logger *slog.Logger
}

func NewPostgresTeamRepository(baseLogger *slog.Logger) repository.BaseTeamRepository {
	logger := baseLogger.With("module", "teamsrepo")
	return &PostgresTeamRepository{
		logger: logger,
	}
}

func (p *PostgresTeamRepository) GetTeam(
	ctx context.Context,
	db repository.Querier,
	teamName string,
) (*entity.Team, error) {
	query := `
		SELECT name FROM teams
		WHERE name = $1
	`

	var team entity.Team

	err := db.QueryRow(ctx, query, teamName).Scan(&team.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.logger.Debug("failed to GetTeam: not found", "teamName", teamName)
			return nil, errs.ErrNotFound("team", "name", teamName)
		}
		p.logger.Debug("failed to GetTeam", "teamName", teamName, "err", err)
		return nil, errs.ErrInternal("failed to GetTeam", err)
	}
	return &team, nil
}

func (p *PostgresTeamRepository) AddTeam(
	ctx context.Context,
	db repository.Querier,
	new *entity.Team,
) error {
	query := `
		INSERT INTO teams
		(name) VALUES ($1);
	`

	_, err := db.Exec(ctx, query, new.TeamName)
	if err != nil {
		p.logger.Debug("failed to AddTeam", "teamName", new.TeamName, "err", err)
		return errs.ErrInternal("failed to AddTeam", err)
	}
	return nil
}
