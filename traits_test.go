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

// setupTestService creates a test HTTP server and a baseService pointing at it.
// The caller is responsible for closing the server.
func setupTestService(handler http.HandlerFunc) (*httptest.Server, *Client, *baseService) {
	srv := httptest.NewServer(handler)
	c, _ := NewClient(Config{
		Username:                 "test@test.com",
		Secret:                   "secret",
		IntegrationCode:          "TEST",
		BaseURL:                  srv.URL,
		DisableRateLimitTracking: true,
	})
	base := &baseService{
		client:     c,
		entityPath: "/V1.0/TestEntities",
		entityName: "TestEntity",
	}
	return srv, c, base
}

func TestReader_Get_Success(t *testing.T) {
	id := int64(42)
	name := "Widget"

	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		entity := testEntity{ID: &id, Name: &name}
		itemBytes, _ := json.Marshal(entity)
		json.NewEncoder(w).Encode(map[string]any{"item": json.RawMessage(itemBytes)})
	})
	defer srv.Close()

	reader := &Reader[testEntity]{*base}
	got, err := reader.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID == nil || *got.ID != 42 {
		t.Errorf("ID = %v, want 42", got.ID)
	}
	if got.Name == nil || *got.Name != "Widget" {
		t.Errorf("Name = %v, want Widget", got.Name)
	}
}

func TestReader_Get_NotFound(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"not found"}})
	})
	defer srv.Close()

	reader := &Reader[testEntity]{*base}
	_, err := reader.Get(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestReader_Query_MultiPage(t *testing.T) {
	var srvURL string

	page1ID := int64(1)
	page1Name := "Alpha"
	page2ID := int64(2)
	page2Name := "Beta"

	var requestMethods []string

	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		requestMethods = append(requestMethods, r.Method)

		if r.Method == http.MethodPost && r.URL.Path == "/V1.0/TestEntities/query" {
			// Page 1 — POST
			entity := testEntity{ID: &page1ID, Name: &page1Name}
			itemBytes, _ := json.Marshal(entity)
			nextURL := srvURL + "/V1.0/TestEntities/query?page=2"
			json.NewEncoder(w).Encode(map[string]any{
				"items": []json.RawMessage{itemBytes},
				"pageDetails": map[string]any{
					"count":       1,
					"nextPageUrl": nextURL,
				},
			})
		} else {
			// Page 2 — GET via absolute nextPageUrl
			entity := testEntity{ID: &page2ID, Name: &page2Name}
			itemBytes, _ := json.Marshal(entity)
			json.NewEncoder(w).Encode(map[string]any{
				"items": []json.RawMessage{itemBytes},
				"pageDetails": map[string]any{
					"count": 1,
				},
			})
		}
	})
	defer srv.Close()
	srvURL = srv.URL

	reader := &Reader[testEntity]{*base}
	results, err := reader.Query(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if *results[0].ID != 1 || *results[1].ID != 2 {
		t.Errorf("unexpected IDs: %v, %v", *results[0].ID, *results[1].ID)
	}

	// Verify first request was POST, second was GET.
	if len(requestMethods) < 2 {
		t.Fatalf("expected at least 2 requests, got %d", len(requestMethods))
	}
	if requestMethods[0] != http.MethodPost {
		t.Errorf("page 1 method = %s, want POST", requestMethods[0])
	}
	if requestMethods[1] != http.MethodGet {
		t.Errorf("page 2 method = %s, want GET", requestMethods[1])
	}
}

func TestReader_Count(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities/query/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"queryCount": 57})
	})
	defer srv.Close()

	reader := &Reader[testEntity]{*base}
	count, err := reader.Count(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 57 {
		t.Errorf("count = %d, want 57", count)
	}
}

func TestCreator_Create(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"item": map[string]any{"itemId": 123},
		})
	})
	defer srv.Close()

	creator := &Creator[testEntity]{*base}
	id, err := creator.Create(context.Background(), &testEntity{Name: Ptr("NewWidget")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 123 {
		t.Errorf("id = %d, want 123", id)
	}
}

func TestUpdater_Update(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"item": nil})
	})
	defer srv.Close()

	updater := &Updater[testEntity]{*base}
	err := updater.Update(context.Background(), &testEntity{ID: Ptr(int64(10)), Name: Ptr("Updated")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPatcher_Patch_IDInURL(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		wantPath := fmt.Sprintf("/V1.0/TestEntities/%d", 77)
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		json.NewEncoder(w).Encode(map[string]any{"item": nil})
	})
	defer srv.Close()

	patcher := &Patcher[testEntity]{*base}
	err := patcher.Patch(context.Background(), 77, PatchData{"name": "Patched"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleter_Delete(t *testing.T) {
	srv, _, base := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/V1.0/TestEntities/55" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	deleter := &Deleter[testEntity]{*base}
	err := deleter.Delete(context.Background(), 55)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
