# Webhook Receiver Design

**Date:** 2026-04-09
**Status:** Draft

## Overview

Add webhook receiving capabilities to the autotask Go client library. The library already supports webhook registration via generated CRUD services for all 5 webhook entity types. This design covers the **receiver side**: validating inbound webhook payloads (HMAC-SHA1), parsing them into typed Go structs, and handing them off to user code.

### Design Principles

- **Validate, parse, hand back.** The library does not route, dispatch, or decide what webhooks mean. The user owns routing and business logic.
- **Opinionated primary API, escape hatch for control.** `NewWebhookHandler` is the easy path; `ValidateWebhook` is for users who need custom HTTP behavior.
- **No new dependencies.** Everything uses Go stdlib (`crypto/hmac`, `crypto/sha1`, `encoding/base64`, `encoding/json`, `net/http`).

## Autotask Webhook Callback Format

When a subscribed event occurs, Autotask sends an HTTP POST (port 443) to the configured `WebhookURL`. The payload is JSON with these fields:

| Field | Type | Description |
|---|---|---|
| `Action` | string | `"Create"`, `"Update"`, `"Delete"`, or `"Deactivated"` |
| `Guid` | string | Globally unique callout identifier |
| `EntityType` | string | Autotask entity type (e.g., `"Ticket"`, `"Account"`) |
| `Id` | number | Entity ID in Autotask |
| `Fields` | object | Key-value pairs of entity field data (partial; see below) |
| `EventTime` | string | ISO 8601 timestamp of when the webhook fired |
| `SequenceNumber` | number | Incrementing counter; always increasing but may have gaps |
| `PersonID` | number | Resource ID of the user whose action triggered the event |

### Fields Object Behavior

- **Create:** All subscribed fields with non-null values on the new entity.
- **Update:** Only changed subscribed fields, plus any `IsDisplayAlwaysField` fields.
- **Delete:** Minimal or empty.
- **Deactivated:** Minimal or empty. Sent to the `DeactivationURL`, not the main webhook URL.

Field inclusion is controlled by `WebhookField` and `WebhookUdfField` child entities:
- `IsSubscribedField` — a change to this field triggers the callout.
- `IsDisplayAlwaysField` — this field is always included in the payload when any subscribed field triggers.

### HMAC Signature

- **Header:** `X-Hook-Signature: sha1=<base64-encoded-HMAC-SHA1>`
- **Algorithm:** HMAC-SHA1
- **Key:** The `SecretKey` configured on the webhook entity (up to 64 characters)
- **Signed data:** The raw HTTP request body bytes, exactly as sent
- **Encoding:** Base64

Critical: Autotask applies custom Unicode escaping to JSON string values before signing. Verification must use the exact raw bytes received, not re-serialized JSON.

### Supported Entity Types

| Webhook Entity | Autotask EntityType Value | Go Struct |
|---|---|---|
| CompanyWebhook | `"Account"` | `*Company` |
| ContactWebhook | `"Contact"` | `*Contact` |
| ConfigurationItemWebhook | `"InstalledProduct"` | `*ConfigurationItem` |
| TicketWebhook | `"Ticket"` | `*Ticket` |
| TicketNoteWebhook | `"TicketNote"` | `*TicketNote` |

### Rate Limits

- Callouts fire within 1 minute of the trigger event (up to 5 minutes under heavy load).
- Rate limit: 1,500 callouts per rolling hour per entity type.

## API Surface

### WebhookEvent Struct

```go
type WebhookEvent struct {
    Action         string         // "Create", "Update", "Delete", "Deactivated"
    GUID           string         // Unique callout identifier
    EntityType     string         // "Ticket", "Account", "Contact", "InstalledProduct", "TicketNote"
    ID             int64          // Entity ID in Autotask
    Fields         map[string]any // Raw field data from the payload
    EventTime      time.Time      // When the webhook fired
    SequenceNumber int64          // Incrementing counter (may have gaps)
    PersonID       int64          // Resource ID of the triggering user
}
```

All fields are plain values (not pointers) since they are always present in every payload. `Fields` is the raw map for direct access.

### Entity() Method

```go
func (e *WebhookEvent) Entity(v any) error
```

Unmarshals `Fields` into a typed entity struct. Steps:

1. Validate that `v` is a non-nil pointer to a struct.
2. Look up the expected Go type for `e.EntityType` using a package-level map.
3. If `v`'s type doesn't match the expected type, return an error (e.g., `"entity type Ticket cannot be unmarshaled into *Company"`).
4. Marshal `e.Fields` to JSON, then unmarshal into `v`.

This JSON round-trip lets the existing JSON tags and `Time.UnmarshalJSON` handle all type coercion. Pointer fields not present in the payload remain `nil`. The cost is a few microseconds — negligible for a webhook handler.

For `"Delete"` and `"Deactivated"` actions where `Fields` is minimal, `Entity()` works normally — the struct is mostly nil, and the user has `event.ID` for the entity reference.

#### Entity Type Mapping

```go
var webhookEntityTypes = map[string]reflect.Type{
    "Ticket":           reflect.TypeOf(Ticket{}),
    "Account":          reflect.TypeOf(Company{}),
    "Contact":          reflect.TypeOf(Contact{}),
    "InstalledProduct": reflect.TypeOf(ConfigurationItem{}),
    "TicketNote":       reflect.TypeOf(TicketNote{}),
}
```

### NewWebhookHandler (Primary API)

```go
func NewWebhookHandler(secretKey string, fn func(*WebhookEvent)) http.Handler
```

Returns an `http.Handler` that:

