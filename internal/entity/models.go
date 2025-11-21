package entity

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

type User struct {
	Id       string `db:"id"`
	Username string `db:"username"`
	TeamName string `db:"team_name"`
	IsActive bool   `db:"is_active"`
}

type Team struct {
	TeamName string `db:"team_name"`
}

type PullRequestUser struct {
	UserId        string `db:"user_id"`
	PullRequestId string `db:"pr_id"`
}

type PullRequest struct {
	Id              string `db:"id"`
	PullRequestName string `db:"pull_request_name"`
	AuthorId        string `db:"author_id"`
	Status          string `db:"status"`
}
