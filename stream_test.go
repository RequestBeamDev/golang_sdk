package requestbeam

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func streamClient(t *testing.T, content string) (*Client, *trackedBody) {
	t.Helper()
	body := &trackedBody{Reader: strings.NewReader(content)}
	c := mockClient(t, func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Error("missing SSE accept header")
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["stream"] != true {
			t.Error("stream flag not set")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream; charset=utf-8"}}, Body: body}, nil
	})
	return c, body
}

func TestSSEMultilineCommentsBOMAndCRLF(t *testing.T) {
	c, body := streamClient(t, "\ufeff: keepalive\r\nid: evt1\r\nevent: chunk\r\ndata: {\"a\":\r\ndata: 9007199254740993}\r\n\r\ndata: [DONE]\r\n\r\n")
	var events []SSEEvent
	payload := map[string]any{"model": "test", "stream": false}
	err := c.StreamChatCompletions(context.Background(), payload, func(event SSEEvent) error { events = append(events, event); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Event != "chunk" || events[0].ID != "evt1" {
		t.Fatalf("events=%#v", events)
	}
	decoded, err := events[0].JSON()
	if err != nil || decoded["a"] != json.Number("9007199254740993") {
		t.Errorf("decoded event=%#v err=%v", decoded, err)
	}
	if payload["stream"] != false {
		t.Error("mutated caller payload")
	}
	if !body.closed {
		t.Error("body not closed")
	}
}

func TestProtocolSpecificStreamTermination(t *testing.T) {
	tests := []struct {
		name, path, content string
		call                func(*Client, context.Context, map[string]any, func(SSEEvent) error) error
	}{
		{"chat", "/v1/chat-completions", "data: {\"choices\":[]}\n\ndata: [DONE]\n\n", (*Client).StreamChatCompletions},
		{"responses", "/openai/responses", "event: response.completed\ndata: {\"type\":\"response.completed\"}\n\n", (*Client).StreamResponses},
		{"incomplete", "/openai/responses", "event: response.incomplete\ndata: {\"type\":\"response.incomplete\"}\n\n", (*Client).StreamResponses},
		{"messages", "/anthropic/v1/messages", "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n", (*Client).StreamMessages},
		{"image", "/openai/v1/images/generations", "event: image_generation.completed\ndata: {\"type\":\"image_generation.completed\"}\n\ndata: [DONE]\n\n", (*Client).StreamImage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockClient(t, func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != tt.path || r.Method != "POST" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				res := mockResponse(200, tt.content)
				res.Header.Set("Content-Type", "text/event-stream")
				return res, nil
			})
			count := 0
			if err := tt.call(c, context.Background(), nil, func(SSEEvent) error { count++; return nil }); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Errorf("events=%d", count)
			}
		})
	}
}

func TestStreamRequiresTerminalEvent(t *testing.T) {
	for _, content := range []string{"", "data: {\"partial\":true}\n\n", "data: {\"partial\":true}", "data: not-json\n\n"} {
		t.Run(content, func(t *testing.T) {
			c, body := streamClient(t, content)
			if err := c.StreamChatCompletions(context.Background(), nil, func(SSEEvent) error { return nil }); err == nil {
				t.Error("accepted malformed/truncated stream")
			}
			if !body.closed {
				t.Error("body not closed after error")
			}
		})
	}
	for _, content := range []string{"data: [DONE]\n\n", "event: image_generation.completed\ndata: {\"type\":\"image_generation.completed\"}\n\n"} {
		c, _ := streamClient(t, content)
		if err := c.StreamImage(context.Background(), nil, func(SSEEvent) error { return nil }); err == nil {
			t.Error("image stream did not require completed plus DONE")
		}
	}
}

func TestSSEErrorsPreserveProviderDetails(t *testing.T) {
	tests := []struct {
		name, content string
		call          func(*Client, context.Context, map[string]any, func(SSEEvent) error) error
	}{
		{"claude", "event: error\ndata: {\"error\":{\"type\":\"upstream_error\",\"code\":\"failed\",\"message\":\"original\"},\"request_id\":\"r1\"}\n\n", (*Client).StreamMessages},
		{"responses", "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"failed\",\"message\":\"original\"}},\"request_id\":\"r1\"}\n\n", (*Client).StreamResponses},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := streamClient(t, tt.content)
			err := tt.call(c, context.Background(), nil, func(SSEEvent) error { t.Error("error delivered as successful event"); return nil })
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Code != "failed" || apiErr.RequestID != "r1" || apiErr.Message != "original" {
				t.Errorf("error=%#v", err)
			}
		})
	}
}

func TestStreamCallbackStopsAndCloses(t *testing.T) {
	c, body := streamClient(t, "data: {\"partial\":true}\n\ndata: {\"partial\":true}\n\ndata: [DONE]\n\n")
	stop := errors.New("stop consuming")
	calls := 0
	err := c.StreamChatCompletions(context.Background(), nil, func(SSEEvent) error { calls++; return stop })
	if !errors.Is(err, stop) || calls != 1 || !body.closed {
		t.Errorf("stop=%v calls=%d closed=%v", err, calls, body.closed)
	}
}

func TestStreamHTTPAndContentTypeErrors(t *testing.T) {
	for _, status := range []int{200, 429} {
		c := mockClient(t, func(*http.Request) (*http.Response, error) {
			res := mockResponse(status, `{"error":{"message":"native error","code":"blocked"}}`)
			res.Header.Set("Content-Type", "application/json")
			return res, nil
		})
		err := c.StreamChatCompletions(context.Background(), nil, func(SSEEvent) error { return nil })
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "blocked" {
			t.Errorf("status %d error=%v", status, err)
		}
	}
	c := mockClient(t, func(*http.Request) (*http.Response, error) { return mockResponse(200, `{"result":true}`), nil })
	if err := c.StreamChatCompletions(context.Background(), nil, func(SSEEvent) error { return nil }); err == nil {
		t.Error("accepted JSON response as SSE")
	}
}

func TestLargeSSEEvent(t *testing.T) {
	text := strings.Repeat("x", 128*1024)
	c, _ := streamClient(t, "data: {\"text\":\""+text+"\"}\n\ndata: [DONE]\n\n")
	count := 0
	if err := c.StreamChatCompletions(context.Background(), nil, func(event SSEEvent) error {
		count++
		value, err := event.JSON()
		if err != nil || value["text"] != text {
			t.Error("large event truncated")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("events=%d", count)
	}
}

type byteReader struct{ text string }

func (r *byteReader) Read(p []byte) (int, error) {
	if len(r.text) == 0 {
		return 0, io.EOF
	}
	p[0] = r.text[0]
	r.text = r.text[1:]
	return 1, nil
}

func TestSSEFragmentedCRLFAndPersistentEventID(t *testing.T) {
	body := &trackedBody{Reader: &byteReader{text: "id: one\r\ndata: {\"n\":1}\r\n\r\ndata: {\"n\":2}\r\rdata: [DONE]\r\r"}}
	c := mockClient(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: body}, nil
	})
	var events []SSEEvent
	err := c.StreamChatCompletions(context.Background(), nil, func(event SSEEvent) error { events = append(events, event); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].ID != "one" || events[1].ID != "one" {
		t.Errorf("events=%#v", events)
	}
}
