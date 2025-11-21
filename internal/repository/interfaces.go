package repository

import (
	"context"
	"prservice/internal/entity"
)

type BaseUserRepository interface {
	GetUser(ctx context.Context, userId string) (*entity.User, error)
	GetUsers(ctx context.Context, teamName string) ([]entity.User, error)
	AddUsers(ctx context.Context, new []entity.User) error
	UpdateUser(ctx context.Context, userId string, update *entity.UserUpdate) error
}

type BaseTeamRepository interface {
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
	AddTeam(ctx context.Context, new *entity.Team) error
}

type BasePullRequestRepository interface {
	GetPullRequests(ctx context.Context, prId string) ([]entity.PullRequest, error)
	GetPullRequest(ctx context.Context, prId string) (*entity.PullRequest, error)
	AddPullRequest(ctx context.Context, ent entity.PullRequest) error
	UpdatePullRequest(ctx context.Context, prId string, update *entity.PullRequestUpdate) error
}

type BasePullRequestUserUserRepository interface {
	AddPullRequestUser(ctx context.Context, userId string, prId string) (*entity.PullRequestUser, error)
	DeletePullRequestUser(ctx context.Context, userId string, prId string) error
}
