# Autotask Go Client Library Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a comprehensive, code-generated Go client library for the Datto Autotask PSA REST API with type-safe generic CRUD operations.

**Architecture:** Single `package autotask` using Go generics for capability traits (`Reader[T]`, `Creator[T]`, etc.) composed into entity-specific service types. A code generator reads the Swagger 2.0 spec and produces 245 entity structs and their service wiring. Auto-pagination, query builder DSL, and rate limit tracking built into the client core.

**Tech Stack:** Go 1.26, generics, `iter.Seq2` range-over-func, `net/http`, `encoding/json`, `httptest`

**Spec:** `docs/superpowers/specs/2026-03-26-autotask-go-client-design.md`

**Swagger spec:** `~/.cache/api-explorer/apis/autotask/raw/20260326T0000Z/swagger-apisguru.json`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `errors.go`
- Create: `errors_test.go`

- [ ] **Step 1: Initialize git repo and Go module**

Note: the git repo and initial commit already exist from the spec phase. Initialize the Go module:

```bash
cd /home/emmaly/Projects/autotask
go mod init github.com/scjalliance/autotask
```

- [ ] **Step 2: Write the error types test**

Create `errors_test.go`:

```go
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
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./... -run TestAPIError -v
```

Expected: compilation failure — types not defined yet.

- [ ] **Step 4: Implement error types**

Create `errors.go`:

```go
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

// sentinelByStatus maps HTTP status codes to sentinel errors.
var sentinelByStatus = map[int]error{
	404: ErrNotFound,
	401: ErrUnauthorized,
	403: ErrForbidden,
	429: ErrRateLimited,
}

// APIError represents an error response from the Autotask API.
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
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./... -run TestAPIError -v
```

Expected: all 3 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod errors.go errors_test.go
git commit -m "feat: project scaffolding with error types"
```

---

### Task 2: UDF Type and Time Helpers

**Files:**
- Create: `udf.go`
- Create: `udf_test.go`

- [ ] **Step 1: Write the UDF and time helper tests**

Create `udf_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run "TestUDF|TestTime|TestPtr" -v
```

Expected: compilation failure.

- [ ] **Step 3: Implement UDF type and helpers**

Create `udf.go`:

```go
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

