# AGENTS.md — yt-transcript

YouTube transcript fetcher for Go.

Tool versions pinned in `mise.toml`.

## Architecture

```
transcript.go              Client, NewClient, FetchTranscript, WithHTTPClient
internal/cache/cache.go    In-memory TTL cache for server use
cmd/yt-transcript/main.go  CLI (text, json, srt output)
cmd/server/main.go         HTTP server (/{video_id}, /healthz)
```

Fetch flow: watch page → extract API key from HTML → innertube player API (ANDROID client) → caption track baseUrl → parse XML → segments.

## Server

```
GET /{video_id}[?lang=en][&fmt=text|json|srt]  → transcript
```

Env: `SERVER_HOST` (127.0.0.1), `PORT` (8080), `ALLOWED_ORIGIN` (\*), `BASE_URL` (auto).

Error status codes: 400 (bad request), 404 (no transcript), 502 (upstream failure), 503 (YouTube rate limiting).

## Gotchas

- **Two XML formats**: URLs with `&fmt=srv3` return `<timedtext>` with `<p t d>` (ms); without it returns `<transcript>` with `<text start dur>` (seconds). We strip `fmt=srv3`.
- **HTML entities** in transcript text (`&#39;`, `&amp;`, etc.) are unescaped.
- **No API key needed** — extracted from the watch page at runtime.
- **429 / bot detection** detected in all three HTTP calls; returned as clear errors.

## Tests

```bash
go test -short ./...       # unit (no network)
go test -run Live ./...    # integration (live YouTube calls)
```

Live tests are prefixed `TestLive_`; skipped with `-short`.
