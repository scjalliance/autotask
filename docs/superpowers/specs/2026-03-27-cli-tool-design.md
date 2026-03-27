# Autotask CLI Tool — Design Spec

## Overview

A utility CLI (`cmd/autotask/main.go`) that serves dual purpose: day-to-day IT support tool for quick lookups/searches, and reference example of how to use the `autotask` Go package correctly.

## Auth

Credentials from environment variables:
- `AUTOTASK_USERNAME` (required)
- `AUTOTASK_SECRET` (required)
- `AUTOTASK_INTEGRATION_CODE` (required)
- `AUTOTASK_ZONE` (optional — auto-detected if omitted)

If any required var is missing, print a clear error message listing all three expected vars and exit 1.

## Subcommands

### `autotask whoami`

Connectivity and auth test. Creates the client (triggering zone detection), then prints the resolved zone URL. Demonstrates: `NewClient`, zone detection.

### `autotask ticket <number-or-id>`

Look up a single ticket. Accepts either:
- A display ticket number (contains non-numeric characters, e.g., "T20260327.0001") — queries by `ticketNumber` field
- A numeric internal ID — fetches directly by ID

Displays: ticket number, title, status, priority, company ID, assigned resource ID, created date, due date, description (truncated).

Demonstrates: `Get` by ID, `Query` with `Filter(Field(...).Eq(...))`.

### `autotask tickets <search-term>`

Search tickets by title keyword. Returns up to 25 results in a table.

Demonstrates: `Query` with `Filter(Field("title").Contains(...))`, iterating results.

### `autotask company <id-or-name>`

If numeric: fetch company by ID. Otherwise: search by company name (contains match). Shows: name, phone, address, type, active status.

Demonstrates: `Get` by ID, `Query` with `Contains`.

### `autotask resource <id-or-name>`

If numeric: fetch resource by ID. Otherwise: search by name or email. Shows: name, email, title, active status.

Demonstrates: `Query` with `Or(Field("firstName").Contains(...), Field("lastName").Contains(...), Field("email").Contains(...))`.

## Output

Default: human-readable summary/table to stdout.

`--json` flag on any subcommand: outputs the raw API entity as indented JSON. Useful for scripts and LLM piping.

## Error Handling

- Missing env vars: list all expected vars, exit 1
- API errors: print the error message from `APIError.Errors`, exit 1
- Not found: "no ticket found with number X", exit 1
- No results: "no results", exit 0

## Structure

Single file: `cmd/autotask/main.go`. This is a demo tool — keeping it in one file makes it easy to read as an example. No sub-packages, no framework.

Uses bare `os.Args` parsing — the subcommand set is small enough that a flag library adds nothing.

## Dependencies

Only the `autotask` package and stdlib. Zero external deps.
