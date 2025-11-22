package repository

import (
	"context"
	"log/slog"
	"prservice/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTeamRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresTeamRepository(baseLogger *slog.Logger, pool *pgxpool.Pool) BaseTeamRepository {
	logger := baseLogger.With("module", "teamsrepo")
	return &PostgresTeamRepository{
		pool:   pool,
		logger: logger,
	}
}

func (p *PostgresTeamRepository) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	query := `
		SELECT name FROM teams
		WHERE name = $1
	`

	var team entity.Team

	err := p.pool.QueryRow(ctx, query, teamName).Scan(&team.TeamName)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (p *PostgresTeamRepository) AddTeam(ctx context.Context, new *entity.Team) error {
	query := `
		INSERT INTO teams
		(name) VALUES ($1);
	`

	_, err := p.pool.Exec(ctx, query, new.TeamName)
	if err != nil {
		return err
	}
	return nil
}
