package requestbeam

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func (c *Client) ListModels(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/v1/models", nil, nil)
}

func (c *Client) nonStream(ctx context.Context, path string, payload map[string]any) (map[string]any, error) {
	if value, exists := payload["stream"]; exists && value != nil && value != false {
		return nil, c.localError("invalid_stream", nil)
	}
	return c.Request(ctx, http.MethodPost, path, payload, nil)
}

func (c *Client) ChatCompletions(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.nonStream(ctx, "/v1/chat-completions", payload)
}

func (c *Client) Responses(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.nonStream(ctx, "/openai/responses", payload)
}

func (c *Client) Messages(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.nonStream(ctx, "/anthropic/v1/messages", payload)
}

func (c *Client) GenerateImage(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.nonStream(ctx, "/openai/v1/images/generations", payload)
}

func (c *Client) GenerateVideo(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/v1/video/seedance", payload, nil)
}

func validIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func (c *Client) GetJob(ctx context.Context, id string) (map[string]any, error) {
	if !validIdentifier(id) {
		return nil, c.localError("invalid_job_id", nil)
	}
	return c.Request(ctx, http.MethodGet, "/v1/jobs/"+id, nil, nil)
}

// WaitForJob polls within timeout and the caller's context. Stopping this method
// does not cancel the remote job or any charges. Each request also uses WithTimeout.
func (c *Client) WaitForJob(ctx context.Context, id string, timeout, pollInterval time.Duration) (map[string]any, error) {
	if !validIdentifier(id) {
		return nil, c.localError("invalid_job_id", nil)
	}
	if timeout <= 0 {
		return nil, c.localError("invalid_timeout", nil)
	}
	if pollInterval <= 0 {
		return nil, c.localError("invalid_poll_interval", nil)
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	pollError := func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &PollingTimeoutError{Message: message(c.locale, "poll_timeout"), JobID: id, Cause: pollCtx.Err()}
	}
	for {
		if pollCtx.Err() != nil {
			return nil, pollError()
		}
		body, err := c.GetJob(pollCtx, id)
		if pollCtx.Err() != nil {
			return nil, pollError()
		}
		if err != nil {
			return nil, err
		}
		job, ok := body["data"].(map[string]any)
		if !ok {
			return nil, c.localError("invalid_response", nil)
		}
		switch job["status"] {
		case "COMPLETED":
			return body, nil
		case "FAILED", "CANCELLED":
			return nil, &JobFailedError{Message: message(c.locale, "job_failed"), Job: job}
		case "PENDING", "CONNECT", "PROCESSING":
		default:
			return nil, c.localError("invalid_response", nil)
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-pollCtx.Done():
			timer.Stop()
			return nil, pollError()
		case <-timer.C:
		}
	}
}

var similarwebOperations = map[string]struct{}{
	"keywords/suggestions": {}, "websites/backlinks": {}, "websites/keywords": {},
	"keywords/websites": {}, "websites/traffic-sources": {}, "websites/outgoing-traffic": {},
	"websites/incoming-referrals": {}, "websites/branded-keywords": {}, "websites/geography": {},
	"websites/devices": {}, "websites/search-performance": {}, "websites/social-traffic": {},
	"websites/engagement": {}, "websites/visits": {}, "websites/visits/table": {},
	"websites/visits/trend": {}, "websites/ranks": {}, "websites/ranks/country": {},
	"websites/ranks/global": {}, "websites/ranks/category": {}, "websites/channels": {},
	"websites/similar-sites": {}, "websites/similar-sites/suggestions": {},
	"websites/history/visits": {}, "websites/history/bounce-rate": {},
	"websites/history/visit-duration": {}, "websites/history/pages-per-visit": {},
}

// SimilarwebOperations returns the supported native operation names, sorted.
func SimilarwebOperations() []string {
	operations := make([]string, 0, len(similarwebOperations))
	for operation := range similarwebOperations {
		operations = append(operations, operation)
	}
	sort.Strings(operations)
	return operations
}

// Similarweb preserves query parameters, including dates and native pagination.
// Current RequestBeam Similarweb data is worldwide only: use country=999.
func (c *Client) Similarweb(ctx context.Context, operation string, params url.Values) (map[string]any, error) {
	operation = strings.TrimPrefix(operation, "/v1/similarweb/")
	if _, ok := similarwebOperations[operation]; !ok {
		return nil, c.localError("invalid_path", nil)
	}
	return c.Request(ctx, http.MethodGet, "/v1/similarweb/"+operation, nil, params)
}

func cloneValues(values url.Values) url.Values {
	copy := make(url.Values, len(values))
	for key, values := range values {
		copy[key] = append([]string(nil), values...)
	}
	return copy
}

func (c *Client) TwitterUserByUsername(ctx context.Context, username string, params url.Values) (map[string]any, error) {
	query := cloneValues(params)
	query.Set("username", username)
	return c.Request(ctx, http.MethodGet, "/v1/twitter/users", nil, query)
}

func (c *Client) TwitterUser(ctx context.Context, id string, params url.Values) (map[string]any, error) {
	if !validIdentifier(id) {
		return nil, c.localError("invalid_job_id", nil)
	}
	return c.Request(ctx, http.MethodGet, "/v1/twitter/users/"+id, nil, params)
}

func (c *Client) TwitterUserTweets(ctx context.Context, id string, params url.Values) (map[string]any, error) {
	if !validIdentifier(id) {
		return nil, c.localError("invalid_job_id", nil)
	}
	return c.Request(ctx, http.MethodGet, "/v1/twitter/users/"+id+"/tweets", nil, params)
}

// TwitterUserTweetPages visits at most maxPages responses, returning immediately
// on a callback error. Each page is a separate potentially billable request.
func (c *Client) TwitterUserTweetPages(ctx context.Context, id string, params url.Values, maxPages int, visit func(map[string]any) error) error {
	if maxPages < 1 || visit == nil {
		return c.localError("invalid_pagination", nil)
	}
	query := cloneValues(params)
	seen := map[string]bool{}
	if token := query.Get("pagination_token"); token != "" {
		seen[token] = true
	}
	for page := 0; page < maxPages; page++ {
		body, err := c.TwitterUserTweets(ctx, id, query)
		if err != nil {
			return err
		}
		if err := visit(body); err != nil {
			return err
		}
		meta, ok := body["meta"].(map[string]any)
		if !ok {
			if _, exists := body["meta"]; exists {
				return c.localError("invalid_response", nil)
			}
			return nil
		}
		if meta["next_token"] == nil || meta["next_token"] == "" {
			return nil
		}
		token, ok := meta["next_token"].(string)
		if !ok || seen[token] {
			return c.localError("invalid_pagination", nil)
		}
		seen[token] = true
		query.Set("pagination_token", token)
	}
	return nil
}
