# Design and validation boundaries

The SDK is a separate Go module with no third-party runtime dependencies. It implements the RequestBeam public contracts inspected on 2026-10-29. The Synapse server repository is the source of truth for routes and public languages:

- `wiki/public-api-catalog.json`
- `wiki/en/api-reference/jobs.mdx`
- `wiki/en/api-reference/errors.mdx`
- `wiki/en/api-reference/similarweb/overview.mdx`
- `wiki/en/api-reference/twitter/get-user-tweets.mdx`
- `packages/shared/src/public-api/public-locales.json`

The existing RequestBeam Python SDK provides a companion behavior reference, not a substitute for the server contract. This repository contains public SDK code and documentation only, not server source, deployment configuration, or credentials.

## Decisions

- JSON responses retain their native envelope and use `json.Number` to avoid rounding large numbers. Job and Twitter IDs are decimal strings.
- Each request takes a `context.Context`. A shared client may be used concurrently; callers must not mutate payloads, query maps, or their custom transport while requests use them.
- No application-level retry is automatic, including data GETs that may incur charges. Polling and pagination are explicit bounded operations.
- Images return directly through JSON or SSE. Only video jobs use the Jobs API. Cancelling local work does not cancel remote work or billing.
- Locale selection affects only SDK-owned messages. Provider errors, protocol fields, prompt text, model IDs, and result content remain unchanged.
- HTTPS origins are explicit trust boundaries. Redirects are blocked. Result files are not downloaded automatically; use a separate unauthenticated client for trusted file URLs.
- Tests use local simulated servers or transports. Passing tests and publishing Git commits do not prove model availability, paid API success, deployment, or a versioned release.

## Maintenance

Compare `SupportedLocales()` and the embedded catalog against the source locale registry when languages change. Keep all README translations aligned. API additions should preserve JSON fields and include mock tests for their route, errors, and cancellation behavior. Keep billing-affecting retries opt-in outside this SDK.
