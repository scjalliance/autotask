package autotask

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// maxWebhookBody is the maximum request body size accepted by ValidateWebhook.
const maxWebhookBody = 1 << 20 // 1 MB

// webhookEntityTypes maps Autotask EntityType strings from webhook payloads to
// the corresponding Go struct types.
var webhookEntityTypes = map[string]reflect.Type{
	"Ticket":           reflect.TypeOf(Ticket{}),
	"Account":          reflect.TypeOf(Company{}),
	"Contact":          reflect.TypeOf(Contact{}),
	"InstalledProduct": reflect.TypeOf(ConfigurationItem{}),
	"TicketNote":       reflect.TypeOf(TicketNote{}),
}

// WebhookEvent represents a parsed and validated Autotask webhook callback payload.
type WebhookEvent struct {
	Action         string         `json:"Action"`
	GUID           string         `json:"Guid"`
	EntityType     string         `json:"EntityType"`
	ID             int64          `json:"Id"`
	Fields         map[string]any `json:"Fields"`
	EventTime      Time           `json:"EventTime"`
	SequenceNumber int64          `json:"SequenceNumber"`
	PersonID       int64          `json:"PersonID"`
}

// Entity unmarshals the webhook Fields into a typed entity struct. The target
// must be a non-nil pointer to a struct whose type matches the event's
// EntityType. For example, a "Ticket" event can be unmarshaled into *Ticket.
//
// Fields not present in the payload remain at their zero value (nil for pointer
// fields). This works naturally for Delete and Deactivated events where Fields
// is minimal or empty.
func (e *WebhookEvent) Entity(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("autotask: Entity requires a non-nil pointer, got %T", v)
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("autotask: Entity requires a pointer to a struct, got %T", v)
	}

	expected, ok := webhookEntityTypes[e.EntityType]
	if !ok {
		return fmt.Errorf("autotask: unknown webhook entity type %q", e.EntityType)
	}
	if elem.Type() != expected {
		return fmt.Errorf("autotask: entity type %s cannot be unmarshaled into %T", e.EntityType, v)
	}

	data, err := json.Marshal(e.Fields)
	if err != nil {
		return fmt.Errorf("autotask: marshaling webhook fields: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("autotask: unmarshaling webhook fields into %T: %w", v, err)
	}
	return nil
}

// ValidateWebhook reads and validates an inbound Autotask webhook request. It
// verifies the HMAC-SHA1 signature in the X-Hook-Signature header against the
// raw request body using the provided secret key, then parses the payload into
// a WebhookEvent.
//
// The request body is replaced with a re-readable buffer after reading so that
// downstream code can still access it.
//
// Returns ErrMissingSignature if the signature header is absent, or
// ErrInvalidSignature if the signature does not match.
func ValidateWebhook(r *http.Request, secretKey string) (*WebhookEvent, error) {
	// Read the raw body with a size limit.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
	if err != nil {
		return nil, fmt.Errorf("autotask: reading webhook body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	// Extract and validate the signature header.
	sigHeader := r.Header.Get("X-Hook-Signature")
	if sigHeader == "" {
		return nil, ErrMissingSignature
	}

	if !strings.HasPrefix(sigHeader, "sha1=") {
		return nil, ErrInvalidSignature
	}
	sigB64 := sigHeader[len("sha1="):]

	sigBytes, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	// Compute HMAC-SHA1 over the raw body bytes and compare.
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write(body)
	if !hmac.Equal(mac.Sum(nil), sigBytes) {
		return nil, ErrInvalidSignature
	}

	// Parse the payload.
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("autotask: parsing webhook payload: %w", err)
	}

	return &event, nil
}

// NewWebhookHandler returns an http.Handler that validates and parses inbound
// Autotask webhook callbacks. It verifies the HMAC-SHA1 signature, parses the
// payload, and calls fn with the resulting WebhookEvent.
//
// If fn returns nil the handler responds with 200 OK. If fn returns an error
// the handler responds with 500 Internal Server Error, which signals Autotask
// to retry the delivery.
//
// HTTP error responses:
//   - 405 Method Not Allowed for non-POST requests
//   - 401 Unauthorized for missing or invalid signatures
//   - 400 Bad Request for malformed payloads
//   - 500 Internal Server Error if fn returns an error
func NewWebhookHandler(secretKey string, fn func(*WebhookEvent) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		event, err := ValidateWebhook(r, secretKey)
		if err != nil {
			switch {
			case err == ErrMissingSignature || err == ErrInvalidSignature:
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			default:
				http.Error(w, "bad request", http.StatusBadRequest)
			}
			return
		}

		if err := fn(event); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
