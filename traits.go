package autotask

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
)

// baseService holds the shared state for all entity service types.
type baseService struct {
	client     *Client
	entityPath string
	entityName string
}

// Reader provides read operations (Get, Query, QueryIter, Count) for entity type T.
type Reader[T any] struct {
	baseService
}

// Creator provides Create for entity type T.
type Creator[T any] struct {
	baseService
}

// Updater provides Update for entity type T.
type Updater[T any] struct {
	baseService
}

// Patcher provides Patch for entity type T.
type Patcher[T any] struct {
	baseService
}

// Deleter provides Delete for entity type T.
type Deleter[T any] struct {
	baseService
}

// Get retrieves a single entity by ID.
func (r *Reader[T]) Get(ctx context.Context, id int64) (*T, error) {
	resp, err := r.client.doGet(ctx, fmt.Sprintf("%s/%d", r.entityPath, id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if len(resp.Item) == 0 || string(resp.Item) == "null" {
		return nil, ErrNotFound
	}
	var out T
	if err := json.Unmarshal(resp.Item, &out); err != nil {
		return nil, fmt.Errorf("autotask: unmarshaling %s: %w", r.entityName, err)
	}
	return &out, nil
}

// Query collects all pages of results into a slice. It calls QueryIter internally.
func (r *Reader[T]) Query(ctx context.Context, opts ...FilterOption) ([]*T, error) {
	var results []*T
	for item, err := range r.QueryIter(ctx, opts...) {
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// QueryIter returns an iterator over all pages of results. The first page is
// fetched via POST to {entityPath}/query; subsequent pages are fetched via GET
// using the absolute nextPageUrl from pageDetails.
func (r *Reader[T]) QueryIter(ctx context.Context, opts ...FilterOption) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		fq := buildFilterQuery(opts)

		// First page: POST with filter body.
		resp, err := r.client.doPost(ctx, r.entityPath+"/query", fq)
		if err != nil {
			yield(nil, err)
			return
		}

		for {
			for _, raw := range resp.Items {
				var item T
				if err := json.Unmarshal(raw, &item); err != nil {
					yield(nil, fmt.Errorf("autotask: unmarshaling %s: %w", r.entityName, err))
					return
				}
				if !yield(&item, nil) {
					return
				}
			}

			if resp.PageDetails == nil || resp.PageDetails.NextPageUrl == "" {
				return
			}

			// Subsequent pages: GET with absolute nextPageUrl.
			nextURL := resp.PageDetails.NextPageUrl
			resp, err = r.client.doGet(ctx, nextURL)
			if err != nil {
				yield(nil, err)
				return
			}
		}
	}
}

// Count returns the total number of entities matching the given filters.
func (r *Reader[T]) Count(ctx context.Context, opts ...FilterOption) (int64, error) {
	fq := buildFilterQuery(opts)
	resp, err := r.client.doPost(ctx, r.entityPath+"/query/count", fq)
	if err != nil {
		return 0, err
	}
	if resp.QueryCount == nil {
		return 0, fmt.Errorf("autotask: count response missing queryCount field")
	}
	return *resp.QueryCount, nil
}

// Create creates a new entity and returns the assigned ID.
func (c *Creator[T]) Create(ctx context.Context, entity *T) (int64, error) {
	resp, err := c.client.doPost(ctx, c.entityPath, entity)
	if err != nil {
		return 0, err
	}
	// The created entity is returned in resp.Item; extract the id field.
	var result struct {
		ItemID int64 `json:"itemId"`
		ID     int64 `json:"id"`
	}
	if len(resp.Item) > 0 && string(resp.Item) != "null" {
		if err := json.Unmarshal(resp.Item, &result); err != nil {
			return 0, fmt.Errorf("autotask: unmarshaling create response for %s: %w", c.entityName, err)
		}
	}
	if result.ItemID != 0 {
		return result.ItemID, nil
	}
	return result.ID, nil
}

// Update replaces an entity via PUT.
func (u *Updater[T]) Update(ctx context.Context, entity *T) error {
	_, err := u.client.doPut(ctx, u.entityPath, entity)
	return err
}

// Patch applies a partial update to an entity by ID.
func (p *Patcher[T]) Patch(ctx context.Context, id int64, data PatchData) error {
	_, err := p.client.doPatch(ctx, fmt.Sprintf("%s/%d", p.entityPath, id), data)
	return err
}

// Delete removes an entity by ID.
func (d *Deleter[T]) Delete(ctx context.Context, id int64) error {
	_, err := d.client.doDelete(ctx, fmt.Sprintf("%s/%d", d.entityPath, id))
	return err
}