// Ptr returns a pointer to the given value. Useful for constructing entities
// with pointer fields.
func Ptr[T any](v T) *T {
	return &v
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -run "TestUDF|TestTime|TestPtr" -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add udf.go udf_test.go
git commit -m "feat: UDF type, time helpers, and Ptr generic helper"
```

---

### Task 3: Query Builder

**Files:**
- Create: `query.go`
- Create: `query_test.go`

- [ ] **Step 1: Write query builder tests**

Create `query_test.go`:

```go
package autotask

import (
	"encoding/json"
	"testing"
)

func TestFilter_SimpleEq(t *testing.T) {
	f := Filter(Field("status").Eq(1))
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"filter":[{"op":"eq","field":"status","value":1}]}`
	if string(data) != want {
		t.Errorf("got %s, want %s", string(data), want)
	}
}

func TestFilter_MultipleConditions_ImplicitAnd(t *testing.T) {
	f := Filter(
		Field("status").Eq(1),
		Field("companyID").Eq(12345),
	)
	data, _ := json.Marshal(f)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	filters := parsed["filter"].([]any)
	if len(filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(filters))
	}
}

func TestFilter_Or(t *testing.T) {
	f := Filter(
		Or(
			Field("priority").Eq(1),
			Field("priority").Eq(2),
		),
	)
	data, _ := json.Marshal(f)
	want := `{"filter":[{"op":"or","items":[{"op":"eq","field":"priority","value":1},{"op":"eq","field":"priority","value":2}]}]}`
	if string(data) != want {
		t.Errorf("got %s", string(data))
	}
}

func TestFilter_AllOperators(t *testing.T) {
	operators := []struct {
		fn   func() FilterCondition
		want string
	}{
		{func() FilterCondition { return Field("x").Eq(1) }, "eq"},
		{func() FilterCondition { return Field("x").NotEq(1) }, "noteq"},
		{func() FilterCondition { return Field("x").Gt(1) }, "gt"},
		{func() FilterCondition { return Field("x").Gte(1) }, "gte"},
		{func() FilterCondition { return Field("x").Lt(1) }, "lt"},
		{func() FilterCondition { return Field("x").Lte(1) }, "lte"},
		{func() FilterCondition { return Field("x").BeginsWith("a") }, "beginsWith"},
		{func() FilterCondition { return Field("x").EndsWith("z") }, "endsWith"},
		{func() FilterCondition { return Field("x").Contains("mid") }, "contains"},
		{func() FilterCondition { return Field("x").Exist() }, "exist"},
		{func() FilterCondition { return Field("x").NotExist() }, "notExist"},
		{func() FilterCondition { return Field("x").In([]any{1, 2, 3}) }, "in"},
	}
	for _, tt := range operators {
		f := Filter(tt.fn())
		data, _ := json.Marshal(f)
		var parsed map[string]any
		json.Unmarshal(data, &parsed)
		filters := parsed["filter"].([]any)
		cond := filters[0].(map[string]any)
		if cond["op"] != tt.want {
			t.Errorf("expected op=%s, got %s", tt.want, cond["op"])
		}
	}
}

func TestFilter_UDF(t *testing.T) {
	f := Filter(UDFField("CustomerRanking").Eq("Golden"))
	data, _ := json.Marshal(f)
	want := `{"filter":[{"op":"eq","field":"CustomerRanking","udf":true,"value":"Golden"}]}`
	if string(data) != want {
		t.Errorf("got %s", string(data))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run TestFilter -v
```

Expected: compilation failure.

- [ ] **Step 3: Implement query builder**

Create `query.go`:

```go
package autotask

// FilterCondition is a single filter expression that serializes to the
// Autotask query JSON format.
type FilterCondition struct {
	Op    string           `json:"op"`
	Field string           `json:"field,omitempty"`
	Value any              `json:"value,omitempty"`
	UDF   *bool            `json:"udf,omitempty"`
	Items []FilterCondition `json:"items,omitempty"`
}

// FilterQuery is the top-level query structure sent to the API.
type FilterQuery struct {
	Filter []FilterCondition `json:"filter"`
}

// FilterOption is a functional option for query methods.
type FilterOption func(*FilterQuery)

// Filter creates a FilterOption from one or more conditions.
// Multiple conditions at the top level are implicitly AND'd by the API.
func Filter(conditions ...FilterCondition) FilterOption {
	return func(q *FilterQuery) {
		q.Filter = append(q.Filter, conditions...)
	}
}

// buildFilterQuery applies all FilterOptions and returns the query struct.
func buildFilterQuery(opts []FilterOption) *FilterQuery {
	q := &FilterQuery{}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// FieldSelector allows building filter conditions on a named field.
type FieldSelector struct {
	name string
	udf  bool
}

// Field starts building a filter condition for the named API field.
func Field(name string) FieldSelector {
	return FieldSelector{name: name}
}

// UDFField starts building a filter condition for a user-defined field.
func UDFField(name string) FieldSelector {
	return FieldSelector{name: name, udf: true}
}

func (f FieldSelector) cond(op string, value ...any) FilterCondition {
	c := FilterCondition{Op: op, Field: f.name}
	if len(value) > 0 {
		c.Value = value[0]
	}
	if f.udf {
		b := true
		c.UDF = &b
	}
	return c
}

func (f FieldSelector) Eq(v any) FilterCondition       { return f.cond("eq", v) }
func (f FieldSelector) NotEq(v any) FilterCondition     { return f.cond("noteq", v) }
func (f FieldSelector) Gt(v any) FilterCondition        { return f.cond("gt", v) }
func (f FieldSelector) Gte(v any) FilterCondition       { return f.cond("gte", v) }
func (f FieldSelector) Lt(v any) FilterCondition        { return f.cond("lt", v) }
func (f FieldSelector) Lte(v any) FilterCondition       { return f.cond("lte", v) }
func (f FieldSelector) BeginsWith(v string) FilterCondition { return f.cond("beginsWith", v) }
func (f FieldSelector) EndsWith(v string) FilterCondition   { return f.cond("endsWith", v) }
func (f FieldSelector) Contains(v string) FilterCondition   { return f.cond("contains", v) }
func (f FieldSelector) Exist() FilterCondition              { return f.cond("exist") }
func (f FieldSelector) NotExist() FilterCondition           { return f.cond("notExist") }
func (f FieldSelector) In(v []any) FilterCondition          { return f.cond("in", v) }

// Or groups conditions with OR logic.
func Or(conditions ...FilterCondition) FilterCondition {
	return FilterCondition{Op: "or", Items: conditions}
}

// And groups conditions with AND logic (explicit grouping).
func And(conditions ...FilterCondition) FilterCondition {
	return FilterCondition{Op: "and", Items: conditions}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -run TestFilter -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: query builder with all 12 filter operators"
```

---

### Task 4: Client Core — Config, Auth, HTTP Transport

**Files:**
- Create: `client.go`
- Create: `client_test.go`

- [ ] **Step 1: Write client tests**

Create `client_test.go`:

```go
package autotask

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_RequiredFields(t *testing.T) {
	_, err := NewClient(Config{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestClient_AuthHeaders(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		json.NewEncoder(w).Encode(map[string]any{"item": nil})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		Username:        "user@test.com",
		Secret:          "secret123",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make any request to trigger headers
	c.doGet(context.Background(), "/V1.0/Tickets/1")

	if gotHeaders.Get("UserName") != "user@test.com" {
		t.Errorf("UserName header = %q", gotHeaders.Get("UserName"))
	}
	if gotHeaders.Get("Secret") != "secret123" {
		t.Errorf("Secret header = %q", gotHeaders.Get("Secret"))
	}
	if gotHeaders.Get("ApiIntegrationcode") != "TEST" {
		t.Errorf("ApiIntegrationcode header = %q", gotHeaders.Get("ApiIntegrationcode"))
	}
	if gotHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type header = %q", gotHeaders.Get("Content-Type"))
	}
}

func TestClient_ImpersonationHeader(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		json.NewEncoder(w).Encode(map[string]any{"item": nil})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		Username:        "user@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := WithImpersonation(context.Background(), 12345)
	c.doGet(ctx, "/V1.0/Tickets/1")

	if gotHeaders.Get("ImpersonationResourceId") != "12345" {
		t.Errorf("ImpersonationResourceId = %q", gotHeaders.Get("ImpersonationResourceId"))
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"entity not found"}})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		Username:        "user@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.doGet(context.Background(), "/V1.0/Tickets/999")
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if !errors.Is(err, ErrNotFound) {
		t.Error("expected ErrNotFound")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run "TestNewClient|TestClient_" -v
```

Expected: compilation failure.

- [ ] **Step 3: Implement the client**

Create `client.go`:

```go
package autotask

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds the configuration for an Autotask API client.
type Config struct {
	Username        string
	Secret          string
	IntegrationCode string

	// BaseURL overrides zone auto-detection. Set to the full base URL
	// (e.g., "https://webservices5.autotask.net/ATServicesRest").
	// If empty, the zone is auto-detected on first request.
	BaseURL string

	// HTTPClient is the HTTP client to use. If nil, http.DefaultClient is used.
	HTTPClient *http.Client

	// RateLimitThreshold is the percentage (0-100) at which the client starts
	// voluntary delays. Default is 50. Set to 0 to disable.
	RateLimitThreshold int

	// DisableRateLimitTracking disables all rate limit tracking.
	DisableRateLimitTracking bool
}

// Client is the Autotask REST API client.
type Client struct {
	config  Config
	http    *http.Client
	baseURL string

	zoneMu sync.Mutex

	rateMu       sync.Mutex
	rateWindow   []time.Time
	rateLimit    int
	rateThreshold int
}

// NewClient creates a new Autotask API client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Username == "" {
		return nil, fmt.Errorf("autotask: Username is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("autotask: Secret is required")
	}
	if cfg.IntegrationCode == "" {
		return nil, fmt.Errorf("autotask: IntegrationCode is required")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	threshold := cfg.RateLimitThreshold
	if threshold == 0 && !cfg.DisableRateLimitTracking {
		threshold = 50
	}

	c := &Client{
		config:        cfg,
		http:          httpClient,
		baseURL:       cfg.BaseURL,
		rateLimit:     10000,
		rateThreshold: threshold,
	}

	return c, nil
}

type impersonationKey struct{}

// WithImpersonation adds an impersonation resource ID to the context.
func WithImpersonation(ctx context.Context, resourceID int64) context.Context {
	return context.WithValue(ctx, impersonationKey{}, resourceID)
}

func (c *Client) resolveBaseURL(ctx context.Context) (string, error) {
	if c.baseURL != "" {
		return c.baseURL, nil
	}

	c.zoneMu.Lock()
	defer c.zoneMu.Unlock()

	if c.baseURL != "" {
		return c.baseURL, nil
	}

	// Zone detection: GET https://webservices2.autotask.net/ATServicesRest/V1.0/ZoneInformation?user={username}
	url := fmt.Sprintf("https://webservices2.autotask.net/ATServicesRest/V1.0/ZoneInformation?user=%s", c.config.Username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("autotask: zone detection request: %w", err)
	}
	c.setAuthHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("autotask: zone detection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("autotask: zone detection returned %d", resp.StatusCode)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("autotask: zone detection decode: %w", err)
	}
	if result.URL == "" {
		return "", fmt.Errorf("autotask: zone detection returned empty URL")
	}

	c.baseURL = result.URL
	return c.baseURL, nil
}

func (c *Client) setAuthHeaders(req *http.Request) {
	req.Header.Set("UserName", c.config.Username)
	req.Header.Set("Secret", c.config.Secret)
	req.Header.Set("ApiIntegrationcode", c.config.IntegrationCode)
	req.Header.Set("Content-Type", "application/json")

	if id, ok := req.Context().Value(impersonationKey{}).(int64); ok {
		req.Header.Set("ImpersonationResourceId", strconv.FormatInt(id, 10))
	}
}

func (c *Client) trackRequest() error {
	if c.config.DisableRateLimitTracking || c.rateThreshold == 0 {
		return nil
	}

	c.rateMu.Lock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	// Prune old entries
	valid := c.rateWindow[:0]
	for _, t := range c.rateWindow {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	c.rateWindow = valid

	usage := float64(len(c.rateWindow)) / float64(c.rateLimit) * 100

	if usage >= 90 {
		c.rateMu.Unlock()
		return ErrRateLimited
	}

	c.rateWindow = append(c.rateWindow, now)

	var delay time.Duration
	if usage >= 75 {
		delay = 1 * time.Second
	} else if usage >= float64(c.rateThreshold) {
		delay = 500 * time.Millisecond
	}

	c.rateMu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}
	return nil
}

// doRequest executes an HTTP request with auth headers and error handling.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) ([]byte, error) {
	if err := c.trackRequest(); err != nil {
		return nil, err
	}

	var url string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Absolute URL (e.g., nextPageUrl from pagination) — use as-is
		url = path
	} else {
		base, err := c.resolveBaseURL(ctx)
		if err != nil {
			return nil, err
		}
		url = base + path
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("autotask: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("autotask: create request: %w", err)
	}
	c.setAuthHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("autotask: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("autotask: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var errResp struct {
			Errors []string `json:"errors"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && len(errResp.Errors) > 0 {
			apiErr.Errors = errResp.Errors
		} else {
			apiErr.Errors = []string{string(respBody)}
		}
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) doGet(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

func (c *Client) doPost(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

func (c *Client) doPut(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

func (c *Client) doPatch(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

func (c *Client) doDelete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -run "TestNewClient|TestClient_" -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: client core with auth, zone detection, rate limiting"
```

---

### Task 5: Capability Traits (Reader, Creator, Updater, Patcher, Deleter)

**Files:**
- Create: `traits.go`
- Create: `traits_test.go`

- [ ] **Step 1: Write traits tests**

Create `traits_test.go`:

```go
package autotask

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testEntity struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

func setupTestService[T any](handler http.HandlerFunc) (*Client, *baseService) {
	srv := httptest.NewServer(handler)
	c, _ := NewClient(Config{
		Username:        "test@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
		DisableRateLimitTracking: true,
	})
	base := &baseService{
		client:     c,
		entityPath: "/V1.0/TestEntities",
		entityName: "TestEntity",
	}
	return c, base
}

func TestReader_Get(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/V1.0/TestEntities/42" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"item": map[string]any{"id": 42, "name": "Test"},
		})
	})

	reader := Reader[testEntity]{base: base}
	entity, err := reader.Get(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if *entity.ID != 42 || *entity.Name != "Test" {
		t.Errorf("unexpected entity: %+v", entity)
	}
}

func TestReader_Get_NotFound(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"not found"}})
	})

	reader := Reader[testEntity]{base: base}
	_, err := reader.Get(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestReader_Query(t *testing.T) {
	page := 0
	var srvURL string
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		page++
		// Verify page 1 is POST, page 2 is GET
		if page == 1 && r.Method != http.MethodPost {
			t.Errorf("page 1: expected POST, got %s", r.Method)
		}
		if page == 2 && r.Method != http.MethodGet {
			t.Errorf("page 2: expected GET, got %s", r.Method)
		}
		resp := map[string]any{
			"items": []map[string]any{
				{"id": page, "name": fmt.Sprintf("item%d", page)},
			},
			"pageDetails": map[string]any{
				"count":        1,
				"requestCount": 500,
			},
		}
		if page == 1 {
			// Real API returns absolute URL for nextPageUrl
			resp["pageDetails"].(map[string]any)["nextPageUrl"] = srvURL + r.URL.Path + "?page=2"
		}
		json.NewEncoder(w).Encode(resp)
	})
	srvURL = base.client.baseURL // capture the test server URL

	reader := Reader[testEntity]{base: base}
	results, err := reader.Query(context.Background(), Filter(Field("name").Eq("test")))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (2 pages), got %d", len(results))
	}
}

