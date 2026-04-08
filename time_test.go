package autotask

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTime_MarshalJSON(t *testing.T) {
	ts := Time{time.Date(2025, 10, 29, 14, 53, 30, 0, time.UTC)}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	want := `"2025-10-29T14:53:30.000Z"`
	if string(data) != want {
		t.Errorf("got %s, want %s", string(data), want)
	}
}

func TestTime_MarshalJSON_Zero(t *testing.T) {
	ts := Time{}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	want := `""`
	if string(data) != want {
		t.Errorf("got %s, want %s", string(data), want)
	}
}

func TestTime_RoundTrip_Empty(t *testing.T) {
	// Empty string should round-trip: unmarshal "" → zero → marshal ""
	var ts Time
	if err := json.Unmarshal([]byte(`""`), &ts); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `""` {
		t.Errorf("empty string round-trip failed: got %s", string(data))
	}
}

func TestTime_UnmarshalJSON_WithMillis(t *testing.T) {
	var ts Time
	if err := json.Unmarshal([]byte(`"2025-10-29T14:53:30.000Z"`), &ts); err != nil {
		t.Fatal(err)
	}
	if ts.Year() != 2025 || ts.Month() != 10 || ts.Day() != 29 {
		t.Errorf("unexpected: %v", ts.Time)
	}
	if ts.Hour() != 14 || ts.Minute() != 53 || ts.Second() != 30 {
		t.Errorf("unexpected time: %v", ts.Time)
	}
}

func TestTime_UnmarshalJSON_WithoutMillis(t *testing.T) {
	var ts Time
	if err := json.Unmarshal([]byte(`"2025-10-29T14:53:30Z"`), &ts); err != nil {
		t.Fatal(err)
	}
	if ts.Year() != 2025 || ts.Second() != 30 {
		t.Errorf("unexpected: %v", ts.Time)
	}
}

func TestTime_UnmarshalJSON_Empty(t *testing.T) {
	var ts Time
	if err := json.Unmarshal([]byte(`""`), &ts); err != nil {
		t.Fatal(err)
	}
	if !ts.IsZero() {
		t.Errorf("expected zero time, got %v", ts.Time)
	}
}

func TestTime_RoundTrip(t *testing.T) {
	original := Time{time.Date(2026, 3, 26, 14, 30, 0, 0, time.UTC)}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var parsed Time
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if !original.Equal(parsed.Time) {
		t.Errorf("round trip failed: %v != %v", original.Time, parsed.Time)
	}
}

func TestTime_NilPointer_Omitempty(t *testing.T) {
	type s struct {
		T *Time `json:"t,omitempty"`
	}

	// nil → omitted
	data, _ := json.Marshal(s{T: nil})
	if string(data) != `{}` {
		t.Errorf("nil *Time should be omitted, got %s", string(data))
	}

	// non-nil → included
	data, _ = json.Marshal(s{T: &Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}})
	if string(data) == `{}` {
		t.Error("non-nil *Time should not be omitted")
	}
}

func TestTime_String(t *testing.T) {
	ts := Time{time.Date(2025, 10, 29, 14, 53, 30, 0, time.UTC)}
	want := "2025-10-29T14:53:30.000Z"
	if got := ts.String(); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestTime_String_Zero(t *testing.T) {
	ts := Time{}
	if got := ts.String(); got != "" {
		t.Errorf("zero time String() should be empty, got %s", got)
	}
}

func TestTime_Ptr(t *testing.T) {
	p := Ptr(Time{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if p.Year() != 2026 {
		t.Errorf("unexpected year: %d", p.Year())
	}
}

func TestTime_InFilterCondition(t *testing.T) {
	ts := Time{time.Date(2025, 10, 29, 0, 0, 0, 0, time.UTC)}
	f := Filter(Field("createDate").Gte(ts))
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	// The Time value should marshal to a string inside the filter JSON.
	want := `{"filter":[{"op":"gte","field":"createDate","value":"2025-10-29T00:00:00.000Z"}]}`
	if string(data) != want {
		t.Errorf("got %s", string(data))
	}
}
