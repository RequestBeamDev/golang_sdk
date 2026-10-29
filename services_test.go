package requestbeam

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestSimilarwebAllOperationsAndNativeQuery(t *testing.T) {
	operations := []string{
		"keywords/suggestions", "websites/backlinks", "websites/keywords", "keywords/websites",
		"websites/traffic-sources", "websites/outgoing-traffic", "websites/incoming-referrals",
		"websites/branded-keywords", "websites/geography", "websites/devices", "websites/search-performance",
		"websites/social-traffic", "websites/engagement", "websites/visits", "websites/visits/table",
		"websites/visits/trend", "websites/ranks", "websites/ranks/country", "websites/ranks/global",
		"websites/ranks/category", "websites/channels", "websites/similar-sites", "websites/similar-sites/suggestions",
		"websites/history/visits", "websites/history/bounce-rate", "websites/history/visit-duration", "websites/history/pages-per-visit",
	}
	if len(SimilarwebOperations()) != 27 {
		t.Fatalf("operation count = %d", len(SimilarwebOperations()))
	}
	params := url.Values{"keys": {"example.com"}, "Page": {"2"}, "country": {"999"}, "from": {"2026-01-01"}, "to": {"2026-03-01"}}
	before := params.Encode()
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			c := mockClient(t, func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" || r.URL.Path != "/v1/similarweb/"+operation {
					t.Errorf("wrong operation: %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Query().Encode() != before {
					t.Errorf("query changed: %s", r.URL.RawQuery)
				}
				return mockResponse(200, `{"Records":[],"TotalCount":0}`), nil
			})
			result, err := c.Similarweb(context.Background(), operation, params)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := result["Records"]; !ok {
				t.Error("native key casing changed")
			}
		})
	}
	if params.Encode() != before {
		t.Error("mutated caller query")
	}
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid operation reached network")
		return nil, nil
	})
	if _, err := c.Similarweb(context.Background(), "../../evil", nil); err == nil {
		t.Error("accepted invalid operation")
	}
}

func TestWaitForJobStatesAndStringIDs(t *testing.T) {
	statuses := []string{"PENDING", "CONNECT", "PROCESSING", "COMPLETED"}
	calls := 0
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/jobs/999999999999999999" {
			t.Error("job ID altered")
		}
		status := statuses[calls]
		calls++
		return mockResponse(200, fmt.Sprintf(`{"success":true,"data":{"id":"999999999999999999","status":%q,"outputResult":{"asset":"url"}}}`, status)), nil
	})
	result, err := c.WaitForJob(context.Background(), "999999999999999999", time.Second, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	job := result["data"].(map[string]any)
	if calls != 4 || job["status"] != "COMPLETED" || job["id"] != "999999999999999999" {
		t.Errorf("job result = %#v calls=%d", result, calls)
	}
}

func TestWaitForJobTerminalFailures(t *testing.T) {
	for _, status := range []string{"FAILED", "CANCELLED", "UNKNOWN"} {
		t.Run(status, func(t *testing.T) {
			calls := 0
			c := mockClient(t, func(*http.Request) (*http.Response, error) {
				calls++
				return mockResponse(200, fmt.Sprintf(`{"data":{"id":"123","status":%q}}`, status)), nil
			})
			_, err := c.WaitForJob(context.Background(), "123", time.Second, time.Millisecond)
			if err == nil {
				t.Fatal("terminal failure ignored")
			}
			if status != "UNKNOWN" {
				var failed *JobFailedError
				if !errors.As(err, &failed) || failed.Job["status"] != status {
					t.Errorf("lost failed job: %v", err)
				}
			}
			if calls != 1 {
				t.Errorf("polled terminal status %d times", calls)
			}
		})
	}
}

