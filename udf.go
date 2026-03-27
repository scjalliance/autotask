package autotask

import "time"

// UDF represents a user-defined field value.
type UDF struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// PatchData is a map of field names to values for partial updates.
type PatchData map[string]any

// TimeToString converts a time.Time to the ISO 8601 string format used by the API.
func TimeToString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// StringToTime parses an ISO 8601 string from the API into a time.Time.
func StringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
