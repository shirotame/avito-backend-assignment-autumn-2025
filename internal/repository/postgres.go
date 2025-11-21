package repository

import (
	"context"
	"fmt"
	"log/slog"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresUserRepository(baseLogger *slog.Logger, pool *pgxpool.Pool) BaseUserRepository {
	logger := baseLogger.With("module", "user_repository")
	return &PostgresUserRepository{
		pool:   pool,
		logger: logger,
	}
}

func (p *PostgresUserRepository) GetUser(ctx context.Context, userId string) (*entity.User, error) {
	query := `
        SELECT id, username, team_name, is_active
        FROM users
        WHERE id = $1
    `
	var result entity.User

	err := p.pool.QueryRow(ctx, query, userId).Scan(&result)
	if err != nil {
		p.logger.Debug("failed to GetUser", "userId", userId, "error", err)
		return nil, err
	}
	return &result, nil
}

func (p *PostgresUserRepository) GetUsers(ctx context.Context, teamName string) ([]entity.User, error) {
	query := `
        SELECT id, username, team_name, is_active
        FROM users
        WHERE team_name = $1
    `
	var result []entity.User

	err := p.pool.QueryRow(ctx, query, teamName).Scan(&result)
	if err != nil {
		p.logger.Debug("failed to GetUsers", "teamName", teamName, "error", err)
		return nil, err
	}
	return result, nil
}

func (p *PostgresUserRepository) AddUsers(ctx context.Context, new []entity.User) error {
	query := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES 
	`

	values := make([]string, len(new))
	args := make([]interface{}, len(new)*4)
	currIdx := 0
	for i := 0; i < len(new)*4; i += 4 {
		values[currIdx] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i+1, i+2, i+3, i+4)

		args[i] = new[currIdx].Id
		args[i+1] = new[currIdx].Username
		args[i+2] = new[currIdx].TeamName
		args[i+3] = new[currIdx].IsActive

		currIdx++
	}
	query = fmt.Sprintf("%s %s", query, strings.Join(values, ", "))
	ct, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		p.logger.Debug("failed to AddUsers", "query", query, "args", args, "error", err)
		return err
	}
	if ct.RowsAffected() != int64(len(new)) {
		p.logger.Debug("failed to AddUsers: RowsAffected less than excepted", "RowsAffected", ct.RowsAffected(), "expected", int64(len(new)), "query", query, "args", args)
	}
	return nil
}

func (p *PostgresUserRepository) UpdateUser(ctx context.Context, userId string, update *entity.UserUpdate) error {
	query := `
		UPDATE users
		SET 
	`

	currUpdate := 1
	values := make([]string, 0)
	args := make([]interface{}, 0)
	if update.Username != nil {
		values = append(values, fmt.Sprintf("username = $%d", currUpdate))
		args = append(args, *update.Username)
		currUpdate++
	}
	if update.TeamName != nil {
		values = append(values, fmt.Sprintf("team_name = $%d", currUpdate))
		args = append(args, *update.TeamName)
		currUpdate++
	}
	if update.IsActive != nil {
		values = append(values, fmt.Sprintf("is_active = $%d", currUpdate))
		args = append(args, *update.IsActive)
		currUpdate++
	}
	query = fmt.Sprintf("%s %s %s", query, strings.Join(values, ", "), fmt.Sprintf("WHERE id = $%d", currUpdate))
	args = append(args, userId)

	ct, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		p.logger.Debug("failed to UpdateUser", "query", query, "args", args, "error", err)
		return err
	}
	if ct.RowsAffected() == 0 {
		p.logger.Debug("failed to UpdateUser: not found", "userId", userId)
		return errs.ErrNotFound("user", "id", userId)
	}
	return nil
}
