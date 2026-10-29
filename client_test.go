package requestbeam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mockResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func mockClient(t *testing.T, handler roundTripFunc, opts ...Option) *Client {
	t.Helper()
	opts = append([]Option{WithHTTPClient(&http.Client{Transport: handler})}, opts...)
	c, err := NewClient("secret-test-key", opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestGenerationRoutesAndNativePayload(t *testing.T) {
	tests := []struct {
		name, path string
		call       func(*Client, context.Context, map[string]any) (map[string]any, error)
	}{
		{"chat", "/v1/chat-completions", (*Client).ChatCompletions},
		{"responses", "/openai/responses", (*Client).Responses},
		{"messages", "/anthropic/v1/messages", (*Client).Messages},
		{"image", "/openai/v1/images/generations", (*Client).GenerateImage},
		{"video", "/v1/video/seedance", (*Client).GenerateVideo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			payload := map[string]any{"model": "test-model", "prompt": "你好", "unknown": map[string]any{"a": true}}
			c := mockClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != tt.path {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer secret-test-key" {
					t.Error("missing authentication")
				}
				if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
					t.Error("missing JSON content type")
				}
				var got map[string]any
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, payload) {
					t.Errorf("payload changed: %#v", got)
				}
				return mockResponse(200, `{"future_field":"preserved","id":999999999999999999,"nested":{"n":9007199254740993}}`), nil
			})
			result, err := tt.call(c, context.Background(), payload)
			if err != nil {
				t.Fatal(err)
			}
			if result["future_field"] != "preserved" || result["id"] != json.Number("999999999999999999") {
				t.Errorf("native result lost: %#v", result)
			}
			if result["nested"].(map[string]any)["n"] != json.Number("9007199254740993") {
				t.Error("nested number lost precision")
			}
			if calls != 1 {
				t.Errorf("made %d calls", calls)
			}
		})
	}
}

