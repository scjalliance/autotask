package autotask

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSecret = "test-webhook-secret"

func makeWebhookPayload(t *testing.T, action, entityType string, id int64, fields map[string]any) []byte {
	t.Helper()
	payload := map[string]any{
		"Action":         action,
		"Guid":           "abc-123-def",
		"EntityType":     entityType,
		"Id":             id,
		"Fields":         fields,
		"EventTime":      "2026-04-09T12:00:00.000Z",
		"SequenceNumber": 42,
		"PersonID":       100,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshaling test payload: %v", err)
	}
	return data
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(body)
	return "sha1=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func makeWebhookRequest(body []byte, signature string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	if signature != "" {
		r.Header.Set("X-Hook-Signature", signature)
	}
	return r
}

func TestValidateWebhook_Valid(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 12345, map[string]any{
		"title":  "Test ticket",
		"status": 1,
	})
	sig := signPayload(body, testSecret)
	r := makeWebhookRequest(body, sig)

	event, err := ValidateWebhook(r, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Action != "Create" {
		t.Errorf("Action = %q, want %q", event.Action, "Create")
	}
	if event.GUID != "abc-123-def" {
		t.Errorf("GUID = %q, want %q", event.GUID, "abc-123-def")
	}
	if event.EntityType != "Ticket" {
		t.Errorf("EntityType = %q, want %q", event.EntityType, "Ticket")
	}
	if event.ID != 12345 {
		t.Errorf("ID = %d, want %d", event.ID, 12345)
	}
	if event.SequenceNumber != 42 {
		t.Errorf("SequenceNumber = %d, want %d", event.SequenceNumber, 42)
	}
	if event.PersonID != 100 {
		t.Errorf("PersonID = %d, want %d", event.PersonID, 100)
	}
	if event.EventTime.IsZero() {
		t.Error("EventTime is zero, expected parsed time")
	}
	expected := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	if !event.EventTime.Time.Equal(expected) {
		t.Errorf("EventTime = %v, want %v", event.EventTime.Time, expected)
	}
	if event.Fields["title"] != "Test ticket" {
		t.Errorf("Fields[title] = %v, want %q", event.Fields["title"], "Test ticket")
	}
}

func TestValidateWebhook_MissingSignature(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)
	r := makeWebhookRequest(body, "")

	_, err := ValidateWebhook(r, testSecret)
	if !errors.Is(err, ErrMissingSignature) {
		t.Errorf("err = %v, want ErrMissingSignature", err)
	}
}

func TestValidateWebhook_InvalidSignature(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)
	r := makeWebhookRequest(body, signPayload(body, "wrong-secret"))

	_, err := ValidateWebhook(r, testSecret)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature", err)
	}
}

func TestValidateWebhook_MalformedSignatureHeader(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)

	// Missing sha1= prefix.
	r := makeWebhookRequest(body, "not-a-valid-header")
	_, err := ValidateWebhook(r, testSecret)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature", err)
	}

	// Invalid base64 after sha1= prefix.
	r = makeWebhookRequest(body, "sha1=%%%invalid%%%")
	_, err = ValidateWebhook(r, testSecret)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature for bad base64", err)
	}
}

func TestValidateWebhook_MutatedBody(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)
	sig := signPayload(body, testSecret)

	// Mutate the body after signing.
	mutated := append([]byte{}, body...)
	mutated[0] = '!'

	r := makeWebhookRequest(mutated, sig)
	_, err := ValidateWebhook(r, testSecret)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature for mutated body", err)
	}
}

func TestValidateWebhook_BodyReReadable(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)
	sig := signPayload(body, testSecret)
	r := makeWebhookRequest(body, sig)

	_, err := ValidateWebhook(r, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Body should still be readable.
	reread, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("re-reading body: %v", err)
	}
	if !bytes.Equal(reread, body) {
		t.Error("re-read body does not match original")
	}
}

func TestValidateWebhook_MalformedJSON(t *testing.T) {
	body := []byte(`not json`)
	sig := signPayload(body, testSecret)
	r := makeWebhookRequest(body, sig)

	_, err := ValidateWebhook(r, testSecret)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if errors.Is(err, ErrInvalidSignature) || errors.Is(err, ErrMissingSignature) {
		t.Errorf("err should not be a signature error, got %v", err)
	}
}

func TestWebhookEvent_Entity_Ticket(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "Ticket",
		Fields: map[string]any{
			"id":     float64(42),
			"title":  "Server down",
			"status": float64(1),
		},
	}

	var ticket Ticket
	if err := event.Entity(&ticket); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Title == nil || *ticket.Title != "Server down" {
		t.Errorf("Title = %v, want %q", ticket.Title, "Server down")
	}
	if ticket.ID == nil || *ticket.ID != 42 {
		t.Errorf("ID = %v, want 42", ticket.ID)
	}
}

func TestWebhookEvent_Entity_Company(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "Account",
		Fields: map[string]any{
			"id":          float64(10),
			"companyName": "Acme Corp",
		},
	}

	var company Company
	if err := event.Entity(&company); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.CompanyName == nil || *company.CompanyName != "Acme Corp" {
		t.Errorf("CompanyName = %v, want %q", company.CompanyName, "Acme Corp")
	}
}

func TestWebhookEvent_Entity_ConfigurationItem(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "InstalledProduct",
		Fields:     map[string]any{"id": float64(5)},
	}

	var ci ConfigurationItem
	if err := event.Entity(&ci); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ci.ID == nil || *ci.ID != 5 {
		t.Errorf("ID = %v, want 5", ci.ID)
	}
}

