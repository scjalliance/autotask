package autotask

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

// Picklist provides cached resolution of picklist field values to
// human-readable labels. It lazily fetches and caches field definitions
// per entity on first access.
//
// Usage:
//
//	pl := autotask.NewPicklist(client)
//	label, err := pl.Resolve(ctx, "Tickets", "status", 5)
//	// label = "Complete"
type Picklist struct {
	client *Client
	mu     sync.Mutex
	// cache maps "EntityTag/fieldName" → (value string → label).
	cache map[string]map[string]string
}

// NewPicklist creates a new picklist resolver backed by the given client.
func NewPicklist(client *Client) *Picklist {
	return &Picklist{
		client: client,
		cache:  map[string]map[string]string{},
	}
}

// Resolve returns the human-readable label for a picklist field value.
// entityPath is the REST path segment (e.g., "/V1.0/Tickets").
// If the value is not found or the field is not a picklist, the numeric
// value is returned as a string.
func (p *Picklist) Resolve(ctx context.Context, entityPath, fieldName string, value int64) (string, error) {
	key := entityPath + "/" + fieldName
	fallback := strconv.FormatInt(value, 10)

	p.mu.Lock()
	if m, ok := p.cache[key]; ok {
		p.mu.Unlock()
		if label, ok := m[fallback]; ok {
			return label, nil
		}
		return fallback, nil
	}
	p.mu.Unlock()

	// Fetch and cache all picklist fields for this entity at once.
	if err := p.loadEntity(ctx, entityPath); err != nil {
		return fallback, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if m, ok := p.cache[key]; ok {
		if label, ok := m[fallback]; ok {
			return label, nil
		}
	}
	return fallback, nil
}

// loadEntity fetches field definitions for the entity and caches all
// picklist mappings. Non-picklist fields get an empty map entry so we
// don't re-fetch.
func (p *Picklist) loadEntity(ctx context.Context, entityPath string) error {
	r := Reader[struct{}]{baseService: baseService{
		client:     p.client,
		entityPath: entityPath,
		entityName: entityPath,
	}}

	fields, err := r.FieldDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("loading picklists for %s: %w", entityPath, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, f := range fields {
		key := entityPath + "/" + f.Name
		if _, already := p.cache[key]; already {
			continue
		}
		m := map[string]string{}
		for _, pv := range f.PicklistValues {
			m[pv.Value] = pv.Label
		}
		p.cache[key] = m
	}

	return nil
}

// ResolveAll returns labels for multiple field/value pairs on the same entity
// in a single cache load. Returns a map of fieldName → label.
func (p *Picklist) ResolveAll(ctx context.Context, entityPath string, fields map[string]int64) (map[string]string, error) {
	result := make(map[string]string, len(fields))
	for fieldName, value := range fields {
		label, err := p.Resolve(ctx, entityPath, fieldName, value)
		if err != nil {
			return nil, err
		}
		result[fieldName] = label
	}
	return result, nil
}