func TestReader_Count(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/V1.0/TestEntities/query/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"queryCount": 42})
	})

	reader := Reader[testEntity]{base: base}
	count, err := reader.Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

func TestCreator_Create(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body testEntity
		json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(map[string]any{
			"itemId": 99,
		})
	})

	creator := Creator[testEntity]{base: base}
	id, err := creator.Create(context.Background(), &testEntity{Name: Ptr("New")})
	if err != nil {
		t.Fatal(err)
	}
	if id != 99 {
		t.Errorf("expected id 99, got %d", id)
	}
}

func TestUpdater_Update(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"itemId": 42,
		})
	})

	updater := Updater[testEntity]{base: base}
	err := updater.Update(context.Background(), &testEntity{ID: Ptr(int64(42)), Name: Ptr("Updated")})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPatcher_Patch(t *testing.T) {
	var gotBody map[string]any
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities/42" {
			t.Errorf("expected path /V1.0/TestEntities/42, got %s", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]any{"itemId": 42})
	})

	patcher := Patcher[testEntity]{base: base}
	err := patcher.Patch(context.Background(), 42, PatchData{"name": "Patched"})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["name"] != "Patched" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

func TestDeleter_Delete(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/V1.0/TestEntities/42" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{})
	})

	deleter := Deleter[testEntity]{base: base}
	err := deleter.Delete(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run "TestReader|TestCreator|TestUpdater|TestPatcher|TestDeleter" -v
```

Expected: compilation failure.

- [ ] **Step 3: Implement traits**

Create `traits.go`:

```go
package autotask

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
)

