// Command autotask is a utility CLI for common Autotask PSA operations.
//
// It demonstrates idiomatic usage of the autotask Go package and is useful
// for quick lookups, scripting, and LLM-driven workflows.
//
// # Authentication
//
// Set these environment variables:
//
//	AUTOTASK_USERNAME          API user email
//	AUTOTASK_SECRET            API user secret
//	AUTOTASK_INTEGRATION_CODE  Tracking identifier
//	AUTOTASK_ZONE              Base URL (optional, auto-detected)
//
// # Usage
//
//	autotask whoami
//	autotask ticket <number-or-id>
//	autotask tickets <search-term>
//	autotask company <id-or-name>
//	autotask resource <id-or-name>
//
// Add --json to any subcommand for machine-readable output.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/scjalliance/autotask"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	// Pull --json flag from anywhere in args.
	jsonOutput := false
	var filtered []string
	for _, a := range args {
		if a == "--json" {
			jsonOutput = true
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	cmd := args[0]
	rest := args[1:]

	client, err := newClient()
	if err != nil {
		fatal(err.Error())
	}

	ctx := context.Background()

	switch cmd {
	case "whoami":
		cmdWhoami(ctx, client, jsonOutput)
	case "ticket":
		if len(rest) == 0 {
			fatal("usage: autotask ticket <number-or-id>")
		}
		cmdTicket(ctx, client, rest[0], jsonOutput)
	case "tickets":
		if len(rest) == 0 {
			fatal("usage: autotask tickets <search-term>")
		}
		cmdTickets(ctx, client, strings.Join(rest, " "), jsonOutput)
	case "company":
		if len(rest) == 0 {
			fatal("usage: autotask company <id-or-name>")
		}
		cmdCompany(ctx, client, strings.Join(rest, " "), jsonOutput)
	case "resource":
		if len(rest) == 0 {
			fatal("usage: autotask resource <id-or-name>")
		}
		cmdResource(ctx, client, strings.Join(rest, " "), jsonOutput)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: autotask <command> [--json] [args...]

Commands:
  whoami                   Test connectivity, show zone info
  ticket <number-or-id>   Look up a ticket by display number or internal ID
  tickets <search-term>   Search tickets by title
  company <id-or-name>    Look up company by ID or search by name
  resource <id-or-name>   Look up resource by ID or search by name/email

Environment:
  AUTOTASK_USERNAME          API user email (required)
  AUTOTASK_SECRET            API user secret (required)
  AUTOTASK_INTEGRATION_CODE  Tracking identifier (required)
  AUTOTASK_ZONE              Base URL (optional, auto-detected)`)
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}

func newClient() (*autotask.Client, error) {
	username := os.Getenv("AUTOTASK_USERNAME")
	secret := os.Getenv("AUTOTASK_SECRET")
	code := os.Getenv("AUTOTASK_INTEGRATION_CODE")

	var missing []string
	if username == "" {
		missing = append(missing, "AUTOTASK_USERNAME")
	}
	if secret == "" {
		missing = append(missing, "AUTOTASK_SECRET")
	}
	if code == "" {
		missing = append(missing, "AUTOTASK_INTEGRATION_CODE")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return autotask.NewClient(autotask.Config{
		Username:        username,
		Secret:          secret,
		IntegrationCode: code,
		BaseURL:         os.Getenv("AUTOTASK_ZONE"),
	})
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// cmdWhoami tests connectivity by creating the client (which triggers zone
// detection if no zone is configured) and printing the resolved base URL.
func cmdWhoami(ctx context.Context, client *autotask.Client, jsonOut bool) {
	// Make a lightweight API call to verify credentials work.
	// We query the API version endpoint, which is cheap.
	info, err := client.Tickets.EntityInfo(ctx)
	if err != nil {
		fatal(fmt.Sprintf("API call failed: %v", err))
	}

	if jsonOut {
		printJSON(map[string]any{
			"status":    "ok",
			"entity":    info.Name,
			"canCreate": info.CanCreate,
			"canQuery":  info.CanQuery,
		})
		return
	}

	fmt.Println("Connected to Autotask API")
	fmt.Printf("  Ticket entity: canCreate=%v canQuery=%v\n", info.CanCreate, info.CanQuery)
	fmt.Println("  Auth: OK")
}

// cmdTicket looks up a single ticket by display number or internal ID.
func cmdTicket(ctx context.Context, client *autotask.Client, query string, jsonOut bool) {
	var ticket *autotask.Ticket

	// If it's purely numeric, treat it as an internal ID.
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		t, err := client.Tickets.Get(ctx, id)
		if err != nil {
			fatal(fmt.Sprintf("ticket %d: %v", id, err))
		}
		ticket = t
	} else {
		// Otherwise, it's a display ticket number — query by ticketNumber field.
		filter := autotask.Filter(
			autotask.Field("ticketNumber").Eq(query),
		)
		results, err := client.Tickets.Query(ctx, filter.AsFilterOption())
		if err != nil {
			fatal(fmt.Sprintf("searching for ticket %q: %v", query, err))
		}
		if len(results) == 0 {
			fatal(fmt.Sprintf("no ticket found with number %q", query))
		}
		ticket = results[0]
	}

	if jsonOut {
		printJSON(ticket)
		return
	}

	fmt.Printf("Ticket: %s (ID: %d)\n", deref(ticket.TicketNumber), deref(ticket.ID))
	fmt.Printf("  Title:       %s\n", deref(ticket.Title))
	fmt.Printf("  Status:      %d\n", deref(ticket.Status))
	fmt.Printf("  Priority:    %d\n", deref(ticket.Priority))
	fmt.Printf("  Queue:       %d\n", deref(ticket.QueueID))
	fmt.Printf("  Company:     %d\n", deref(ticket.CompanyID))
	fmt.Printf("  Assigned To: %d\n", deref(ticket.AssignedResourceID))
	fmt.Printf("  Created:     %s\n", deref(ticket.CreateDate))
	fmt.Printf("  Due:         %s\n", deref(ticket.DueDateTime))
	if desc := deref(ticket.Description); desc != "" {
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		fmt.Printf("  Description: %s\n", desc)
	}
}

// cmdTickets searches tickets across title, description, and ticket notes.
//
// This demonstrates a practical multi-source search pattern: query the main
// entity for text field matches, then also search the TicketNotes entity
// (which has a top-level query endpoint) for note content matches. Results
// are merged and deduplicated by ticket ID.
func cmdTickets(ctx context.Context, client *autotask.Client, search string, jsonOut bool) {
	// Search ticket fields: title, description, and ticket number.
	ticketFilter := autotask.Filter(
		autotask.Or(
			autotask.Field("title").Contains(search),
			autotask.Field("description").Contains(search),
			autotask.Field("ticketNumber").Contains(search),
		),
	)

	// Search ticket notes (title and description/body).
	// TicketNotes has a top-level query endpoint at /V1.0/TicketNotes/query
	// that searches across all tickets. We build a Reader manually since the
	// generated code only has the child service (per-ticket).
	noteFilter := autotask.Filter(
		autotask.Or(
			autotask.Field("title").Contains(search),
			autotask.Field("description").Contains(search),
		),
	)

	// Run both searches concurrently.
	type ticketResult struct {
		tickets []*autotask.Ticket
		err     error
	}
	type noteResult struct {
		notes []*autotask.TicketNote
		err   error
	}

	ticketCh := make(chan ticketResult, 1)
	noteCh := make(chan noteResult, 1)

	go func() {
		results, err := client.Tickets.Query(ctx, ticketFilter.AsFilterOption())
		ticketCh <- ticketResult{results, err}
	}()

	go func() {
		results, err := client.TicketNotes.Query(ctx, noteFilter.AsFilterOption())
		noteCh <- noteResult{results, err}
	}()

	tr := <-ticketCh
	if tr.err != nil {
		fatal(fmt.Sprintf("searching tickets: %v", tr.err))
	}

	// Collect ticket IDs from direct matches.
	seen := map[int64]bool{}
	for _, t := range tr.tickets {
		seen[deref(t.ID)] = true
	}

	// Collect ticket IDs from note matches, fetch any tickets we don't have yet.
	nr := <-noteCh
	if nr.err != nil {
		// Note search failure is non-fatal — we still have ticket field matches.
		fmt.Fprintf(os.Stderr, "warning: note search failed: %v\n", nr.err)
	} else {
		var missingIDs []int64
		for _, note := range nr.notes {
			tid := deref(note.TicketID)
			if tid != 0 && !seen[tid] {
				seen[tid] = true
				missingIDs = append(missingIDs, tid)
			}
		}

		// Fetch tickets found only via notes. Batch by querying with In().
		if len(missingIDs) > 0 {
			ids := make([]any, len(missingIDs))
			for i, id := range missingIDs {
				ids[i] = id
			}
			idFilter := autotask.Filter(autotask.Field("id").In(ids))
			extra, err := client.Tickets.Query(ctx, idFilter.AsFilterOption())
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: fetching note-matched tickets: %v\n", err)
			} else {
				tr.tickets = append(tr.tickets, extra...)
			}
		}
	}

	results := tr.tickets
	if len(results) == 0 {
		fmt.Println("no results")
		return
	}

	if jsonOut {
		printJSON(results)
		return
	}

	// Cap display at 25 for readability.
	display := results
	if len(display) > 25 {
		display = display[:25]
	}

	fmt.Printf("%-16s %-8s %-6s %s\n", "NUMBER", "ID", "STATUS", "TITLE")
	fmt.Println(strings.Repeat("-", 72))
	for _, t := range display {
		title := deref(t.Title)
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		fmt.Printf("%-16s %-8d %-6d %s\n",
			deref(t.TicketNumber),
			deref(t.ID),
			deref(t.Status),
			title,
		)
	}
	if len(results) > 25 {
		fmt.Printf("\n... and %d more (showing first 25)\n", len(results)-25)
	}
	fmt.Printf("\n%d ticket(s) found\n", len(results))
}

// cmdCompany looks up a company by ID or searches by name.
func cmdCompany(ctx context.Context, client *autotask.Client, query string, jsonOut bool) {
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		company, err := client.Companies.Get(ctx, id)
		if err != nil {
			fatal(fmt.Sprintf("company %d: %v", id, err))
		}
		if jsonOut {
			printJSON(company)
			return
		}
		printCompany(company)
		return
	}

	filter := autotask.Filter(
		autotask.Field("companyName").Contains(query),
	)
	results, err := client.Companies.Query(ctx, filter.AsFilterOption())
	if err != nil {
		fatal(fmt.Sprintf("searching companies: %v", err))
	}
	if len(results) == 0 {
		fmt.Println("no results")
		return
	}

	if jsonOut {
		printJSON(results)
		return
	}

	if len(results) == 1 {
		printCompany(results[0])
		return
	}

	fmt.Printf("%-8s %-30s %s\n", "ID", "NAME", "PHONE")
	fmt.Println(strings.Repeat("-", 60))
	for _, c := range results {
		name := deref(c.CompanyName)
		if len(name) > 28 {
			name = name[:25] + "..."
		}
		fmt.Printf("%-8d %-30s %s\n", deref(c.ID), name, deref(c.Phone))
	}
}

func printCompany(c *autotask.Company) {
	fmt.Printf("Company: %s (ID: %d)\n", deref(c.CompanyName), deref(c.ID))
	fmt.Printf("  Phone:   %s\n", deref(c.Phone))
	fmt.Printf("  Address: %s\n", deref(c.Address1))
	if a2 := deref(c.Address2); a2 != "" {
		fmt.Printf("           %s\n", a2)
	}
	fmt.Printf("  City:    %s\n", deref(c.City))
	fmt.Printf("  State:   %s\n", deref(c.State))
	fmt.Printf("  Zip:     %s\n", deref(c.PostalCode))
	fmt.Printf("  Type:    %d\n", deref(c.CompanyType))
	fmt.Printf("  Active:  %v\n", deref(c.IsActive))
}

// cmdResource looks up a resource by ID or searches by name/email.
func cmdResource(ctx context.Context, client *autotask.Client, query string, jsonOut bool) {
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		resource, err := client.Resources.Get(ctx, id)
		if err != nil {
			fatal(fmt.Sprintf("resource %d: %v", id, err))
		}
		if jsonOut {
			printJSON(resource)
			return
		}
		printResource(resource)
		return
	}

	filter := autotask.Filter(
		autotask.Or(
			autotask.Field("firstName").Contains(query),
			autotask.Field("lastName").Contains(query),
			autotask.Field("email").Contains(query),
		),
	)
	results, err := client.Resources.Query(ctx, filter.AsFilterOption())
	if err != nil {
		fatal(fmt.Sprintf("searching resources: %v", err))
	}
	if len(results) == 0 {
		fmt.Println("no results")
		return
	}

	if jsonOut {
		printJSON(results)
		return
	}

	if len(results) == 1 {
		printResource(results[0])
		return
	}

	fmt.Printf("%-8s %-20s %-20s %s\n", "ID", "FIRST", "LAST", "EMAIL")
	fmt.Println(strings.Repeat("-", 72))
	for _, r := range results {
		fmt.Printf("%-8d %-20s %-20s %s\n",
			deref(r.ID),
			deref(r.FirstName),
			deref(r.LastName),
			deref(r.Email),
		)
	}
}

func printResource(r *autotask.Resource) {
	fmt.Printf("Resource: %s %s (ID: %d)\n", deref(r.FirstName), deref(r.LastName), deref(r.ID))
	fmt.Printf("  Email:    %s\n", deref(r.Email))
	fmt.Printf("  Title:    %s\n", deref(r.Title))
	fmt.Printf("  Phone:    %s\n", deref(r.OfficePhone))
	fmt.Printf("  Mobile:   %s\n", deref(r.MobilePhone))
	fmt.Printf("  Active:   %v\n", deref(r.IsActive))
	fmt.Printf("  Username: %s\n", deref(r.UserName))
}
