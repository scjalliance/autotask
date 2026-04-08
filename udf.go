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
//
// Deprecated: Use the Time type for JSON marshaling instead.
func TimeToString(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// StringToTime parses an ISO 8601 string from the API into a time.Time.
// It handles both the millisecond format ("2006-01-02T15:04:05.000Z") and
// the standard RFC3339 format ("2006-01-02T15:04:05Z").
//
// Deprecated: Use the Time type for JSON unmarshaling instead.
func StringToTime(s string) (time.Time, error) {
	for _, f := range []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	} {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Parse(time.RFC3339, s)
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
