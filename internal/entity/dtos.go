package entity

type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UserDTO struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type UserPullRequestsDTO struct {
	UserId       string           `json:"user_id"`
	PullRequests []PullRequestDTO `json:"pull_requests"`
}

type SetUserIsActiveDTO struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string    `json:"team_name"`
	Members  []UserDTO `json:"members"`
}

type ResponseTeamDTO struct {
	Team TeamDTO `json:"team"`
}

type PullRequestCreateDTO struct {
	PullRequestId   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId        string `json:"author_id"`
}

type MergePullRequestDTO struct {
	PullRequestId string `json:"pull_request_id"`
}

type ReassignPullRequestDTO struct {
	PullRequestId string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
}

type PullRequestDTO struct {
	PullRequestId     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorId          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

type PullRequestResponseDTO struct {
	PullRequest PullRequestDTO `json:"pr"`
	ReplacedBy  *string        `json:"replaced_by,omitempty"`
}
