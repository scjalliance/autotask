package autotask

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultZoneInfoURL     = "https://webservices2.autotask.net/ATServicesRest/V1.0/ZoneInformation"
	defaultRateLimit       = 10 // requests per second threshold before backing off
	rateLimitBackoffWindow = time.Second
)

// Config holds the configuration for the Autotask API client.
type Config struct {
	// Username is the Autotask API user's email address. Required.
	Username string
	// Secret is the Autotask API user's secret key. Required.
	Secret string
	// IntegrationCode is the API integration code. Required.
	IntegrationCode string
	// BaseURL overrides the API base URL (useful for testing or manual zone specification).
	// When empty, zone detection is performed automatically.
	BaseURL string
	// HTTPClient allows injecting a custom http.Client. Defaults to http.DefaultClient.
	HTTPClient *http.Client
	// RateLimitThreshold is the number of requests per second before the client
	// automatically backs off. Defaults to 10.
	RateLimitThreshold int
	// DisableRateLimitTracking disables automatic rate limit tracking and backoff.
	DisableRateLimitTracking bool
	// MaxRetries is the number of times to retry on transient failures (5xx,
	// network errors). Defaults to 3. Set to 0 to disable retries.
	MaxRetries int
}

type impersonationKey struct{}

// WithImpersonation returns a new context that carries the given Autotask resource ID
// for impersonation. The client will include it as the ImpersonationResourceId header.
func WithImpersonation(ctx context.Context, resourceID int64) context.Context {
	return context.WithValue(ctx, impersonationKey{}, resourceID)
}

// Client is an authenticated HTTP client for the Autotask PSA REST API.
type Client struct {
	entityServiceFields // generated entity services

	config  Config
	http    *http.Client
	baseURL string

	zoneMu sync.Mutex // protects baseURL during lazy zone detection (unused after init)

	// rate limiting
	rateMu     sync.Mutex
	rateWindow time.Time
	rateCount  int
}

// NewClient creates and returns a new Client. Zone detection is performed during
// construction unless Config.BaseURL is provided.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Username == "" {
		return nil, fmt.Errorf("autotask: Username is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("autotask: Secret is required")
	}
	if cfg.IntegrationCode == "" {
		return nil, fmt.Errorf("autotask: IntegrationCode is required")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if cfg.RateLimitThreshold <= 0 {
		cfg.RateLimitThreshold = defaultRateLimit
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	c := &Client{
		config: cfg,
		http:   httpClient,
	}

	if cfg.BaseURL != "" {
		c.baseURL = strings.TrimRight(cfg.BaseURL, "/")
	} else {
		base, err := resolveBaseURL(cfg.Username, cfg.Secret, cfg.IntegrationCode, httpClient)
		if err != nil {
			return nil, fmt.Errorf("autotask: zone detection failed: %w", err)
		}
		c.baseURL = base
	}

	c.initServices()
	return c, nil
}

// resolveBaseURL performs zone detection by calling the ZoneInformation endpoint.
func resolveBaseURL(username, secret, integrationCode string, httpClient *http.Client) (string, error) {
	url := fmt.Sprintf("%s?user=%s", defaultZoneInfoURL, username)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("UserName", username)
	req.Header.Set("Secret", secret)
	req.Header.Set("ApiIntegrationcode", integrationCode)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("zone information returned HTTP %d", resp.StatusCode)
	}

	var zoneResp struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&zoneResp); err != nil {
		return "", fmt.Errorf("decoding zone response: %w", err)
	}
	if zoneResp.URL == "" {
		return "", fmt.Errorf("zone response contained empty URL")
	}

	return strings.TrimRight(zoneResp.URL, "/"), nil
}

// setAuthHeaders applies authentication and optional impersonation headers to r.
func (c *Client) setAuthHeaders(ctx context.Context, r *http.Request) {
	r.Header.Set("UserName", c.config.Username)
	r.Header.Set("Secret", c.config.Secret)
	r.Header.Set("ApiIntegrationcode", c.config.IntegrationCode)
	r.Header.Set("Content-Type", "application/json")

	if id, ok := ctx.Value(impersonationKey{}).(int64); ok {
		r.Header.Set("ImpersonationResourceId", fmt.Sprintf("%d", id))
	}
}

