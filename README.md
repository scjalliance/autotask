# autotask

A Go client library for the [Autotask PSA](https://www.autotask.net/) REST API. Generated from the Autotask Swagger spec — covers 227 entity types across 228 top-level services.

## Installation

```
go get github.com/scjalliance/autotask
```

Requires Go 1.26+.

## Quick Start

```go
client, err := autotask.NewClient(autotask.Config{
    Username:        "api@example.com",
    Secret:          "your-secret",
    IntegrationCode: "YOUR_CODE",
})
if err != nil {
    log.Fatal(err)
}

// Zone is auto-detected from the username. Query open tickets:
tickets, err := client.Tickets.Query(ctx, autotask.Filter(
    autotask.Field("status").Eq(1),
))
```

## Authentication & Zones

Every request carries three headers: `UserName`, `Secret`, and `ApiIntegrationcode`. On `NewClient`, the library calls the Autotask ZoneInformation endpoint to resolve the correct regional base URL automatically. To skip zone detection (e.g. in tests), set `Config.BaseURL` directly.

Impersonation is supported via context:

```go
ctx = autotask.WithImpersonation(ctx, resourceID)
```

## Query Builder

`Filter` accepts one or more `FilterCondition` values built with `Field` (or `UDFField` for user-defined fields):

```go
filter := autotask.Filter(
    autotask.Field("status").Eq(1),
    autotask.Field("assignedResourceID").NotEq(0),
)
```

### Operators

| Method | API op |
|---|---|
| `Eq(v)` | `eq` |
| `NotEq(v)` | `noteq` |
| `Gt(v)` | `gt` |
| `Gte(v)` | `gte` |
| `Lt(v)` | `lt` |
| `Lte(v)` | `lte` |
| `BeginsWith(s)` | `beginsWith` |
| `EndsWith(s)` | `endsWith` |
| `Contains(s)` | `contains` |
| `Exist()` | `exist` |
| `NotExist()` | `notExist` |
| `In([]any{…})` | `in` |

### Logical grouping

```go
filter := autotask.Filter(
    autotask.Or(
        autotask.Field("priority").Eq(1),
        autotask.Field("priority").Eq(2),
    ),
    autotask.Field("status").Eq(1),
)
```

`And` is also available for explicit grouping.

### UDF filtering

```go
filter := autotask.Filter(
    autotask.UDFField("Customer Type").Eq("Enterprise"),
)
```

## Pagination

`Query` automatically fetches all pages and returns them as a slice:

```go
tickets, err := client.Tickets.Query(ctx, filter)
```

For large result sets, `QueryIter` returns a Go 1.23 range iterator that fetches pages on demand:

```go
for ticket, err := range client.Tickets.QueryIter(ctx, filter) {
    if err != nil {
        return err
    }
    // process one ticket at a time; pages are fetched transparently
}
```

`Count` returns the total match count without fetching records:

```go
n, err := client.Tickets.Count(ctx, filter)
```

## Entity Services

### Top-level services

Top-level entity services live as fields on `*Client`:

```
client.Tickets
client.Companies
client.Contacts
client.Projects
client.TimeEntries
client.Resources
// … 228 top-level services total
```

### Child entity services

Child entities are accessed via methods that accept the parent ID:

```go
notes, err := client.TicketNotes(ticketID).Query(ctx, filter)
```

There are 22 child service methods covering relationships like ticket notes, attachments, contract services, webhook fields, and more. Many entities that also exist as children (e.g., `TicketNotes`) have top-level services as well.

### Introspection

Every service with a `Reader` exposes entity metadata:

```go
info, err := client.Tickets.EntityInfo(ctx)
fields, err := client.Tickets.FieldDefinitions(ctx)
udfs, err := client.Tickets.UDFDefinitions(ctx)
```

## CRUD Operations

```go
// Get by ID
ticket, err := client.Tickets.Get(ctx, 12345)

// Create — returns the new entity ID
id, err := client.Tickets.Create(ctx, &autotask.Ticket{
    Title:     autotask.Ptr("Server is on fire"),
    CompanyID: autotask.Ptr[int64](67890),
    Status:    autotask.Ptr[int64](1),
    Priority:  autotask.Ptr[int64](1),
})

// Full replace
err = client.Tickets.Update(ctx, ticket)

// Partial update — only the fields you supply are changed
err = client.Tickets.Patch(ctx, 12345, autotask.PatchData{
    "status":   2,
    "priority": 3,
})

// Delete
err = client.Tickets.Delete(ctx, 12345)
```

Not all operations are available on every entity; the available traits are determined by the Swagger spec and enforced at compile time.

## Error Handling

Sentinel errors for common HTTP status codes:

```go
ticket, err := client.Tickets.Get(ctx, id)
if errors.Is(err, autotask.ErrNotFound) {
    // 404
}
if errors.Is(err, autotask.ErrUnauthorized) {
    // 401
}
if errors.Is(err, autotask.ErrForbidden) {
    // 403
}
if errors.Is(err, autotask.ErrRateLimited) {
    // 429
}

// Full error detail
var apiErr *autotask.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Errors)
}
```

## Rate Limiting

The client tracks requests per second and automatically sleeps to stay within the configured threshold (default: 10 req/s). Configure or disable it via `Config`:

```go
client, err := autotask.NewClient(autotask.Config{
    // …
    RateLimitThreshold:       20,   // raise the limit
    DisableRateLimitTracking: true, // or disable entirely
})
```

## Retry

Transient failures (5xx server errors and network errors) are automatically retried with exponential backoff (500ms, 1s, 2s). Client errors (4xx) are never retried. Configure via `Config.MaxRetries` (default: 3, set to 0 to disable).

## Time Type

Date-time fields on all entities use the `autotask.Time` type, which wraps `time.Time` with automatic ISO 8601 JSON marshaling:

```go
ticket, _ := client.Tickets.Get(ctx, id)

// Time fields are *autotask.Time — use .Time to get the underlying time.Time
fmt.Println(ticket.CreateDate.Time)           // 2025-10-29 14:53:30 +0000 UTC
fmt.Println(ticket.CreateDate)                // 2025-10-29T14:53:30.000Z
fmt.Println(ticket.CreateDate.Before(time.Now())) // true

// Use in filters — marshals to the correct API string format automatically
filter := autotask.Filter(
    autotask.Field("createDate").Gte(autotask.Time{time.Now().AddDate(0, -1, 0)}),
)

// Construct with Ptr for entity fields
ticket := &autotask.Ticket{
    DueDateTime: autotask.Ptr(autotask.Time{time.Now().Add(24 * time.Hour)}),
}
```

Nil `*Time` fields are omitted from JSON (for PATCH semantics). Zero time marshals as an empty string.

## Picklist Resolution

Translate numeric field values (status, priority, queue, etc.) to human-readable labels:

```go
pl := autotask.NewPicklist(client)

label, err := pl.Resolve(ctx, "/V1.0/Tickets", "status", 5)
// label = "Complete"

labels, err := pl.ResolveAll(ctx, "/V1.0/Tickets", map[string]int64{
    "status":   5,
    "priority": 2,
})
// labels = {"status": "Complete", "priority": "Medium"}
```

Field definitions are fetched once per entity and cached for the lifetime of the `Picklist`.

## Helper Functions

```go
// Generic pointer helper — useful when setting optional struct fields
autotask.Ptr("some string")   // *string
autotask.Ptr[int64](42)       // *int64
```

## Code Generation

The entity types and service wiring in `gen_models.go` and `gen_services.go` are generated from the Autotask Swagger spec. To regenerate after updating the spec:

```
go generate ./...
```

Or run the generator directly:

```
go run ./cmd/generate -spec path/to/swagger.json
```

## CLI Tool

A utility CLI is included at `cmd/autotask` for quick lookups and as a usage example:

```
go install github.com/scjalliance/autotask/cmd/autotask@latest

export AUTOTASK_USERNAME=api@example.com
export AUTOTASK_SECRET=your-secret
export AUTOTASK_INTEGRATION_CODE=YOUR_CODE

autotask whoami                     # test connectivity
autotask ticket T20251029.0002      # look up by display number
autotask ticket 8326                # look up by internal ID
autotask tickets "server"           # search across title, description, notes
autotask company "ACME"             # search companies by name
autotask resource "emmaly"          # search resources by name/email
```

Add `--json` to any command for machine-readable output.

## Documentation

Architecture and design decisions: [`docs/superpowers/specs/2026-03-26-autotask-go-client-design.md`](docs/superpowers/specs/2026-03-26-autotask-go-client-design.md)
