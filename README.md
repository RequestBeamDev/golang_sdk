# RequestBeam Go SDK

[English](README.md) | [Deutsch](docs/README.de.md) | [Türkçe](docs/README.tr.md) | [Italiano](docs/README.it.md) | [Español](docs/README.es.md) | [Français](docs/README.fr.md) | [Português](docs/README.pt-BR.md) | [日本語](docs/README.ja.md) | [한국어](docs/README.ko.md) | [简体中文](docs/README.zh-CN.md) | [繁體中文](docs/README.zh-TW.md) | [Русский](docs/README.ru.md)

A Go client for RequestBeam text, image, video, Similarweb, and Twitter APIs. Includes context cancellation, SSE streaming, bounded job polling and pagination, structured errors, and SDK messages in all 12 public project languages. Uses the Go standard library, with no third-party runtime dependencies.

## Install

Go 1.22 or newer, inside an existing Go module:

```bash
go get github.com/RequestBeamDev/golang_sdk@latest
```

The import path follows this repository. Without a release tag, Go resolves an available commit to a pseudo-version. Pin the version recorded in your `go.mod` for reproducible builds.

```bash
export REQUESTBEAM_API_KEY="YOUR_API_KEY"
export REQUESTBEAM_TEXT_MODEL="YOUR_TEXT_MODEL_ID"
```

## Quick start

Choose a model supported by the target endpoint and your account; `ListModels(ctx)` returns the current catalog, not a guarantee of provider availability. Responses are `map[string]any`, with numbers decoded as `json.Number` to preserve precision. Keep IDs as strings.

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"

    requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("en"))
    if err != nil {
        log.Fatal(err)
    }
    result, err := client.ChatCompletions(context.Background(), map[string]any{
        "model": os.Getenv("REQUESTBEAM_TEXT_MODEL"),
        "messages": []map[string]any{{"role": "user", "content": "Hello"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
        log.Fatal(err)
    }
}
```

## Language and configuration

`WithLocale` changes SDK-owned messages only. It does not translate API fields, prompts, model IDs, server error messages, or returned content. `SupportedLocales()` returns `en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, and `ru`. Aliases include `pt`, `cn`, `zh`, `zh-hans`, `tw`, and `zh-hant`. Case and underscores are normalized; unknown or empty locales fall back to English.

The default API origin is `https://api.requestbeam.com`, without `/v1`. Configure a trusted HTTPS origin with `WithBaseURL`, a request duration with `WithTimeout`, or a custom `*http.Client` with `WithHTTPClient`. The default 60-second timeout covers the entire request and response, including streaming. For longer generation streams, configure e.g. `requestbeam.WithTimeout(10*time.Minute)` and an appropriate caller context; an earlier context or custom HTTP-client deadline still wins. The SDK blocks redirects even for custom clients. Keep API keys on the server or in the local environment, never in Git or browser code.

## Streaming, images, and video

Use `StreamChatCompletions`, `StreamResponses`, `StreamMessages`, or `StreamImage` with a callback. `SSEEvent` exposes `Event`, `Data`, and `ID`; `JSON()` decodes an object with number precision preserved. Payload shapes depend on the protocol, and not every event contains text. `[DONE]` is handled internally. Return a callback error or cancel the context to stop reading. Stream errors and premature endings are reported. Individual SSE events are limited to 64 MiB. A terminal `response.incomplete` ends the stream but is an incomplete result: inspect the response status and `incomplete_details`.

Images return results directly as JSON or SSE and do not use video Jobs. `GenerateVideo` returns `data.jobId` as a string. `WaitForJob(ctx, jobID, 10*time.Minute, 5*time.Second)` returns the complete Jobs response after `COMPLETED`; `FAILED` and `CANCELLED` return `JobFailedError`. A timeout or cancelled context stops local waiting only, not remote generation or billing. Non-streaming helpers reject `stream: true` before sending.

Runnable examples: [text](examples/text/main.go), [stream](examples/stream/main.go), [video](examples/video/main.go), [Similarweb](examples/data/main.go). Running examples sends real requests and may incur charges; tests do not run them.

## Similarweb and Twitter

`Similarweb(ctx, "websites/visits", url.Values{"keys": {"example.com"}, "country": {"999"}})` supports the 27 public operations and preserves native response fields. Similarweb provides worldwide data only (`country=999`); other accepted country values are normalized by the service and do not provide country-specific data. Pass explicit `from` and `to` together when required; the SDK does not rewrite dates. Empty results are not proof of zero traffic.

Twitter helpers cover usernames, user IDs, and timelines. `TwitterUserTweetPages` visits complete pages, follows `meta.next_token`, requires a positive page limit, and rejects repeated cursors. It does not merge results or guarantee a complete timeline. Each data GET/page can incur its own charge. Requests are never retried automatically by SDK code.

## Errors and result files

Use `errors.As` to inspect `*requestbeam.APIError`, including HTTP status, service code/type, request ID, original body, and `Retry-After`. `SDKError` identifies local validation, protocol, and transport errors. `JobFailedError` and `PollingTimeoutError` distinguish terminal job failure from local polling expiry. Server messages remain unchanged. A network error does not prove the server did not process or bill a request; verify before retrying.

The SDK never downloads output files automatically. Inspect result structures before saving them. Result URLs may expire; download trusted files with a separate unauthenticated HTTP client so the RequestBeam key is not forwarded to the file host. Avoid logging full request bodies or responses containing private data.

## Development

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

Tests use local mocks, not paid live API requests. This implementation does not establish live provider availability or perform a production deployment. See the [API reference](docs/api-reference.md) for signatures and protocol details, and [design notes](docs/design.md) for source contracts and validation boundaries.

[RequestBeam](https://requestbeam.com/en) · [Public API documentation](https://requestbeam.com/en/docs/api-reference/models)
