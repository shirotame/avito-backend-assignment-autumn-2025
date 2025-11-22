package repository

import (
	"context"
	"prservice/internal/entity"
)

type BaseUserRepository interface {
	GetById(ctx context.Context, id string) (*entity.User, error)
	GetByTeamName(ctx context.Context, teamName string) ([]entity.User, error)
	GetActiveByTeamName(ctx context.Context, teamName string) ([]entity.User, error)
	AddUsers(ctx context.Context, new []entity.User) error
	UpdateUser(ctx context.Context, userId string, update *entity.UserUpdate) error
}

type BaseTeamRepository interface {
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
	AddTeam(ctx context.Context, new *entity.Team) error
}

type BasePullRequestRepository interface {
	GetPullRequestsByAuthorId(ctx context.Context, authorId string) ([]entity.PullRequest, error)
	GetPullRequestById(ctx context.Context, prId string) (*entity.PullRequest, error)

	AddPullRequest(ctx context.Context, ent *entity.PullRequest) error

	AddReviewerToPullRequest(ctx context.Context, prId string, reviewerId string) error
	RemoveReviewerFromPullRequest(ctx context.Context, prId string, reviewerId string) error
}