// baseService holds shared state for entity service traits.
type baseService struct {
	client     *Client
	entityPath string
	entityName string
}

// Response envelope types
type itemResponse[T any] struct {
	Item *T `json:"item"`
}

type queryResponse[T any] struct {
	Items       []*T        `json:"items"`
	PageDetails pageDetails `json:"pageDetails"`
}

type pageDetails struct {
	Count        int    `json:"count"`
	RequestCount int    `json:"requestCount"`
	PrevPageURL  string `json:"prevPageUrl"`
	NextPageURL  string `json:"nextPageUrl"`
}

type countResponse struct {
	QueryCount int64 `json:"queryCount"`
}

type createResponse[T any] struct {
	ItemID int64 `json:"itemId"`
	Item   *T    `json:"item"`
}

// Reader provides read operations for an entity type.
type Reader[T any] struct{ base *baseService }

func (r Reader[T]) Get(ctx context.Context, id int64) (*T, error) {
	path := fmt.Sprintf("%s/%d", r.base.entityPath, id)
	data, err := r.base.client.doGet(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp itemResponse[T]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("autotask: decode %s: %w", r.base.entityName, err)
	}
	if resp.Item == nil {
		return nil, ErrNotFound
	}
	return resp.Item, nil
}

func (r Reader[T]) Query(ctx context.Context, opts ...FilterOption) ([]*T, error) {
	var all []*T
	for item, err := range r.QueryIter(ctx, opts...) {
		if err != nil {
			return nil, err
		}
		all = append(all, item)
	}
	return all, nil
}

func (r Reader[T]) QueryIter(ctx context.Context, opts ...FilterOption) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		q := buildFilterQuery(opts)

		// First page: POST with the filter query
		data, err := r.base.client.doPost(ctx, r.base.entityPath+"/query", q)
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			var resp queryResponse[T]
			if err := json.Unmarshal(data, &resp); err != nil {
				yield(nil, fmt.Errorf("autotask: decode %s query: %w", r.base.entityName, err))
				return
			}

			for _, item := range resp.Items {
				if !yield(item, nil) {
					return
				}
			}

			if resp.PageDetails.NextPageURL == "" {
				return
			}

			// Subsequent pages: GET with nextPageUrl
			data, err = r.base.client.doGet(ctx, resp.PageDetails.NextPageURL)
			if err != nil {
				yield(nil, err)
				return
			}
		}
	}
}

func (r Reader[T]) Count(ctx context.Context, opts ...FilterOption) (int64, error) {
	q := buildFilterQuery(opts)
	path := r.base.entityPath + "/query/count"
	data, err := r.base.client.doPost(ctx, path, q)
	if err != nil {
		return 0, err
	}
	var resp countResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("autotask: decode %s count: %w", r.base.entityName, err)
	}
	return resp.QueryCount, nil
}

// Creator provides create operations for an entity type.
type Creator[T any] struct{ base *baseService }

func (cr Creator[T]) Create(ctx context.Context, entity *T) (int64, error) {
	data, err := cr.base.client.doPost(ctx, cr.base.entityPath, entity)
	if err != nil {
		return 0, err
	}
	var resp createResponse[T]
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, fmt.Errorf("autotask: decode %s create: %w", cr.base.entityName, err)
	}
	return resp.ItemID, nil
}

// Updater provides full update operations for an entity type.
type Updater[T any] struct{ base *baseService }

func (u Updater[T]) Update(ctx context.Context, entity *T) error {
	_, err := u.base.client.doPut(ctx, u.base.entityPath, entity)
	return err
}

// Patcher provides partial update operations for an entity type.
type Patcher[T any] struct{ base *baseService }

func (p Patcher[T]) Patch(ctx context.Context, id int64, data PatchData) error {
	path := fmt.Sprintf("%s/%d", p.base.entityPath, id)
	_, err := p.base.client.doPatch(ctx, path, data)
	return err
}

// Deleter provides delete operations for an entity type.
type Deleter[T any] struct{ base *baseService }

func (d Deleter[T]) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("%s/%d", d.base.entityPath, id)
	_, err := d.base.client.doDelete(ctx, path)
	return err
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -run "TestReader|TestCreator|TestUpdater|TestPatcher|TestDeleter" -v
```

Expected: all 8 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add traits.go traits_test.go
git commit -m "feat: generic capability traits with pagination"
```

---

### Task 6: Code Generator — Models

**Files:**
- Create: `cmd/generate/main.go`
- Create: `generate.go`

This is the first half of the code generator: reading the swagger spec and producing `gen_models.go`.

- [ ] **Step 1: Create the generate directive**

Create `generate.go`:

```go
package autotask

//go:generate go run ./cmd/generate -spec ~/.cache/api-explorer/apis/autotask/raw/20260326T0000Z/swagger-apisguru.json
```

- [ ] **Step 2: Implement the code generator**

Create `cmd/generate/main.go`. This is a substantial file — it reads the swagger spec and generates Go source. The key logic:

