package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"prservice/internal/entity"
	errs "prservice/internal/errors"
	"prservice/internal/repository"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresUserRepository struct {
	logger *slog.Logger
}

func NewPostgresUserRepository(baseLogger *slog.Logger) repository.BaseUserRepository {
	logger := baseLogger.With("module", "userrepo")
	return &PostgresUserRepository{
		logger: logger,
	}
}

func (p *PostgresUserRepository) GetById(ctx context.Context, db repository.Querier, id string) (*entity.User, error) {
	query := `
        SELECT id, username, team_name, is_active
        FROM users
        WHERE id = $1
    `
	var result entity.User

	err := db.QueryRow(ctx, query, id).Scan(
		&result.Id,
		&result.Username,
		&result.TeamName,
		&result.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.logger.Debug("failed to GetById: not found", "id", id)
			return nil, errs.ErrNotFound("user", "id", id)
		}
		p.logger.Debug("failed to GetById", "id", id, "error", err)
		return nil, errs.ErrInternal("failed to GetById", err)
	}
	return &result, nil
}

func (p *PostgresUserRepository) GetByTeamName(ctx context.Context, db repository.Querier, teamName string) ([]entity.User, error) {
	query := `
        SELECT id, username, team_name, is_active
        FROM users
        WHERE team_name = $1
    `
	var result []entity.User

	rows, err := db.Query(ctx, query, teamName)
	if err != nil {
		p.logger.Debug("failed to GetByTeamName", "teamName", teamName, "error", err)
		return nil, errs.ErrInternal("failed to GetByTeamName", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User

		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.TeamName,
			&user.IsActive)
		if err != nil {
			p.logger.Debug("failed to GetByTeamName: scan error", "teamName", teamName, "error", err)
			return nil, errs.ErrInternal("failed to GetByTeamName: scan error", err)
		}
		result = append(result, user)
	}
	return result, nil
}

func (p *PostgresUserRepository) GetActiveByTeamName(ctx context.Context, db repository.Querier, teamName string) ([]entity.User, error) {
	query := `
        SELECT id, username, team_name, is_active
        FROM users
        WHERE team_name = $1 AND is_active = true
    `
	var result []entity.User

	rows, err := db.Query(ctx, query, teamName)
	if err != nil {
		p.logger.Debug("failed to GetActiveByTeamName", "teamName", teamName, "error", err)
		return nil, errs.ErrInternal("failed to GetActiveByTeamName", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User

		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.TeamName,
			&user.IsActive)
		if err != nil {
			p.logger.Debug("failed to GetActiveByTeamName: scan error", "teamName", teamName, "error", err)
			return nil, errs.ErrInternal("failed to GetActiveByTeamName: scan error", err)
		}
		result = append(result, user)
	}
	return result, nil
}

func (p *PostgresUserRepository) GetReviewersByPrId(ctx context.Context, db repository.Querier, prId string) ([]entity.User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users u
		JOIN pull_requests_users pr_u ON u.id = pr_u.user_id
		WHERE pr_u.pr_id = $1
	`
	var result []entity.User
	rows, err := db.Query(ctx, query, prId)
	if err != nil {
		p.logger.Debug("failed to GetReviewersByPrId", "prId", prId, "error", err)
		return nil, errs.ErrInternal("failed to GetReviewersByPrId", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.Id,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
		)
		if err != nil {
			p.logger.Debug("failed to GetReviewersByPrId", "prId", prId, "error", err)
			return nil, errs.ErrInternal("failed to GetReviewersByPrId", err)
		}
		result = append(result, user)
	}
	return result, nil
}

func (p *PostgresUserRepository) AddUsers(ctx context.Context, db repository.Querier, new []entity.User) error {
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
	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				p.logger.Debug("failed to AddUsers: some of users already exists", "query", query, "args", args, "error", err)
				return errs.ErrUserAlreadyExists
			}
		}
		p.logger.Debug("failed to AddUsers", "query", query, "args", args, "error", err)
		return errs.ErrInternal("failed to AddUsers", err)
	}
	return nil
}

func (p *PostgresUserRepository) UpdateUser(ctx context.Context, db repository.Querier, userId string, update *entity.UserUpdate) error {
	if update.Username == nil && update.TeamName == nil && update.IsActive == nil {
		return errs.ErrBadFilter("Username or TeamName or IsActive is required")
	}

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

	ct, err := db.Exec(ctx, query, args...)
	if err != nil {
		p.logger.Debug("failed to UpdateUser", "query", query, "args", args, "error", err)
		return errs.ErrInternal("failed to UpdateUser", err)
	}
	if ct.RowsAffected() == 0 {
		p.logger.Debug("failed to UpdateUser: not found", "userId", userId)
		return errs.ErrNotFound("user", "id", userId)
	}
	return nil
}
