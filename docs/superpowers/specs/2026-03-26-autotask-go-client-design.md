# Autotask Go Client Library — Design Spec

## Overview

A comprehensive Go client library for the Datto Autotask PSA REST API. Code-generated from the official Swagger 2.0 spec, using Go generics for type-safe CRUD operations across all 245 entity types.

Primary focus: internal IT support workflows (tickets, companies, contacts, configuration items, time entries, resources, projects, contracts, knowledge base). Sales-oriented entities (opportunities, quotes, sales orders) are included but lower priority.

## API Summary

- **Spec:** Swagger 2.0, 2031 paths, 371 entity tags, 245 model definitions
- **Base URL:** `https://webservices{N}.autotask.net/ATServicesRest/V1.0/`
- **Auth:** Three required headers per request: `UserName`, `Secret`, `ApiIntegrationcode`
- **Pagination:** 500 records max per page, cursor-based via `nextPageUrl`/`prevPageUrl`
- **Rate limit:** 10,000 requests/hour per database, latency penalty at 50%+ usage
- **Query:** JSON filter DSL with 12 operators, AND/OR grouping, UDF support
- **IDs:** int64, timestamps: ISO 8601 strings

## Architecture

### Package Structure

```
autotask/
  client.go          — Config, Client, HTTP transport, zone detection
  query.go           — Filter, operators, serialization
  pagination.go      — auto-paging iterator (range-over-func)
  traits.go          — Reader[T], Creator[T], Updater[T], Patcher[T], Deleter[T]
  errors.go          — error types
  udf.go             — UserDefinedField type and helpers
  gen_models.go      — generated entity structs
  gen_services.go    — generated service types + Client accessors
  generate.go        — //go:generate directive
  cmd/generate/
    main.go          — codegen tool
```

Single `package autotask` — no sub-packages. Import as one unit.

### Client & Auth

```go
client, err := autotask.NewClient(autotask.Config{
    Username:        "api@example.com",
    Secret:          "...",
    IntegrationCode: "INTERNAL_IT",
    Zone:            "", // auto-detected if omitted
})
```

- **Zone detection:** On first request, calls `GET /V1.0/ZoneInformation?user={username}` to resolve the correct zone URL, caches for client lifetime.
- **Auth headers:** `UserName`, `Secret`, `ApiIntegrationcode` injected on every request automatically.
- **Impersonation:** `ImpersonationResourceId` settable per-request via context or client option.
- **HTTP client:** Injectable `*http.Client` for testing/proxying.
- **Rate limit awareness:** Tracks request rate, backs off at 50%+ threshold to avoid latency penalties.

### Capability Traits (Generics)

Each CRUD capability is a separate generic struct with its methods:

```go
type Reader[T any]  struct { base *baseService }
type Creator[T any] struct { base *baseService }
type Updater[T any] struct { base *baseService }
type Patcher[T any] struct { base *baseService }
type Deleter[T any] struct { base *baseService }
```

Methods on each trait:

- **Reader[T]:** `Get(ctx, id)`, `Query(ctx, ...FilterOption)`, `QueryIter(ctx, ...FilterOption)`, `Count(ctx, ...FilterOption)`, `EntityInfo(ctx)`, `FieldDefinitions(ctx)`, `UDFDefinitions(ctx)`
- **Creator[T]:** `Create(ctx, *T)`
- **Updater[T]:** `Update(ctx, *T)`
- **Patcher[T]:** `Patch(ctx, id, Patch)`
- **Deleter[T]:** `Delete(ctx, id)`

### Generated Entity Services

Codegen composes the correct traits per entity based on swagger capabilities:

```go
// Full CRUD (tickets)
type TicketService struct {
    Reader[Ticket]
    Creator[Ticket]
    Updater[Ticket]
    Patcher[Ticket]
}

// Read-only (contacts)
type ContactService struct {
    Reader[Contact]
}

// CRUD + delete (time entries)
type TimeEntryService struct {
    Reader[TimeEntry]
    Creator[TimeEntry]
    Updater[TimeEntry]
    Patcher[TimeEntry]
    Deleter[TimeEntry]
}
```

Compile-time safe: `client.Contacts.Create()` does not exist.

### Client Accessors