1. Rejects non-POST methods with `405 Method Not Allowed`.
2. Calls `ValidateWebhook(r, secretKey)`.
3. On `ErrMissingSignature` or `ErrInvalidSignature` → `401 Unauthorized`.
4. On malformed body or parse error → `400 Bad Request`.
5. On success → calls `fn(event)`, returns `200 OK`.

Error responses are plain text with generic messages (no internal details leaked).

**What it does not do:**
- No logging — user wraps with their own middleware.
- No routing — user mounts this wherever they want.
- No goroutine dispatch — `fn` runs synchronously; user can `go` it themselves.

### ValidateWebhook (Escape Hatch)

```go
func ValidateWebhook(r *http.Request, secretKey string) (*WebhookEvent, error)
```

For users who need custom HTTP error handling, logging, or integration with non-standard HTTP stacks.

Steps:

1. Read the raw body bytes. Replace `r.Body` with a re-readable buffer so downstream code can still read it.
2. Extract the `X-Hook-Signature` header. Expected format: `sha1=<base64>`.
3. If header is missing, return `ErrMissingSignature`.
4. Compute HMAC-SHA1 over raw bytes using `secretKey`.
5. Base64-encode the result and constant-time compare against the header value.
6. If mismatch, return `ErrInvalidSignature`.
7. JSON-unmarshal the raw bytes into `WebhookEvent`.
8. Parse `EventTime` string into `time.Time`.
9. Return the populated `*WebhookEvent`.

### New Sentinel Errors

```go
var ErrInvalidSignature = errors.New("invalid webhook signature")
var ErrMissingSignature = errors.New("missing webhook signature")
```

These follow the existing pattern (`ErrNotFound`, `ErrUnauthorized`, etc.) and work with `errors.Is()`.

## Registration

No new code. The existing generated services provide full CRUD for all 5 webhook entity types plus child resources (`WebhookField`, `WebhookUdfField`, `WebhookExcludedResource`). Users construct the entity structs directly and call `client.TicketWebhooks.Create(ctx, &webhook)`, etc.

## Usage Examples

### Primary API (NewWebhookHandler)

```go
mux := http.NewServeMux()

mux.Handle("/webhooks/tickets", autotask.NewWebhookHandler(secretKey, func(event *autotask.WebhookEvent) {
    var ticket autotask.Ticket
    if err := event.Entity(&ticket); err != nil {
        log.Printf("failed to parse ticket: %v", err)
        return
    }
    log.Printf("ticket %d %s: %s", event.ID, event.Action, *ticket.Title)
}))

http.ListenAndServeTLS(":443", certFile, keyFile, mux)
```

### Escape Hatch (ValidateWebhook)

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    event, err := autotask.ValidateWebhook(r, secretKey)
    if err != nil {
        switch {
        case errors.Is(err, autotask.ErrMissingSignature),
             errors.Is(err, autotask.ErrInvalidSignature):
            mylogger.Warn("webhook auth failed", "err", err, "remote", r.RemoteAddr)
            http.Error(w, "unauthorized", http.StatusUnauthorized)
        default:
            mylogger.Error("webhook parse failed", "err", err)
            http.Error(w, "bad request", http.StatusBadRequest)
        }
        return
    }

    switch event.Action {
    case "Create":
        var ticket autotask.Ticket
        if err := event.Entity(&ticket); err != nil {
            mylogger.Error("entity parse failed", "err", err)
            return
        }
        handleNewTicket(event, &ticket)
    case "Delete":
        handleDeletedTicket(event.ID)
    case "Deactivated":
        mylogger.Warn("webhook deactivated", "guid", event.GUID)
    }

    w.WriteHeader(http.StatusOK)
}
```

### Registering a Webhook

```go
webhook := &autotask.TicketWebhook{
    Name:                       autotask.Ptr("My Ticket Webhook"),
    WebhookURL:                 autotask.Ptr("https://example.com/webhooks/tickets"),
    SecretKey:                  autotask.Ptr("my-secret-key-here"),
    IsActive:                   autotask.Ptr(true),
    IsSubscribedToCreateEvents: autotask.Ptr(true),
    IsSubscribedToUpdateEvents: autotask.Ptr(true),
    IsSubscribedToDeleteEvents: autotask.Ptr(false),
    DeactivationURL:            autotask.Ptr("https://example.com/webhooks/deactivated"),
}

created, err := client.TicketWebhooks.Create(ctx, webhook)
```

## Files

| File | Purpose |
|---|---|
| `webhook.go` | `WebhookEvent`, `ValidateWebhook`, `NewWebhookHandler`, `Entity()`, sentinel errors, entity type map |
| `webhook_test.go` | Tests for HMAC validation, payload parsing, `Entity()` type checking, handler HTTP behavior |

## Testing Strategy

- **HMAC validation:** Compute known signatures, verify acceptance; mutate body or signature, verify rejection.
- **Missing/malformed signature header:** Verify correct sentinel errors.
- **Payload parsing:** Verify all fields unmarshal correctly, including `EventTime` as `time.Time`.
- **Entity() happy path:** Unmarshal `Fields` into correct typed struct, verify `*Time` fields parse.
- **Entity() type mismatch:** Pass `*Company` when `EntityType` is `"Ticket"`, verify error.
- **Entity() on Delete/Deactivated:** Verify mostly-nil struct, no error.
- **NewWebhookHandler:** Test 405 on GET, 401 on bad signature, 400 on bad body, 200 on valid payload. Verify callback receives correct event.
- **ValidateWebhook body re-readability:** Verify `r.Body` can still be read after validation.
