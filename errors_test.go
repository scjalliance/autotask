package autotask

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 400, Errors: []string{"field required", "invalid value"}}
	got := err.Error()
	if got != "autotask: HTTP 400: field required; invalid value" {
		t.Errorf("unexpected error string: %s", got)
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		target error
		want   bool
	}{
		{"404 matches ErrNotFound", &APIError{StatusCode: 404}, ErrNotFound, true},
		{"401 matches ErrUnauthorized", &APIError{StatusCode: 401}, ErrUnauthorized, true},
		{"403 matches ErrForbidden", &APIError{StatusCode: 403}, ErrForbidden, true},
		{"429 matches ErrRateLimited", &APIError{StatusCode: 429}, ErrRateLimited, true},
		{"400 does not match ErrNotFound", &APIError{StatusCode: 400}, ErrNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_As(t *testing.T) {
	orig := &APIError{StatusCode: 500, Errors: []string{"internal"}}
	var wrapped error = orig
	var target *APIError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should succeed")
	}
	if target.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", target.StatusCode)
	}
}
