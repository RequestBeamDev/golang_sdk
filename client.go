// Package requestbeam provides a context-aware client for the RequestBeam API.
package requestbeam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.requestbeam.com"
const Version = "0.1.0"

// Client can be shared by goroutines. Do not modify a supplied transport while in use.
type Client struct {
	apiKey  string
	baseURL string
	locale  string
	timeout time.Duration
	http    *http.Client
}

// Option configures a new client.
type Option func(*Client)

func WithBaseURL(value string) Option        { return func(c *Client) { c.baseURL = value } }
func WithLocale(value string) Option         { return func(c *Client) { c.locale = value } }
func WithTimeout(value time.Duration) Option { return func(c *Client) { c.timeout = value } }

// WithHTTPClient uses a shallow copy of value; its transport is shared. Redirects
// are always disabled, and a client cookie jar is not used for API requests.
func WithHTTPClient(value *http.Client) Option { return func(c *Client) { c.http = value } }

// NewClient reads REQUESTBEAM_API_KEY when apiKey is empty. The base URL must be
// an HTTPS origin, without credentials, a path, query, or fragment.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	c := &Client{apiKey: apiKey, baseURL: DefaultBaseURL, locale: "en", timeout: 60 * time.Second}
	for _, option := range opts {
		if option != nil {
			option(c)
		}
	}
	c.locale = normalizeLocale(c.locale)
	if c.apiKey == "" {
		c.apiKey = os.Getenv("REQUESTBEAM_API_KEY")
	}
	if strings.TrimSpace(c.apiKey) == "" || strings.ContainsAny(c.apiKey, "\r\n") {
		return nil, c.localError("missing_api_key", nil)
	}
	u, err := url.Parse(c.baseURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(c.baseURL, "#") || (u.Path != "" && u.Path != "/") || u.RawPath != "" {
		return nil, c.localError("invalid_base_url", nil)
	}
	c.baseURL = strings.TrimSuffix(c.baseURL, "/")
	if c.timeout <= 0 {
		return nil, c.localError("invalid_timeout", nil)
	}
	h := http.Client{}
	if c.http != nil {
		h = *c.http
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	h.Jar = nil
	c.http = &h
	return c, nil
}

func (c *Client) localError(code string, cause error) error {
	return &SDKError{Code: code, Message: message(c.locale, code), Cause: cause}
}

func (c *Client) requestURL(path string, params url.Values) (string, error) {
	decoded := path
	for i := 0; i < 4; i++ {
		next, err := url.PathUnescape(decoded)
		if err != nil {
			return "", c.localError("invalid_path", nil)
		}
		if next == decoded {
			break
		}
		decoded = next
	}
	if !strings.HasPrefix(decoded, "/") || strings.Contains(decoded, "//") || strings.ContainsAny(decoded, "\\?#%") {
		return "", c.localError("invalid_path", nil)
	}
	for _, ch := range decoded {
		if ch <= 32 || ch == 127 {
			return "", c.localError("invalid_path", nil)
		}
	}
	for _, segment := range strings.Split(decoded, "/") {
		if segment == "." || segment == ".." {
			return "", c.localError("invalid_path", nil)
		}
	}
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any, params url.Values) (*http.Request, error) {
	u, err := c.requestURL(path, params)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, c.localError("invalid_response", nil)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, c.localError("invalid_path", nil)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "requestbeam-go/"+Version)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Request makes one API request. It does not implement retries. Use a relative
// API path and separate query parameters. JSON numbers are decoded as json.Number.
func (c *Client) Request(ctx context.Context, method, path string, body any, params url.Values) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := c.newRequest(ctx, method, path, body, params)
	if err != nil {
		return nil, err
	}
	response, err := c.http.Do(req)
	if err != nil {
		return nil, c.transportError(ctx, err)
	}
	defer response.Body.Close()
	result, err := c.responseBody(response)
	if ctx.Err() != nil {
		return nil, c.transportError(ctx, ctx.Err())
	}
	return result, err
}

func (c *Client) transportError(ctx context.Context, err error) error {
	cause := ctx.Err()
	if cause == nil {
		cause = safeTransportCause(err)
	}
	return c.localError("transport_error", cause)
}

// Retain cancellation classification without exposing URL or credential text.
func safeTransportCause(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	var timeout net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
		return context.DeadlineExceeded
	}
	return nil
}

func decodeObject(reader io.Reader) (map[string]any, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		return nil, err
	}
	if body == nil {
		return nil, io.ErrUnexpectedEOF
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, io.ErrUnexpectedEOF
	}
	return body, nil
}

func (c *Client) responseBody(response *http.Response) (map[string]any, error) {
	body, err := decodeObject(response.Body)
	if cause := safeTransportCause(err); cause != nil {
		return nil, c.localError("transport_error", cause)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, c.apiError(response, body)
	}
	if err != nil {
		return nil, c.localError("invalid_response", nil)
	}
	if isErrorBody(body) {
		return nil, c.apiError(response, body)
	}
	return body, nil
}

func isErrorBody(body map[string]any) bool {
	return body["success"] == false || body["error"] != nil || body["type"] == "error" || body["type"] == "response.error"
}

func (c *Client) apiError(response *http.Response, body map[string]any) *APIError {
	fields, _ := body["error"].(map[string]any)
	if fields == nil {
		if nested, ok := body["response"].(map[string]any); ok {
			fields, _ = nested["error"].(map[string]any)
		}
	}
	if fields == nil {
		fields = body
	}
	msg, _ := fields["message"].(string)
	if msg == "" {
		msg = message(c.locale, "invalid_response")
	}
	requestID, _ := body["requestId"].(string)
	if requestID == "" {
		requestID, _ = body["request_id"].(string)
	}
	if requestID == "" {
		requestID = response.Header.Get("X-Request-ID")
	}
	errorType, _ := fields["type"].(string)
	return &APIError{StatusCode: response.StatusCode, Code: fields["code"], Type: errorType,
		RequestID: requestID, RetryAfter: response.Header.Get("Retry-After"), Body: body, Message: msg}
}
