package autotask

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// A contact is written through its company (the generated CompanyContacts
// child service), with its id in the body.
func TestCompanyContactsPatch(t *testing.T) {
	srv, c, _ := setupTestService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/V1.0/Companies/7/Contacts" {
			t.Errorf("request = %s %s, want PATCH /V1.0/Companies/7/Contacts", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			return
		}
		if id, _ := body["id"].(float64); int64(id) != 30682885 || body["isActive"] != float64(0) {
			t.Errorf("body = %v, want id 30682885 and isActive 0", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"itemId": 30682885})
	})
	defer srv.Close()
	contacts := c.CompanyContacts(7)
	if err := contacts.Patch(context.Background(), 30682885, PatchData{"isActive": 0}); err != nil {
		t.Fatal(err)
	}
}
