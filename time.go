package autotask

import (
	"encoding/json"
	"fmt"
	"time"
)

// Time wraps time.Time to marshal and unmarshal as the ISO 8601 string format
// used by the Autotask API. Use *Time in struct fields for omitempty and PATCH
// semantics: nil means the field is absent, while a non-nil pointer (even to
// the zero time) is included in the JSON output.
type Time struct {
	time.Time
}

// timeFormat is the primary format used by the Autotask API.
const timeFormat = "2006-01-02T15:04:05.000Z"

// parseFormats lists all timestamp formats the API might return.
var parseFormats = []string{
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.999999999Z",
	time.RFC3339,
	time.RFC3339Nano,
}

// MarshalJSON outputs the time as a UTC ISO 8601 string with milliseconds,
// matching the format the Autotask API returns. The zero time marshals as
// an empty string to preserve round-trip fidelity with blank API fields.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(t.UTC().Format(timeFormat))
}

// UnmarshalJSON parses the Autotask API's ISO 8601 timestamp variants.
// An empty string is treated as the zero time.
func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	var parseErr error
	for _, f := range parseFormats {
		parsed, err := time.Parse(f, s)
		if err == nil {
			t.Time = parsed
			return nil
		}
		parseErr = err
	}
	return fmt.Errorf("autotask: parsing time %q: %w", s, parseErr)
}

// String returns the time in the Autotask API format, or an empty string for
// the zero time.
func (t Time) String() string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(timeFormat)
}