func TestWaitForJobTimeoutAndCancellation(t *testing.T) {
	calls := 0
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("request has no poll deadline")
		}
		return mockResponse(200, `{"data":{"status":"PENDING"}}`), nil
	})
	start := time.Now()
	_, err := c.WaitForJob(context.Background(), "123", 20*time.Millisecond, time.Second)
	var timeout *PollingTimeoutError
	if !errors.As(err, &timeout) || timeout.JobID != "123" {
		t.Errorf("timeout = %T %v", err, err)
	}
	if time.Since(start) > 500*time.Millisecond || calls != 1 {
		t.Errorf("poll deadline not bounded, calls=%d", calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c = mockClient(t, func(r *http.Request) (*http.Response, error) {
		cancel()
		return mockResponse(200, `{"data":{"status":"PENDING"}}`), nil
	})
	if _, err = c.WaitForJob(ctx, "123", time.Minute, time.Minute); !errors.Is(err, context.Canceled) {
		t.Errorf("lost polling cancellation: %v", err)
	}
}

func TestWaitDeadlineBoundsInflightRequest(t *testing.T) {
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	start := time.Now()
	_, err := c.WaitForJob(context.Background(), "123", 10*time.Millisecond, time.Second)
	var timeout *PollingTimeoutError
	if !errors.As(err, &timeout) {
		t.Errorf("inflight deadline = %T %v", err, err)
	}
	if time.Since(start) > time.Second {
		t.Error("inflight request exceeded polling deadline")
	}
}

func TestPollingValidationBeforeNetwork(t *testing.T) {
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid poll input reached network")
		return nil, nil
	})
	for _, pair := range [][2]time.Duration{{0, time.Second}, {-1, time.Second}, {time.Second, 0}, {time.Second, -1}} {
		if _, err := c.WaitForJob(context.Background(), "123", pair[0], pair[1]); err == nil {
			t.Error("invalid polling duration accepted")
		}
	}
}

func TestTwitterPaginationBoundsAndNoInputMutation(t *testing.T) {
	params := url.Values{"max_results": {"5"}, "pagination_token": {"initial"}}
	before := params.Encode()
	var seen []string
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.URL.Query().Get("pagination_token"))
		if r.URL.Query().Get("max_results") != "5" {
			t.Error("lost timeline params")
		}
		return mockResponse(200, fmt.Sprintf(`{"data":[],"meta":{"next_token":"t%d"}}`, len(seen))), nil
	})
	visits := 0
	err := c.TwitterUserTweetPages(context.Background(), "123", params, 2, func(page map[string]any) error { visits++; return nil })
	if err != nil {
		t.Fatal(err)
	}
	if visits != 2 || !reflect.DeepEqual(seen, []string{"initial", "t1"}) {
		t.Errorf("pages=%d tokens=%v", visits, seen)
	}
	if params.Encode() != before {
		t.Error("mutated caller params")
	}
}

func TestTwitterPaginationEndsAndCallbackError(t *testing.T) {
	calls := 0
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		calls++
		return mockResponse(200, `{"data":[],"meta":{}}`), nil
	})
	if err := c.TwitterUserTweetPages(context.Background(), "123", nil, 10, func(map[string]any) error { return nil }); err != nil || calls != 1 {
		t.Errorf("end = %v calls=%d", err, calls)
	}
	stop := errors.New("consumer stopped")
	if err := c.TwitterUserTweetPages(context.Background(), "123", nil, 10, func(map[string]any) error { return stop }); !errors.Is(err, stop) {
		t.Errorf("lost callback error: %v", err)
	}
}

func TestTwitterPaginationRejectsMalformedAndRepeatedTokens(t *testing.T) {
	for _, body := range []string{`{"meta":{"next_token":"same"}}`, `{"meta":{"next_token":42}}`, `{"meta":"bad"}`} {
		t.Run(body, func(t *testing.T) {
			calls := 0
			c := mockClient(t, func(*http.Request) (*http.Response, error) { calls++; return mockResponse(200, body), nil })
			err := c.TwitterUserTweetPages(context.Background(), "123", nil, 10, func(map[string]any) error { return nil })
			if err == nil {
				t.Error("accepted malformed or repeated token")
			}
			if calls > 2 {
				t.Errorf("unbounded repeated requests: %d", calls)
			}
		})
	}
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid pagination reached network")
		return nil, nil
	})
	if err := c.TwitterUserTweetPages(context.Background(), "123", nil, 0, func(map[string]any) error { return nil }); err == nil {
		t.Error("accepted zero max pages")
	}
}
