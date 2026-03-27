package autotask

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUDF_JSON_RoundTrip(t *testing.T) {
	udfs := []UDF{
		{Name: "Priority", Value: "High"},
		{Name: "Count", Value: float64(42)},
	}
	data, err := json.Marshal(udfs)
	if err != nil {
		t.Fatal(err)
	}
	var got []UDF
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Priority" || got[0].Value != "High" {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestTimeToString(t *testing.T) {
	ts := time.Date(2026, 3, 26, 14, 30, 0, 0, time.UTC)
	got := TimeToString(ts)
	if got != "2026-03-26T14:30:00Z" {
		t.Errorf("got %s", got)
	}
}

func TestStringToTime(t *testing.T) {
	ts, err := StringToTime("2026-03-26T14:30:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if ts.Year() != 2026 || ts.Month() != 3 || ts.Day() != 26 {
		t.Errorf("unexpected: %v", ts)
	}
}

func TestPtr(t *testing.T) {
	s := Ptr("hello")
	if *s != "hello" {
		t.Errorf("got %s", *s)
	}
	n := Ptr(int64(42))
	if *n != 42 {
		t.Errorf("got %d", *n)
	}
}
