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

func TestClient_ZoneDetection_Skipped(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{"item": map[string]any{"id": 1}})
	}))
	defer srv.Close()

	c, _ := NewClient(Config{
		Username: "test@test.com", Secret: "secret", IntegrationCode: "TEST",
		BaseURL: srv.URL, DisableRateLimitTracking: true,
	})
	c.doGet(context.Background(), "/V1.0/Tickets/1")
	if calls != 1 {
		t.Errorf("expected 1 call (no zone detection), got %d", calls)
	}
}

func TestClient_ResolveBaseURL(t *testing.T) {
	c, _ := NewClient(Config{
		Username: "test@test.com", Secret: "secret", IntegrationCode: "TEST",
		BaseURL: "https://example.com", DisableRateLimitTracking: true,
	})
	if c.baseURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", c.baseURL)
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