func TestWebhookEvent_Entity_Contact(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "Contact",
		Fields:     map[string]any{"id": float64(7), "firstName": "Jane"},
	}

	var contact Contact
	if err := event.Entity(&contact); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contact.FirstName == nil || *contact.FirstName != "Jane" {
		t.Errorf("FirstName = %v, want %q", contact.FirstName, "Jane")
	}
}

func TestWebhookEvent_Entity_TicketNote(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "TicketNote",
		Fields:     map[string]any{"id": float64(99), "title": "Note title"},
	}

	var note TicketNote
	if err := event.Entity(&note); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Title == nil || *note.Title != "Note title" {
		t.Errorf("Title = %v, want %q", note.Title, "Note title")
	}
}

func TestWebhookEvent_Entity_TypeMismatch(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "Ticket",
		Fields:     map[string]any{"id": float64(1)},
	}

	var company Company
	err := event.Entity(&company)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestWebhookEvent_Entity_UnknownEntityType(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "UnknownThing",
		Fields:     map[string]any{},
	}

	var ticket Ticket
	err := event.Entity(&ticket)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
}

func TestWebhookEvent_Entity_NilPointer(t *testing.T) {
	event := &WebhookEvent{EntityType: "Ticket", Fields: map[string]any{}}

	err := event.Entity((*Ticket)(nil))
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

func TestWebhookEvent_Entity_NonPointer(t *testing.T) {
	event := &WebhookEvent{EntityType: "Ticket", Fields: map[string]any{}}

	var ticket Ticket
	err := event.Entity(ticket)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestWebhookEvent_Entity_DeleteEvent(t *testing.T) {
	event := &WebhookEvent{
		Action:     "Delete",
		EntityType: "Ticket",
		ID:         42,
		Fields:     map[string]any{},
	}

	var ticket Ticket
	if err := event.Entity(&ticket); err != nil {
		t.Fatalf("unexpected error on delete event: %v", err)
	}
	// All fields should be nil on an empty Fields map.
	if ticket.ID != nil {
		t.Errorf("expected nil ID on delete event, got %v", ticket.ID)
	}
}

func TestWebhookEvent_Entity_DeactivatedEvent(t *testing.T) {
	event := &WebhookEvent{
		Action:     "Deactivated",
		EntityType: "Ticket",
		Fields:     map[string]any{},
	}

	var ticket Ticket
	if err := event.Entity(&ticket); err != nil {
		t.Fatalf("unexpected error on deactivated event: %v", err)
	}
}

func TestWebhookEvent_Entity_TimeField(t *testing.T) {
	event := &WebhookEvent{
		EntityType: "Ticket",
		Fields: map[string]any{
			"createDate": "2026-04-09T12:00:00.000Z",
		},
	}

	var ticket Ticket
	if err := event.Entity(&ticket); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.CreateDate == nil {
		t.Fatal("CreateDate is nil, expected parsed time")
	}
	expected := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	if !ticket.CreateDate.Time.Equal(expected) {
		t.Errorf("CreateDate = %v, want %v", ticket.CreateDate.Time, expected)
	}
}

func TestNewWebhookHandler_ValidRequest(t *testing.T) {
	body := makeWebhookPayload(t, "Update", "Ticket", 55, map[string]any{"status": float64(2)})
	sig := signPayload(body, testSecret)

	var received *WebhookEvent
	handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
		received = event
		return nil
	})

	w := httptest.NewRecorder()
	r := makeWebhookRequest(body, sig)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if received == nil {
		t.Fatal("callback was not called")
	}
	if received.ID != 55 {
		t.Errorf("received.ID = %d, want 55", received.ID)
	}
}

func TestNewWebhookHandler_MethodNotAllowed(t *testing.T) {
	handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
		t.Error("callback should not be called for GET")
		return nil
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestNewWebhookHandler_Unauthorized(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)

	handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
		t.Error("callback should not be called for bad signature")
		return nil
	})

	// Missing signature.
	w := httptest.NewRecorder()
	r := makeWebhookRequest(body, "")
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("missing sig: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Wrong signature.
	w = httptest.NewRecorder()
	r = makeWebhookRequest(body, signPayload(body, "wrong"))
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("bad sig: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestNewWebhookHandler_BadRequest(t *testing.T) {
	body := []byte(`not json`)
	sig := signPayload(body, testSecret)

	handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
		t.Error("callback should not be called for bad payload")
		return nil
	})

	w := httptest.NewRecorder()
	r := makeWebhookRequest(body, sig)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestNewWebhookHandler_CallbackError(t *testing.T) {
	body := makeWebhookPayload(t, "Create", "Ticket", 1, nil)
	sig := signPayload(body, testSecret)

	handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
		return errors.New("processing failed")
	})

	w := httptest.NewRecorder()
	r := makeWebhookRequest(body, sig)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestNewWebhookHandler_AllActions(t *testing.T) {
	for _, action := range []string{"Create", "Update", "Delete", "Deactivated"} {
		t.Run(action, func(t *testing.T) {
			body := makeWebhookPayload(t, action, "Ticket", 1, nil)
			sig := signPayload(body, testSecret)

			var gotAction string
			handler := NewWebhookHandler(testSecret, func(event *WebhookEvent) error {
				gotAction = event.Action
				return nil
			})

			w := httptest.NewRecorder()
			r := makeWebhookRequest(body, sig)
			handler.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if gotAction != action {
				t.Errorf("action = %q, want %q", gotAction, action)
			}
		})
	}
}
