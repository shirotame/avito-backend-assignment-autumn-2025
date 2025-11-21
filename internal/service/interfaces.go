package service

import (
	"context"
	"prservice/internal/entity"
)

type BaseUserService interface {
	SetIsActive(ctx context.Context, dto entity.SetUserIsActiveDTO) (*entity.UserDTO, error)
	GetReview(ctx context.Context, userId string) (entity.UserPullRequestsDTO, error)
}

type BaseTeamService interface {
	AddTeam(ctx context.Context, dto entity.TeamDTO) (entity.ResponseTeamDTO, error)
	GetTeam(ctx context.Context, teamName string) (entity.TeamDTO, error)
}

type BasePullRequestService interface {
	CreatePullRequest(ctx context.Context, dto entity.PullRequestCreateDTO) (entity.PullRequestResponseDTO, error)
	MergePullRequest(ctx context.Context, dto entity.MergePullRequestDTO) (entity.PullRequestResponseDTO, error)
	ReassignPullRequest(ctx context.Context, dto entity.ReassignPullRequestDTO) (entity.PullRequestResponseDTO, error)
}
