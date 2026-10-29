# API reference

Import `requestbeam "github.com/RequestBeamDev/golang_sdk"`. Every network method takes `context.Context`; the SDK has no background workers or automatic application-level retries. Unless stated otherwise, responses are `(map[string]any, error)` with native JSON envelopes and `json.Number` values.

Go's standard `net/http` transport may transparently retry eligible transport-level failures; custom transports govern their own retry behavior. The absence of SDK retry logic is not a guarantee of exactly one wire-level attempt.

## Client

```go
NewClient(apiKey string, opts ...Option) (*Client, error)
WithBaseURL(origin string) Option
WithLocale(locale string) Option
WithTimeout(timeout time.Duration) Option
WithHTTPClient(client *http.Client) Option
SupportedLocales() []string
```

Defaults: `https://api.requestbeam.com`, English, 60-second request timeout. The timeout includes the entire response body and streaming duration. Use e.g. `WithTimeout(10*time.Minute)` for longer generation streams. An earlier caller-context deadline or custom HTTP-client timeout still applies. An empty key uses `REQUESTBEAM_API_KEY`. Origins must use HTTPS and contain no credentials, path, query, or fragment. Paths passed to `Request` must be relative API paths beginning with `/`, with no traversal or foreign host. HTTP redirects are not followed. A custom HTTP client is copied; its transport remains caller-owned and shared, and its cookie jar is not used.

## Routes

| Method | HTTP route |
| --- | --- |
| `ListModels(ctx)` | `GET /v1/models` |
| `ChatCompletions(ctx, payload)` | `POST /v1/chat-completions` |
| `Responses(ctx, payload)` | `POST /openai/responses` |
| `Messages(ctx, payload)` | `POST /anthropic/v1/messages` |
| `GenerateImage(ctx, payload)` | `POST /openai/v1/images/generations` |
| `GenerateVideo(ctx, payload)` | `POST /v1/video/seedance` |
| `GetJob(ctx, jobID)` | `GET /v1/jobs/{jobID}` |
| `Similarweb(ctx, operation, params)` | `GET /v1/similarweb/{operation}` |
| `TwitterUserByUsername(ctx, username, params)` | `GET /v1/twitter/users?username=...` |
| `TwitterUser(ctx, userID, params)` | `GET /v1/twitter/users/{userID}` |
| `TwitterUserTweets(ctx, userID, params)` | `GET /v1/twitter/users/{userID}/tweets` |

`payload` is `map[string]any`; `params` is `url.Values` and can be `nil`. IDs are nonempty ASCII decimal strings. Unknown payload fields remain intact. Non-streaming text/image helpers reject `stream: true`.

`Request(ctx context.Context, method, path string, body any, params url.Values)` is the generic escape hatch for same-origin JSON APIs. It preserves application errors and precision, but does not infer model-specific validation or response schemas. Pass `nil` for an absent body/query.

## Streaming

```go
StreamChatCompletions(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error
StreamResponses(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error
StreamMessages(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error
StreamImage(ctx context.Context, payload map[string]any, handle func(SSEEvent) error) error

type SSEEvent struct {
    Event string
    Data  string
    ID    string
}

func (e SSEEvent) JSON() (map[string]any, error)
```

Callbacks run synchronously, in order, while the response body is open. Returning an error stops consumption and returns that error. `JSON()` decodes `Data` as an object using `json.Number`. Chat terminates at `[DONE]`; Responses terminates at `response.completed` or `response.incomplete`; Messages terminates at `message_stop`. A `response.incomplete` event is terminal but does not mean the requested output was completed; inspect the response status and `incomplete_details`. Image streaming requires `image_generation.completed` followed by `[DONE]`. Terminal JSON events are delivered to the callback; `[DONE]` is consumed internally. Error events and unexpected EOF produce errors, not a successful partial result. Each SSE event is limited to 64 MiB to bound buffering; an oversized event fails with a stream error.

## Jobs and pagination

```go
WaitForJob(ctx context.Context, jobID string, timeout, pollInterval time.Duration) (map[string]any, error)
TwitterUserTweetPages(ctx context.Context, userID string, params url.Values, maxPages int, visit func(map[string]any) error) error
```

Durations and page limits must be positive. Polling uses an overall deadline, and context cancellation interrupts both HTTP work and the interval wait. `PENDING`, `CONNECT`, and `PROCESSING` continue; `COMPLETED` returns the complete response; `FAILED` and `CANCELLED` return a job failure. Local cancellation/expiry never cancels the remote job or its billing.

Twitter pagination sends one request per page, follows `meta.next_token`, and rejects invalid or repeated cursors. A callback error stops immediately. Reaching `maxPages` is an intentional bound, not evidence of a complete timeline. Query inputs are not modified.

## Similarweb operations

`SimilarwebOperations() []string` returns a sorted independent copy of the supported operation names. The `/v1/similarweb/` prefix is optional:

```text
keywords/suggestions
keywords/websites
websites/backlinks
websites/keywords
websites/traffic-sources
websites/outgoing-traffic
websites/incoming-referrals
websites/branded-keywords
websites/geography
websites/devices
websites/search-performance
websites/social-traffic
websites/engagement
websites/visits
websites/visits/table
websites/visits/trend
websites/ranks
websites/ranks/country
websites/ranks/global
websites/ranks/category
websites/channels
websites/similar-sites
websites/similar-sites/suggestions
websites/history/visits
websites/history/bounce-rate
websites/history/visit-duration
websites/history/pages-per-visit
```

Only worldwide data is supported: `country=999`. Native pagination and dates are passed unchanged. No generic Similarweb paginator is provided because provider response shapes differ. Check endpoint-specific required query fields in the public RequestBeam docs.

## Errors

API failures include unsuccessful HTTP responses and JSON responses that declare `success: false` or an `error`. `APIError` retains service metadata and original decoded body. Structured bodies may include private data: inspect before logging. Local `SDKError` codes remain English machine identifiers even when their human messages are translated. Use `errors.As` for typed errors and `errors.Is` for context cancellation when applicable. Never retry a request solely because a network error was returned: the server may already have processed it.

See [design notes](design.md) for the source-of-truth files and test boundaries.
