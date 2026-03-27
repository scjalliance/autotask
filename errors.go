package autotask

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound     = errors.New("autotask: not found")
	ErrUnauthorized = errors.New("autotask: unauthorized")
	ErrForbidden    = errors.New("autotask: forbidden")
	ErrRateLimited  = errors.New("autotask: rate limited")
)

var sentinelByStatus = map[int]error{
	404: ErrNotFound,
	401: ErrUnauthorized,
	403: ErrForbidden,
	429: ErrRateLimited,
}

type APIError struct {
	StatusCode int
	Errors     []string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("autotask: HTTP %d: %s", e.StatusCode, strings.Join(e.Errors, "; "))
}

func (e *APIError) Is(target error) bool {
	if sentinel, ok := sentinelByStatus[e.StatusCode]; ok {
		return errors.Is(sentinel, target)
	}
	return false
}