1. Parse swagger JSON, extract all definitions ending in `Model`
2. For each model: convert fields to Go types, generate struct with JSON tags
3. Map swagger types → Go types: `integer` → `*int64`, `number` → `*float64`, `string` → `*string`, `boolean` → `*bool`
4. Write generated file with `// Code generated` header

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type SwaggerSpec struct {
	Paths       map[string]map[string]Operation `json:"paths"`
	Definitions map[string]Definition           `json:"definitions"`
}

type Operation struct {
	Tags        []string `json:"tags"`
	OperationID string   `json:"operationId"`
}

type Definition struct {
	Properties map[string]Property `json:"properties"`
	Type       string              `json:"type"`
}

type Property struct {
	Type        string   `json:"type"`
	Format      string   `json:"format"`
	Description string   `json:"description"`
	ReadOnly    bool     `json:"readOnly"`
	Ref         string   `json:"$ref"`
	Items       *RefType `json:"items"`
}

type RefType struct {
	Type string `json:"type"`
	Ref  string `json:"$ref"`
}

func main() {
	specPath := flag.String("spec", "", "path to swagger JSON spec")
	outDir := flag.String("out", ".", "output directory for generated files")
	flag.Parse()

	if *specPath == "" {
		// Try default path
		matches, _ := filepath.Glob(os.ExpandEnv("$HOME/.cache/api-explorer/apis/autotask/raw/*/swagger-apisguru.json"))
		if len(matches) == 0 {
			log.Fatal("no swagger spec found; use -spec flag")
		}
		*specPath = matches[len(matches)-1]
	}

	data, err := os.ReadFile(*specPath)
	if err != nil {
		log.Fatalf("read spec: %v", err)
	}

	var spec SwaggerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		log.Fatalf("parse spec: %v", err)
	}

	if err := generateModels(&spec, *outDir); err != nil {
		log.Fatalf("generate models: %v", err)
	}

	if err := generateServices(&spec, *outDir); err != nil {
		log.Fatalf("generate services: %v", err)
	}

	fmt.Println("code generation complete")
}

func generateModels(spec *SwaggerSpec, outDir string) error {
	var buf strings.Builder
	buf.WriteString("// Code generated by cmd/generate; DO NOT EDIT.\n\n")
	buf.WriteString("package autotask\n\n")

	// Collect and sort model names
	var names []string
	for name := range spec.Definitions {
		if strings.HasSuffix(name, "Model") && !isResponseWrapper(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		def := spec.Definitions[name]
		goName := strings.TrimSuffix(name, "Model")

		buf.WriteString(fmt.Sprintf("// %s represents a Datto Autotask %s entity.\n", goName, goName))
		buf.WriteString(fmt.Sprintf("type %s struct {\n", goName))

		// Sort fields for deterministic output
		var fieldNames []string
		for fn := range def.Properties {
			fieldNames = append(fieldNames, fn)
		}
		sort.Strings(fieldNames)

		for _, fn := range fieldNames {
			prop := def.Properties[fn]
			goFieldName := exportedName(fn)
			goType := swaggerToGoType(prop)
			tag := fmt.Sprintf("`json:\"%s,omitempty\"`", fn)

			comment := ""
			if prop.ReadOnly {
				comment = " // READ-ONLY"
			}

			buf.WriteString(fmt.Sprintf("\t%s %s %s%s\n", goFieldName, goType, tag, comment))
		}

		// Add UserDefinedFields if the entity has UDF endpoints
		if hasUDFs(spec, goName) {
			buf.WriteString(fmt.Sprintf("\tUserDefinedFields []UDF `json:\"userDefinedFields,omitempty\"`\n"))
		}

		buf.WriteString("}\n\n")
	}

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		// Write unformatted for debugging
		os.WriteFile(filepath.Join(outDir, "gen_models.go"), []byte(buf.String()), 0644)
		return fmt.Errorf("format gen_models.go: %w", err)
	}

	return os.WriteFile(filepath.Join(outDir, "gen_models.go"), formatted, 0644)
}

func isResponseWrapper(name string) bool {
	return strings.Contains(name, "QueryActionResult") ||
		strings.Contains(name, "ItemQueryResultModel") ||
		strings.Contains(name, "OperationResultModel") ||
		name == "Object" || name == "Byte" || name == "CollectionItem"
}

func swaggerToGoType(prop Property) string {
	if prop.Ref != "" {
		refName := prop.Ref[strings.LastIndex(prop.Ref, "/")+1:]
		return "*" + strings.TrimSuffix(refName, "Model")
	}

	if prop.Items != nil {
		if prop.Items.Ref != "" {
			refName := prop.Items.Ref[strings.LastIndex(prop.Items.Ref, "/")+1:]
			return "[]" + strings.TrimSuffix(refName, "Model")
		}
		return "[]" + primitiveGoType(prop.Items.Type, "")
	}

	return "*" + primitiveGoType(prop.Type, prop.Format)
}

