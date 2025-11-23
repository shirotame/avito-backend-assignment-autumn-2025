package repository

import (
	"context"
	"prservice/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type BaseUserRepository interface {
	GetById(ctx context.Context, db Querier, id string) (*entity.User, error)
	GetByTeamName(ctx context.Context, db Querier, teamName string) ([]entity.User, error)
	GetActiveByTeamName(ctx context.Context, db Querier, teamName string) ([]entity.User, error)
	AddUsers(ctx context.Context, db Querier, new []entity.User) error
	UpdateUser(ctx context.Context, db Querier, userId string, update *entity.UserUpdate) error
}

type BaseTeamRepository interface {
	GetTeam(ctx context.Context, db Querier, teamName string) (*entity.Team, error)
	AddTeam(ctx context.Context, db Querier, new *entity.Team) error
}

type BasePullRequestRepository interface {
	GetPullRequestsByReviewerId(ctx context.Context, db Querier, reviewerId string) ([]entity.PullRequest, error)
	GetPullRequestById(ctx context.Context, db Querier, prId string) (*entity.PullRequest, error)

	AddPullRequest(ctx context.Context, db Querier, ent *entity.PullRequest) error
	UpdatePullRequestStatus(ctx context.Context, db Querier, prId string, newStatus string) error

	AddReviewerToPullRequest(ctx context.Context, db Querier, prId string, reviewerId string) error
	RemoveReviewerFromPullRequest(ctx context.Context, db Querier, prId string, reviewerId string) error
}