func TestReadRoutes(t *testing.T) {
	tests := []struct {
		name, path string
		call       func(*Client) (map[string]any, error)
	}{
		{"models", "/v1/models", func(c *Client) (map[string]any, error) { return c.ListModels(context.Background()) }},
		{"job", "/v1/jobs/999999999999999999", func(c *Client) (map[string]any, error) { return c.GetJob(context.Background(), "999999999999999999") }},
		{"username", "/v1/twitter/users", func(c *Client) (map[string]any, error) {
			return c.TwitterUserByUsername(context.Background(), "@alice", url.Values{"user.fields": {"public_metrics"}})
		}},
		{"user", "/v1/twitter/users/123", func(c *Client) (map[string]any, error) { return c.TwitterUser(context.Background(), "123", nil) }},
		{"tweets", "/v1/twitter/users/123/tweets", func(c *Client) (map[string]any, error) {
			return c.TwitterUserTweets(context.Background(), "123", url.Values{"max_results": {"5"}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockClient(t, func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" || r.URL.Path != tt.path {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer secret-test-key" {
					t.Error("missing authentication")
				}
				if tt.name == "username" && (r.URL.Query().Get("username") != "@alice" || r.URL.Query().Get("user.fields") != "public_metrics") {
					t.Error("username query lost")
				}
				if tt.name == "tweets" && r.URL.Query().Get("max_results") != "5" {
					t.Error("timeline query lost")
				}
				return mockResponse(200, `{}`), nil
			})
			if _, err := tt.call(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRejectUnsafeURLsBeforeNetwork(t *testing.T) {
	c := mockClient(t, func(r *http.Request) (*http.Response, error) { t.Fatal("unsafe URL reached network"); return nil, nil })
	for _, path := range []string{"https://evil.example/a", "//evil.example/a", "relative", "/x/../a", "/x/./a", "/x/%2e%2e/a", "/x/%252e%252e/a", "/x\\evil", "/x?q=secret", "/x#fragment", "/x//y", "/x\na", "/x/%00", "/x/%25252525252e%25252525252e/a"} {
		t.Run(path, func(t *testing.T) {
			if _, err := c.Request(context.Background(), "GET", path, nil, nil); err == nil {
				t.Error("accepted unsafe path")
			}
		})
	}
	for _, base := range []string{"http://api.requestbeam.com", "https://user:password@api.requestbeam.com", "https://api.requestbeam.com/v1", "https://api.requestbeam.com?key=secret", "https://api.requestbeam.com#fragment", "https://", "https:///missing"} {
		t.Run(base, func(t *testing.T) {
			if _, err := NewClient("key", WithBaseURL(base)); err == nil {
				t.Error("accepted unsafe base URL")
			}
		})
	}
}

func TestConstructorValidationAndEnvironment(t *testing.T) {
	t.Setenv("REQUESTBEAM_API_KEY", "")
	for _, key := range []string{"", " ", "key\r\nx: injected"} {
		if _, err := NewClient(key); err == nil {
			t.Errorf("accepted invalid key %q", key)
		}
	}
	for _, timeout := range []time.Duration{0, -time.Second} {
		if _, err := NewClient("key", WithTimeout(timeout)); err == nil {
			t.Error("accepted nonpositive timeout")
		}
	}
	t.Setenv("REQUESTBEAM_API_KEY", "from-env")
	c, err := NewClient("", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer from-env" {
			t.Error("environment key not used")
		}
		return mockResponse(200, `{}`), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.ListModels(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRedirectNeverFollowedAndInjectedClientUnchanged(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Redirect(w, r, "/redirect-target", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	httpClient := server.Client()
	redirectCalled := false
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { redirectCalled = true; return nil }
	c, err := NewClient("key", WithBaseURL(server.URL), WithHTTPClient(httpClient))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ListModels(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 307 {
		t.Fatalf("redirect error = %#v", err)
	}
	if calls.Load() != 1 || redirectCalled {
		t.Error("followed redirect")
	}
	if httpClient.CheckRedirect == nil {
		t.Error("mutated caller client")
	}
	_ = httpClient.CheckRedirect(nil, nil)
	if !redirectCalled {
		t.Error("replaced caller redirect policy")
	}
}

func TestAPIErrorsPreserveMetadataWithoutRetry(t *testing.T) {
	for _, status := range []int{200, 400, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := mockClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				res := mockResponse(status, `{"success":false,"error":{"code":30000,"type":"upstream_error","message":"original"},"requestId":"req-1"}`)
				res.Header.Set("Retry-After", "10")
				return res, nil
			}, WithLocale("zh-CN"))
			_, err := c.ListModels(context.Background())
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %T %v", err, err)
			}
			if apiErr.StatusCode != status || fmt.Sprint(apiErr.Code) != "30000" || apiErr.Type != "upstream_error" || apiErr.RequestID != "req-1" || apiErr.RetryAfter != "10" || apiErr.Body["success"] != false {
				t.Errorf("lost metadata: %+v", apiErr)
			}
			if !strings.Contains(err.Error(), "original") {
				t.Error("server message was translated or discarded")
			}
			if calls != 1 {
				t.Errorf("retried %d times", calls)
			}
		})
	}
}

func TestFlatErrorsAndHeaderRequestID(t *testing.T) {
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		res := mockResponse(400, `{"code":"bad","message":"native"}`)
		res.Header.Set("X-Request-ID", "header-id")
		return res, nil
	})
	_, err := c.ListModels(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "bad" || apiErr.RequestID != "header-id" || !strings.Contains(err.Error(), "native") {
		t.Fatalf("error metadata = %#v", err)
	}
}

func TestMalformedResponsesAndTransportErrors(t *testing.T) {
	for _, body := range []string{"", "not JSON", "null", "[]", `{} {}`, `{"broken":`} {
		t.Run(body, func(t *testing.T) {
			c := mockClient(t, func(*http.Request) (*http.Response, error) { return mockResponse(200, body), nil })
			if _, err := c.ListModels(context.Background()); err == nil {
				t.Error("accepted invalid object response")
			}
		})
	}
	c := mockClient(t, func(*http.Request) (*http.Response, error) { return nil, errors.New("secret-test-key") })
	if _, err := c.ListModels(context.Background()); err == nil || strings.Contains(err.Error(), "secret-test-key") {
		t.Errorf("unsafe transport error: %v", err)
	}
}

func TestNonStreamValidationAndIdentifierValidation(t *testing.T) {
	c := mockClient(t, func(*http.Request) (*http.Response, error) { t.Fatal("invalid input reached network"); return nil, nil })
	for _, method := range []func(context.Context, map[string]any) (map[string]any, error){c.ChatCompletions, c.Responses, c.Messages, c.GenerateImage} {
		if _, err := method(context.Background(), map[string]any{"stream": true}); err == nil {
			t.Error("accepted stream in JSON method")
		}
	}
	for _, id := range []string{"", "123/../../foo", "-1", "1.0", "１２３", " 123"} {
		if _, err := c.GetJob(context.Background(), id); err == nil {
			t.Errorf("accepted identifier %q", id)
		}
		if _, err := c.TwitterUser(context.Background(), id, nil); err == nil {
			t.Errorf("accepted user identifier %q", id)
		}
	}
	if _, err := c.Request(context.Background(), "POST", "/v1/test", map[string]any{"bad": make(chan int)}, nil); err == nil {
		t.Error("accepted unencodable payload")
	}
}

func TestContextCancellationAndDeadline(t *testing.T) {
	c := mockClient(t, func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.ListModels(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("lost cancellation: %v", err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := c.ListModels(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("lost deadline: %v", err)
	}
}

func TestClientConcurrentRequests(t *testing.T) {
	var calls atomic.Int32
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return mockResponse(200, `{"ok":true}`), nil
	})
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := c.ListModels(context.Background()); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if calls.Load() != 20 {
		t.Errorf("calls=%d", calls.Load())
	}
}

func TestResponseReadTimeoutPreservesClassification(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
				}
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer server.Close()
			c, err := NewClient("key", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithTimeout(50*time.Millisecond))
			if err != nil {
				t.Fatal(err)
			}
			if stream {
				err = c.StreamChatCompletions(context.Background(), nil, func(SSEEvent) error { return nil })
			} else {
				_, err = c.ListModels(context.Background())
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("read deadline classification lost: %T %v", err, err)
			}
		})
	}
}

func TestJSONResponseBodyClosed(t *testing.T) {
	for _, content := range []string{`{"ok":true}`, `invalid`} {
		body := &trackedBody{Reader: strings.NewReader(content)}
		c := mockClient(t, func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
		})
		_, _ = c.ListModels(context.Background())
		if !body.closed {
			t.Error("JSON body not closed")
		}
	}
}
