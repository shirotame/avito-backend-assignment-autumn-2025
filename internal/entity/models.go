package entity

import "time"

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

type User struct {
	Id       string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	TeamName string
}

type PullRequestUser struct {
	UserId        string
	PullRequestId string
}

type PullRequest struct {
	Id              string
	PullRequestName string
	AuthorId        string
	Status          string
	UpdatedAt       *time.Time
}