// trackRequest enforces client-side rate limiting. It increments the in-window
// counter and, if the threshold is exceeded, sleeps until the next window.
// The mutex is released before sleeping to avoid holding it during the delay.
func (c *Client) trackRequest() {
	if c.config.DisableRateLimitTracking {
		return
	}

	c.rateMu.Lock()

	now := time.Now()
	if now.After(c.rateWindow) {
		// Start a new 1-second window.
		c.rateWindow = now.Add(rateLimitBackoffWindow)
		c.rateCount = 0
	}

	c.rateCount++
	shouldSleep := c.rateCount >= c.config.RateLimitThreshold
	var sleepUntil time.Time
	if shouldSleep {
		sleepUntil = c.rateWindow
		// Reset for the next window so subsequent callers don't pile up.
		c.rateWindow = sleepUntil.Add(rateLimitBackoffWindow)
		c.rateCount = 0
	}

	c.rateMu.Unlock() // release before sleeping

	if shouldSleep {
		delay := time.Until(sleepUntil)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
}

// apiResponse is the envelope returned by all Autotask REST endpoints.
type apiResponse struct {
	Item        json.RawMessage   `json:"item"`
	Items       []json.RawMessage `json:"items"`
	PageDetails *pageDetails      `json:"pageDetails"`
	Errors      []string          `json:"errors"`
	QueryCount  *int64            `json:"queryCount"`

	// RawBody holds the full unparsed response body for endpoints that return
	// data at the root level rather than nested under item/items (e.g. entity
	// information endpoints).
	RawBody json.RawMessage `json:"-"`
}

type pageDetails struct {
	Count         int    `json:"count"`
	RequestCount  int    `json:"requestCount"`
	PrevPageUrl   string `json:"prevPageUrl"`
	NextPageUrl   string `json:"nextPageUrl"`
}

// doRequest is the central HTTP dispatch method. path may be a relative path
// (e.g. "/V1.0/Tickets") or an absolute URL (e.g. the nextPageUrl from pagination).
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*apiResponse, error) {
	c.trackRequest()

	// Marshal body once so it can be replayed on retries.
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("autotask: marshaling request body: %w", err)
		}
		bodyBytes = b
	}

	// Support absolute URLs (e.g. pagination nextPageUrl) as-is.
	var fullURL string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		fullURL = path
	} else {
		fullURL = c.baseURL + path
	}

	var lastErr error
	for attempt := range c.config.MaxRetries {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s, 2s, ...
			delay := time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("autotask: building request: %w", err)
		}
		c.setAuthHeaders(ctx, req)

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("autotask: executing request: %w", err)
			continue // network error — retry
		}

		rawBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("autotask: reading response body: %w", err)
			continue
		}

		// 5xx → transient server error, retry
		if resp.StatusCode >= 500 {
			lastErr = &APIError{StatusCode: resp.StatusCode, Errors: []string{string(rawBody)}}
			continue
		}

		var apiResp apiResponse
		if len(rawBody) > 0 {
			if err := json.Unmarshal(rawBody, &apiResp); err != nil {
				return nil, fmt.Errorf("autotask: decoding response: %w", err)
			}
			apiResp.RawBody = json.RawMessage(rawBody)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Errors:     apiResp.Errors,
			}
		}

		return &apiResp, nil
	}

	return nil, lastErr
}

// doGet performs a GET request against path (relative or absolute).
func (c *Client) doGet(ctx context.Context, path string) (*apiResponse, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// doPost performs a POST request with body serialized as JSON.
func (c *Client) doPost(ctx context.Context, path string, body any) (*apiResponse, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// doPut performs a PUT request with body serialized as JSON.
func (c *Client) doPut(ctx context.Context, path string, body any) (*apiResponse, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// doPatch performs a PATCH request with body serialized as JSON.
func (c *Client) doPatch(ctx context.Context, path string, body any) (*apiResponse, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

// doDelete performs a DELETE request against path.
func (c *Client) doDelete(ctx context.Context, path string) (*apiResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}