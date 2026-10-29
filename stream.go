package requestbeam

import (
	"bufio"
	"bytes"
	"context"
	"mime"
	"net/http"
	"strings"
)

// SSEEvent preserves one server-sent event. ID persists across events per SSE.
type SSEEvent struct {
	Event string
	Data  string
	ID    string
}

// JSON decodes an event object without converting numeric IDs to floating point.
func (e SSEEvent) JSON() (map[string]any, error) {
	return decodeObject(strings.NewReader(e.Data))
}

func (c *Client) StreamChatCompletions(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error {
	return c.stream(ctx, "/v1/chat-completions", "chat", payload, handle)
}

func (c *Client) StreamResponses(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error {
	return c.stream(ctx, "/openai/responses", "responses", payload, handle)
}

func (c *Client) StreamMessages(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error {
	return c.stream(ctx, "/anthropic/v1/messages", "messages", payload, handle)
}

func (c *Client) StreamImage(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error {
	return c.stream(ctx, "/openai/v1/images/generations", "image", payload, handle)
}

func (c *Client) stream(ctx context.Context, path, kind string, payload map[string]any, handle func(SSEEvent) error) error {
	if handle == nil {
		return c.localError("invalid_stream", nil)
	}
	copy := make(map[string]any, len(payload)+1)
	for key, value := range payload {
		copy[key] = value
	}
	copy["stream"] = true
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := c.newRequest(ctx, http.MethodPost, path, copy, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	response, err := c.http.Do(req)
	if err != nil {
		return c.transportError(ctx, err)
	}
	defer response.Body.Close()
	mediaType, _, _ := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if response.StatusCode < 200 || response.StatusCode >= 300 || !strings.EqualFold(mediaType, "text/event-stream") {
		_, err := c.responseBody(response)
		if ctx.Err() != nil {
			return c.transportError(ctx, ctx.Err())
		}
		if err != nil {
			return err
		}
		return c.localError("invalid_response", nil)
	}
	return c.events(ctx, response, kind, handle)
}

// An event is bounded to avoid unbounded buffering from a malformed stream.
const maxEventBytes = 64 * 1024 * 1024

func (c *Client) events(ctx context.Context, response *http.Response, kind string, handle func(SSEEvent) error) error {
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), maxEventBytes)
	scanner.Split(splitSSELines)
	var event SSEEvent
	var data strings.Builder
	hasData, firstLine, imageCompleted := false, true, false
	for scanner.Scan() {
		if ctx.Err() != nil {
			return c.transportError(ctx, ctx.Err())
		}
		line := scanner.Text()
		if firstLine {
			line = strings.TrimPrefix(line, "\ufeff")
			firstLine = false
		}
		if line == "" {
			if hasData {
				if event.Event == "" {
					event.Event = "message"
				}
				event.Data = strings.TrimSuffix(data.String(), "\n")
				terminal, err := c.streamEvent(response, kind, event)
				if err != nil {
					return err
				}
				if kind == "image" {
					if event.Data == "[DONE]" {
						if !imageCompleted {
							return c.localError("stream_error", nil)
						}
					} else {
						imageCompleted = imageCompleted || terminal
						terminal = false
					}
				}
				if event.Data != "[DONE]" {
					if err := handle(event); err != nil {
						return err
					}
				}
				if ctx.Err() != nil {
					return c.transportError(ctx, ctx.Err())
				}
				if terminal {
					return nil
				}
			}
			event.Event, event.Data = "", ""
			data.Reset()
			hasData = false
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "data":
			if data.Len()+len(value)+1 > maxEventBytes {
				return c.localError("stream_error", nil)
			}
			data.WriteString(value)
			data.WriteByte('\n')
			hasData = true
		case "event":
			event.Event = value
		case "id":
			if !strings.ContainsRune(value, 0) {
				event.ID = value
			}
		}
	}
	if ctx.Err() != nil {
		return c.transportError(ctx, ctx.Err())
	}
	if err := scanner.Err(); err != nil {
		return c.transportError(ctx, err)
	}
	// SSE dispatch requires a blank line; EOF never proves successful completion.
	return c.localError("stream_error", nil)
}

func (c *Client) streamEvent(response *http.Response, kind string, event SSEEvent) (bool, error) {
	if event.Data == "[DONE]" {
		if event.Event == "error" || event.Event == "response.error" {
			return false, c.localError("invalid_response", nil)
		}
		if kind == "chat" || kind == "image" {
			return true, nil
		}
		return false, c.localError("stream_error", nil)
	}
	body, err := event.JSON()
	if err != nil {
		return false, c.localError("invalid_response", nil)
	}
	if event.Event == "error" || event.Event == "response.error" || isErrorBody(body) {
		return false, c.apiError(response, body)
	}
	eventType, ok := body["type"].(string)
	if !ok || eventType == "" {
		eventType = event.Event
	}
	switch kind {
	case "responses":
		if eventType == "response.failed" {
			return false, c.apiError(response, body)
		}
		return eventType == "response.completed" || eventType == "response.incomplete", nil
	case "messages":
		return eventType == "message_stop", nil
	case "image":
		return eventType == "image_generation.completed", nil
	}
	return false, nil
}

// SSE accepts LF, CRLF, and bare CR. Delay a trailing CR until its possible LF arrives.
func splitSSELines(data []byte, atEOF bool) (int, []byte, error) {
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\r' {
			if i+1 == len(data) && !atEOF {
				return 0, nil, nil
			}
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[:i], nil
			}
		}
		return i + 1, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}
