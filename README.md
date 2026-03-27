# autotask

A Go client library for the [Autotask PSA](https://www.autotask.net/) REST API. Generated from the Autotask Swagger spec — covers 202 entity types across 208 services.

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
// … 79 top-level services total
```

### Child entity services

Child entities are accessed via methods that accept the parent ID:

```go
notes, err := client.TicketNotes(ticketID).Query(ctx, filter)
```

There are 126 child service methods covering relationships like ticket notes, attachments, contract services, webhook fields, and more.

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

## Helper Functions

```go
// Generic pointer helper — useful when setting optional struct fields
autotask.Ptr("some string")   // *string
autotask.Ptr[int64](42)       // *int64

// Time conversion to/from the API's ISO 8601 format
s := autotask.TimeToString(time.Now())
t, err := autotask.StringToTime(s)
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

## Documentation

Architecture and design decisions: [`docs/superpowers/specs/2026-03-26-autotask-go-client-design.md`](docs/superpowers/specs/2026-03-26-autotask-go-client-design.md)
