package entity

type UserUpdate struct {
	Username *string
	TeamName *string
	IsActive *bool
}

type PullRequestUpdate struct {
	PullRequestName *string
	AuthorId        *string
	Status          *string
}
