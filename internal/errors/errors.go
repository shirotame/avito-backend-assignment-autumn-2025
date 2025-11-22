package errors

import (
	"errors"
	"fmt"
)

var ErrBaseNotFound = errors.New("not found")
var ErrBaseBadFilter = errors.New("bad filter")

func ErrNotFound(entity string, param string, value any) error {
	return fmt.Errorf("%s with %s: %v %w. ", entity, param, value, ErrBaseNotFound)
}

func ErrBadFilter(msg string) error {
	return fmt.Errorf("%w, message: %v", ErrBaseBadFilter, msg)
}