```go
// Top-level entities as fields on Client
client.Tickets        // TicketService
client.Companies      // CompanyService
client.TimeEntries    // TimeEntryService

// Child entities via method (needs parent ID)
client.TicketNotes(parentID)       // ChildReader[TicketNote] etc.
client.CompanyLocations(parentID)
```

### Query Builder

Fluent filter DSL:

```go
// Implicit AND
results, err := client.Tickets.Query(ctx,
    autotask.Filter(
        autotask.Field("status").Eq(1),
        autotask.Field("companyID").Eq(12345),
    ),
)

// OR grouping
results, err := client.Tickets.Query(ctx,
    autotask.Filter(
        autotask.Or(
            autotask.Field("priority").Eq(1),
            autotask.Field("priority").Eq(2),
        ),
    ),
)

// UDF
autotask.UDF("CustomerRanking").Eq("Golden")

// Streaming large result sets (range-over-func, Go 1.23+)
for ticket, err := range client.Tickets.QueryIter(ctx, autotask.Filter(...)).All() {
    // yields one at a time, auto-pages transparently
}
```

**Operators:** `Eq`, `NotEq`, `Gt`, `Gte`, `Lt`, `Lte`, `BeginsWith`, `EndsWith`, `Contains`, `Exist`, `NotExist`, `In`

Filter serializes to the JSON structure the API expects:
```json
{"filter": [{"op": "eq", "field": "status", "value": 1}]}
```

### Pagination

- `Query()` returns `[]T` — auto-pages internally, collects all results.
- `QueryIter()` returns an iterator — streams results, fetches next page on demand.
- Follows `nextPageUrl` from `pageDetails` in each response.
- `Count()` uses the `/query/count` endpoint for efficient counting without fetching data.

### Generated Models

All 245 entity structs with:

- Exported Go fields, `json:"camelCase,omitempty"` tags
- Pointer types for all fields (distinguishes zero from absent, critical for PATCH)
- `UserDefinedFields []UDF` where the entity supports UDFs
- ReadOnly fields marked with comments

```go
type Ticket struct {
    ID                *int64  `json:"id,omitempty"`
    CompanyID         *int64  `json:"companyID,omitempty"`
    Title             *string `json:"title,omitempty"`
    Description       *string `json:"description,omitempty"`
    Status            *int64  `json:"status,omitempty"`
    Priority          *int64  `json:"priority,omitempty"`
    DueDateTime       *string `json:"dueDateTime,omitempty"`
    // ... remaining fields
    UserDefinedFields []UDF   `json:"userDefinedFields,omitempty"`
}

type UDF struct {
    Name  string      `json:"name"`
    Value interface{} `json:"value"`
}
```

### Error Handling

```go
var ErrNotFound      // 404 or empty item response
var ErrUnauthorized  // 401
var ErrForbidden     // 403
var ErrRateLimited   // threshold exceeded

type APIError struct {
    StatusCode int
    Errors     []string
}
```

Standard `errors.Is` / `errors.As` support.

### Code Generation

**Tool:** `cmd/generate/main.go` — standalone Go program.

**Inputs:** Swagger spec from `~/.cache/api-explorer/apis/autotask/` (raw swagger JSON).

**Outputs:** `gen_models.go`, `gen_services.go`

**Invocation:** `go generate ./...` via `//go:generate` directive in package root.

**What it does:**
1. Reads swagger spec, parses definitions and paths
2. For each definition ending in `Model`: generates a Go struct with typed fields
3. For each entity tag: determines capabilities from available operations, generates a service struct composing the correct traits
4. Generates Client field/method accessors for all entity services

### Testing

- Injectable `*http.Client` for HTTP-level testing
- Tests use `httptest.Server` to verify request shaping and response parsing
- No mocking the Autotask API itself — test the HTTP contract
- Focus on: auth header injection, query serialization, pagination traversal, error mapping

## Constraints & Decisions

- **Go 1.23+** required (generics, range-over-func iterators)
- **All fields are pointers** — necessary for the PATCH semantic (omit vs zero)
- **Single package** — no sub-packages, flat structure, clean import
- **Generated code checked in** — consumers don't need the codegen tool
- **Swagger spec cached globally** at `~/.cache/api-explorer/` — not vendored into the project
