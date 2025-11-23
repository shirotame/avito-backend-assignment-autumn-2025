package errors

import (
	"errors"
	"fmt"
)

var ErrBaseInternal = errors.New("internal error")
var ErrBaseNotFound = errors.New("not found")
var ErrBaseBadFilter = errors.New("bad filter")
var ErrBaseBadRequest = errors.New("bad request")
var ErrBaseAlreadyExists = errors.New("already exists")

var ErrUserNotAssigned = errors.New("reviewer is not assigned to this PR")
var ErrNoActiveUsers = errors.New("no active replacement candidate in team")
var ErrReassignOnMergedPR = errors.New("cannot reassign on merged PR")

var ErrTeamAlreadyExists = fmt.Errorf("team %w", ErrBaseAlreadyExists)
var ErrPullRequestAlreadyExists = fmt.Errorf("pull request %w", ErrBaseAlreadyExists)

func ErrNotFound(entity string, param string, value any) error {
	return fmt.Errorf("%s with %s: %v %w", entity, param, value, ErrBaseNotFound)
}

func ErrBadFilter(msg string) error {
	return fmt.Errorf("%w, message: %v", ErrBaseBadFilter, msg)
}

func ErrInternal(data string, errorWrap error) error {
	if errorWrap == nil {
		return fmt.Errorf("%w: %s", ErrBaseInternal, data)
	}
	return fmt.Errorf("%w: %s. error: %w", ErrBaseInternal, data, errorWrap)
}