func primitiveGoType(typ, format string) string {
	switch typ {
	case "integer":
		if format == "int32" {
			return "int32"
		}
		return "int64"
	case "number":
		return "float64"
	case "string":
		return "string"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

func exportedName(s string) string {
	if s == "" {
		return s
	}
	// Handle common abbreviations
	upper := strings.ToUpper(s)
	if upper == "ID" || upper == "URL" || upper == "API" || upper == "UDF" || upper == "IP" || upper == "DNS" || upper == "SSL" || upper == "HTTP" || upper == "RMA" || upper == "SLA" {
		return upper
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	// Fix common suffixes: ...ID, ...URL
	result := string(runes)
	for _, suffix := range []string{"Id", "Url", "Api", "Udf", "Ip", "Dns", "Ssl", "Http", "Rma", "Sla"} {
		upper := strings.ToUpper(suffix)
		if strings.HasSuffix(result, suffix) {
			result = result[:len(result)-len(suffix)] + upper
		}
	}

	return result
}

func hasUDFs(spec *SwaggerSpec, entityName string) bool {
	for _, methods := range spec.Paths {
		for _, op := range methods {
			for _, tag := range op.Tags {
				if tag == entityName+"s" || tag == entityName {
					if strings.Contains(op.OperationID, "UserDefinedFieldDefinitions") {
						return true
					}
				}
			}
		}
	}
	return false
}

func generateServices(spec *SwaggerSpec, outDir string) error {
	// Analyze capabilities per entity tag
	type entityInfo struct {
		tag        string
		goName     string
		modelName  string
		isChild    bool
		parentTag  string
		parentPath string
		childPath  string
		canCreate  bool
		canUpdate  bool
		canPatch   bool
		canDelete  bool
		canQuery   bool
		canGet     bool
	}

	entities := map[string]*entityInfo{}

	for path, methods := range spec.Paths {
		for _, op := range methods {
			for _, tag := range op.Tags {
				if _, ok := entities[tag]; !ok {
					isChild := strings.HasSuffix(tag, "Child")
					goName := tag
					if isChild {
						goName = strings.TrimSuffix(tag, "Child")
					}
					entities[tag] = &entityInfo{
						tag:     tag,
						goName:  goName,
						isChild: isChild,
					}
				}
				e := entities[tag]

				opSuffix := ""
				if idx := strings.LastIndex(op.OperationID, "_"); idx >= 0 {
					opSuffix = op.OperationID[idx+1:]
				}

				switch opSuffix {
				case "QueryItem":
					e.canGet = true
				case "Query", "UrlParameterQuery":
					e.canQuery = true
				case "CreateEntity":
					e.canCreate = true
				case "UpdateEntity":
					e.canUpdate = true
				case "PatchEntity":
					e.canPatch = true
				case "DeleteEntity":
					e.canDelete = true
				}

				// Detect child entity paths
				if e.isChild && strings.Contains(path, "{parentId}") {
					// e.g., /V1.0/Tickets/{parentId}/Notes
					parts := strings.Split(strings.TrimPrefix(path, "/V1.0/"), "/")
					if len(parts) >= 3 {
						e.parentTag = parts[0]
						e.parentPath = "/V1.0/" + parts[0]
						e.childPath = parts[2]
					}
				}
			}
		}
	}

	// Find the model name for each entity
	for _, e := range entities {
		candidate := e.goName + "Model"
		if _, ok := spec.Definitions[candidate]; ok {
			e.modelName = strings.TrimSuffix(candidate, "Model")
		} else {
			// Try singular forms or other patterns
			e.modelName = e.goName
		}
	}

	var buf strings.Builder
	buf.WriteString("// Code generated by cmd/generate; DO NOT EDIT.\n\n")
	buf.WriteString("package autotask\n\n")
	buf.WriteString("import \"fmt\"\n\n")

	// Sort entities for deterministic output
	var tags []string
	for tag := range entities {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	// Generate service types
	var topLevel []*entityInfo
	var children []*entityInfo

	for _, tag := range tags {
		e := entities[tag]
		if strings.Contains(tag, "ApiIntegration") || tag == "MetadataApiIntegration" {
			continue
		}

		// Check that a model definition exists
		modelDef := e.modelName + "Model"
		if _, ok := spec.Definitions[modelDef]; !ok {
			continue
		}

		if e.isChild {
			children = append(children, e)
		} else {
			topLevel = append(topLevel, e)
		}
	}

	// Generate top-level service types
	for _, e := range topLevel {
		buf.WriteString(fmt.Sprintf("// %sService provides operations for %s entities.\n", e.modelName, e.modelName))
		buf.WriteString(fmt.Sprintf("type %sService struct {\n", e.modelName))
		if e.canQuery || e.canGet {
			buf.WriteString(fmt.Sprintf("\tReader[%s]\n", e.modelName))
		}
		if e.canCreate {
			buf.WriteString(fmt.Sprintf("\tCreator[%s]\n", e.modelName))
		}
		if e.canUpdate {
			buf.WriteString(fmt.Sprintf("\tUpdater[%s]\n", e.modelName))
		}
		if e.canPatch {
			buf.WriteString(fmt.Sprintf("\tPatcher[%s]\n", e.modelName))
		}
		if e.canDelete {
			buf.WriteString(fmt.Sprintf("\tDeleter[%s]\n", e.modelName))
		}
		buf.WriteString("}\n\n")
	}

	// Generate child service types
	for _, e := range children {
		buf.WriteString(fmt.Sprintf("type %sChildService struct {\n", e.modelName))
		if e.canQuery || e.canGet {
			buf.WriteString(fmt.Sprintf("\tReader[%s]\n", e.modelName))
		}
		if e.canCreate {
			buf.WriteString(fmt.Sprintf("\tCreator[%s]\n", e.modelName))
		}
		if e.canUpdate {
			buf.WriteString(fmt.Sprintf("\tUpdater[%s]\n", e.modelName))
		}
		if e.canPatch {
			buf.WriteString(fmt.Sprintf("\tPatcher[%s]\n", e.modelName))
		}
		if e.canDelete {
			buf.WriteString(fmt.Sprintf("\tDeleter[%s]\n", e.modelName))
		}
		buf.WriteString("}\n\n")
	}

	// Generate Client fields for top-level entities
	buf.WriteString("// initServices initializes all entity service fields on the Client.\n")
	buf.WriteString("func (c *Client) initServices() {\n")
	for _, e := range topLevel {
		baseName := fmt.Sprintf("base%s", e.modelName)
		buf.WriteString(fmt.Sprintf("\t%s := &baseService{client: c, entityPath: \"/V1.0/%s\", entityName: %q}\n", baseName, e.tag, e.modelName))
		buf.WriteString(fmt.Sprintf("\tc.%s = %sService{\n", e.tag, e.modelName))
		if e.canQuery || e.canGet {
			buf.WriteString(fmt.Sprintf("\t\tReader: Reader[%s]{base: %s},\n", e.modelName, baseName))
		}
		if e.canCreate {
			buf.WriteString(fmt.Sprintf("\t\tCreator: Creator[%s]{base: %s},\n", e.modelName, baseName))
		}
		if e.canUpdate {
			buf.WriteString(fmt.Sprintf("\t\tUpdater: Updater[%s]{base: %s},\n", e.modelName, baseName))
		}
		if e.canPatch {
			buf.WriteString(fmt.Sprintf("\t\tPatcher: Patcher[%s]{base: %s},\n", e.modelName, baseName))
		}
		if e.canDelete {
			buf.WriteString(fmt.Sprintf("\t\tDeleter: Deleter[%s]{base: %s},\n", e.modelName, baseName))
		}
		buf.WriteString("\t}\n")
	}
	buf.WriteString("}\n\n")

	// Generate Client struct fields
	buf.WriteString("// entityServiceFields contains all generated entity service fields for the Client.\n")
	buf.WriteString("type entityServiceFields struct {\n")
	for _, e := range topLevel {
		buf.WriteString(fmt.Sprintf("\t%s %sService\n", e.tag, e.modelName))
	}
	buf.WriteString("}\n\n")

	// Generate child entity accessor methods
	for _, e := range children {
		if e.parentPath == "" || e.childPath == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("// %s returns a child service for %s under the given parent.\n", e.goName, e.modelName))
		buf.WriteString(fmt.Sprintf("func (c *Client) %s(parentID int64) %sChildService {\n", e.goName, e.modelName))
		buf.WriteString(fmt.Sprintf("\tbase := &baseService{\n"))
		buf.WriteString(fmt.Sprintf("\t\tclient:     c,\n"))
		buf.WriteString(fmt.Sprintf("\t\tentityPath: fmt.Sprintf(\"%s/%%d/%s\", parentID),\n", e.parentPath, e.childPath))
		buf.WriteString(fmt.Sprintf("\t\tentityName: %q,\n", e.modelName))
		buf.WriteString(fmt.Sprintf("\t}\n"))
		buf.WriteString(fmt.Sprintf("\treturn %sChildService{\n", e.modelName))
		if e.canQuery || e.canGet {
			buf.WriteString(fmt.Sprintf("\t\tReader: Reader[%s]{base: base},\n", e.modelName))
		}
		if e.canCreate {
			buf.WriteString(fmt.Sprintf("\t\tCreator: Creator[%s]{base: base},\n", e.modelName))
		}
		if e.canUpdate {
			buf.WriteString(fmt.Sprintf("\t\tUpdater: Updater[%s]{base: base},\n", e.modelName))
		}
		if e.canPatch {
			buf.WriteString(fmt.Sprintf("\t\tPatcher: Patcher[%s]{base: base},\n", e.modelName))
		}
		if e.canDelete {
			buf.WriteString(fmt.Sprintf("\t\tDeleter: Deleter[%s]{base: base},\n", e.modelName))
		}
		buf.WriteString("\t}\n}\n\n")
	}

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		os.WriteFile(filepath.Join(outDir, "gen_services.go"), []byte(buf.String()), 0644)
		return fmt.Errorf("format gen_services.go: %w", err)
	}

	return os.WriteFile(filepath.Join(outDir, "gen_services.go"), formatted, 0644)
}
```

- [ ] **Step 3: Run the code generator**

```bash
cd /home/emmaly/Projects/autotask
go run ./cmd/generate -spec ~/.cache/api-explorer/apis/autotask/raw/20260326T0000Z/swagger-apisguru.json -out .
```

Expected: `gen_models.go` and `gen_services.go` created, "code generation complete" printed.

- [ ] **Step 4: Verify generated code compiles**

```bash
go build ./...
```

Expected: clean compile, no errors.

- [ ] **Step 5: Verify key entities exist in generated code**

```bash
grep -c "^type .* struct" gen_models.go
grep "type Ticket struct" gen_models.go
grep "type Company struct" gen_models.go
grep "type TicketService struct" gen_services.go
grep "type CompanyService struct" gen_services.go
```

Expected: ~245 structs in gen_models.go, key types found.

- [ ] **Step 6: Commit**

```bash
git add cmd/generate/main.go generate.go gen_models.go gen_services.go
git commit -m "feat: code generator producing entity models and services"
```

---

### Task 7: Wire Generated Services into Client

**Files:**
- Modify: `client.go` — embed `entityServiceFields` and call `initServices()`

- [ ] **Step 1: Write integration test**

Add to `client_test.go`:

```go
func TestClient_GeneratedServices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"item": map[string]any{"id": 1, "companyName": "ACME"},
		})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		Username:        "test@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
		DisableRateLimitTracking: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test that generated service fields are accessible (plural tag names)
	company, err := c.Companies.Get(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if company == nil {
		t.Fatal("expected company")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run TestClient_GeneratedServices -v
```

Expected: compilation failure — `c.Company` doesn't exist yet.

- [ ] **Step 3: Add entityServiceFields to Client**

Modify `client.go` — embed the generated `entityServiceFields` struct in `Client` and call `initServices()` in `NewClient`:

Add `entityServiceFields` embed to the Client struct:
```go
type Client struct {
	entityServiceFields // embedded generated fields

	config  Config
	// ... rest unchanged
}
```

Add `c.initServices()` call at the end of `NewClient`, before the return:
```go
	c.initServices()
	return c, nil
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -v
```

Expected: all tests PASS including the new integration test.

- [ ] **Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: wire generated entity services into Client"
```

---

### Task 8: Entity Information Methods

**Files:**
- Modify: `traits.go` — add EntityInfo, FieldDefinitions, UDFDefinitions methods

- [ ] **Step 1: Write entity info tests**

Add to `traits_test.go`:

```go
func TestReader_EntityInfo(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/V1.0/TestEntities/entityInformation" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"name":              "TestEntity",
			"canCreate":         true,
			"canUpdate":         true,
			"canDelete":         false,
			"canQuery":          true,
			"hasUserDefinedFields": true,
		})
	})

	reader := Reader[testEntity]{base: base}
	info, err := reader.EntityInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "TestEntity" || !info.CanCreate {
		t.Errorf("unexpected info: %+v", info)
	}
}

func TestReader_FieldDefinitions(t *testing.T) {
	_, base := setupTestService[testEntity](func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/V1.0/TestEntities/entityInformation/fields" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"fields": []map[string]any{
				{"name": "id", "dataType": "integer", "isRequired": true},
				{"name": "name", "dataType": "string", "isRequired": false},
			},
		})
	})

	reader := Reader[testEntity]{base: base}
	fields, err := reader.FieldDefinitions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./... -run "TestReader_Entity|TestReader_Field" -v
```

Expected: compilation failure.

- [ ] **Step 3: Implement entity information methods**

Add to `traits.go`:

```go
// EntityInfo contains metadata about an entity type.
type EntityInfo struct {
	Name                  string `json:"name"`
	CanCreate             bool   `json:"canCreate"`
	CanUpdate             bool   `json:"canUpdate"`
	CanDelete             bool   `json:"canDelete"`
	CanQuery              bool   `json:"canQuery"`
	HasUserDefinedFields  bool   `json:"hasUserDefinedFields"`
}

// FieldDefinition describes a single field on an entity.
type FieldDefinition struct {
	Name         string `json:"name"`
	DataType     string `json:"dataType"`
	IsRequired   bool   `json:"isRequired"`
	IsReadOnly   bool   `json:"isReadOnly"`
	IsQueryable  bool   `json:"isQueryable"`
	MaxLength    int    `json:"maxLength"`
	IsReference  bool   `json:"isReference"`
	ReferenceEntity string `json:"referenceEntityType"`
}

type fieldDefsResponse struct {
	Fields []FieldDefinition `json:"fields"`
}

func (r Reader[T]) EntityInfo(ctx context.Context) (*EntityInfo, error) {
	path := r.base.entityPath + "/entityInformation"
	data, err := r.base.client.doGet(ctx, path)
	if err != nil {
		return nil, err
	}
	var info EntityInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("autotask: decode %s entity info: %w", r.base.entityName, err)
	}
	return &info, nil
}

func (r Reader[T]) FieldDefinitions(ctx context.Context) ([]FieldDefinition, error) {
	path := r.base.entityPath + "/entityInformation/fields"
	data, err := r.base.client.doGet(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp fieldDefsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("autotask: decode %s fields: %w", r.base.entityName, err)
	}
	return resp.Fields, nil
}

func (r Reader[T]) UDFDefinitions(ctx context.Context) ([]FieldDefinition, error) {
	path := r.base.entityPath + "/entityInformation/userDefinedFields"
	data, err := r.base.client.doGet(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp fieldDefsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("autotask: decode %s UDFs: %w", r.base.entityName, err)
	}
	return resp.Fields, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./... -run "TestReader_Entity|TestReader_Field" -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add traits.go traits_test.go
git commit -m "feat: entity info, field definitions, and UDF definitions"
```

---

### Task 9: Full Test Suite and Final Verification

**Files:**
- Modify: `client_test.go` — add zone detection test
- Create: `example_test.go` — usage examples as tests

- [ ] **Step 1: Add zone detection test**

Add to `client_test.go`:

```go
func TestClient_ZoneDetection_Skipped(t *testing.T) {
	// When BaseURL is provided, zone detection should be skipped
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{
			"item": map[string]any{"id": 1},
		})
	}))
	defer srv.Close()

	c, err := NewClient(Config{
		Username:        "test@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
		DisableRateLimitTracking: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.doGet(context.Background(), "/V1.0/Tickets/1")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no zone detection), got %d", calls)
	}
}

func TestClient_ResolveBaseURL(t *testing.T) {
	// Test the resolveBaseURL method directly
	c, err := NewClient(Config{
		Username:        "test@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         "https://example.com",
		DisableRateLimitTracking: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	url, err := c.resolveBaseURL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", url)
	}
}
```

- [ ] **Step 2: Create example test file**

Create `example_test.go`:

```go
package autotask_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/scjalliance/autotask"
)

func ExampleNewClient() {
	// Create a client — in real usage, zone is auto-detected
	client, err := autotask.NewClient(autotask.Config{
		Username:        "api@example.com",
		Secret:          "your-secret",
		IntegrationCode: "YOUR_CODE",
		BaseURL:         "https://webservices5.autotask.net/ATServicesRest",
	})
	if err != nil {
		panic(err)
	}
	_ = client
	fmt.Println("client created")
	// Output: client created
}

func ExampleFilter() {
	q := &autotask.FilterQuery{}
	opt := autotask.Filter(
		autotask.Field("status").Eq(1),
		autotask.Field("companyID").Eq(12345),
	)
	opt(q)
	data, _ := json.Marshal(q)
	fmt.Println(string(data))
	// Output: {"filter":[{"op":"eq","field":"status","value":1},{"op":"eq","field":"companyID","value":12345}]}
}

func ExamplePtr() {
	name := autotask.Ptr("ACME Corp")
	fmt.Println(*name)
	// Output: ACME Corp
}

func ExampleReader_Query() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"id": 1, "companyName": "ACME"},
				{"id": 2, "companyName": "Globex"},
			},
			"pageDetails": map[string]any{"count": 2, "requestCount": 500},
		})
	}))
	defer srv.Close()

	client, _ := autotask.NewClient(autotask.Config{
		Username:        "test@test.com",
		Secret:          "secret",
		IntegrationCode: "TEST",
		BaseURL:         srv.URL,
		DisableRateLimitTracking: true,
	})

	ctx := context.Background()
	companies, err := client.Companies.Query(ctx,
		autotask.Filter(autotask.Field("companyType").Eq(1)),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d companies\n", len(companies))
	// Output: found 2 companies
}
```

- [ ] **Step 3: Run the full test suite**

```bash
go test ./... -v -count=1
```

Expected: all tests PASS.

- [ ] **Step 4: Run go vet**

```bash
go vet ./...
```

Expected: no issues.

- [ ] **Step 5: Commit**

```bash
git add client_test.go example_test.go
git commit -m "test: full test suite with examples"
```

---

### Task 10: README and Documentation

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README**

Create `README.md` with: installation, quick start, query examples, entity list, auth setup, rate limiting notes. Reference the design spec for architecture details.

Key sections:
- Installation: `go get github.com/scjalliance/autotask`
- Quick Start: NewClient with Config, query tickets
- Query Builder: Filter, operators, OR grouping, UDF
- Pagination: auto-paging Query vs streaming QueryIter
- Entity Services: top-level fields, child entity methods
- Error Handling: sentinel errors, APIError
- Rate Limiting: automatic tracking, thresholds
- Code Generation: how to regenerate from updated spec

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: README with usage examples and API reference"
```

---

### Task 11: Final Verification

- [ ] **Step 1: Clean build from scratch**

```bash
cd /home/emmaly/Projects/autotask
go clean -testcache
go build ./...
go test ./... -v -count=1
go vet ./...
```

Expected: clean build, all tests pass, no vet issues.

- [ ] **Step 2: Verify generated code is up to date**

```bash
go generate ./...
git diff --stat
```

Expected: no changes (generated code matches).

- [ ] **Step 3: Review git log**

```bash
git log --oneline
```

Expected: clean commit history with conventional commits.
