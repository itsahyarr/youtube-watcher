# YouTube Watcher — Design Spec

**Date**: 2026-07-09
**Status**: Approved
**Scope**: MVP — single endpoint YouTube scrape service

## Overview

Go service that receives a YouTube URL, opens it in a browser via Rod, clicks the play button, logs the result to MongoDB. One request = one browser launch (ephemeral per-request model).

## Tech Stack (Fixed)

- **Language**: Go
- **HTTP**: Gin
- **Browser**: Rod (`headless` per-request via query param, default from env)
- **Database**: MongoDB (`scrape_logs` collection)
- **Config**: `.env` file + env vars (`godotenv`)

## Package Structure (Flat MVP)

```
module github.com/itsahyarr/youtube-watcher

cmd/api/main.go          — entrypoint, wires config → handler, starts Gin
internal/
  config.go              — Config struct from env/.env
  handler.go             — POST /api/v1/scrape/youtube/play
  service.go             — orchestrate: browser → scrape → log → response
  browser.go             — Rod: launch, navigate, click, close (per-request)
  repository.go          — MongoDB: InsertLog(ctx, log) → (id, error)
  model.go               — ScrapeLog struct, JSON response types
  validator.go           — URL validation (YouTube patterns)
.env.example             — env var template (no secrets)
```

~300 lines total. Flat by design — packages split when second endpoint arrives.

## API Contract

### Endpoint

```
POST /api/v1/scrape/youtube/play?headless=false
```

### Request

```json
{
  "url": "https://www.youtube.com/watch?v=VIDEO_ID",
  "proxy": "http://user:pass@proxy:8080"
}
```

- `url` (required) — must match YouTube patterns: `youtube.com/watch`, `youtu.be`
- `proxy` (optional) — client-supplied proxy URL. If absent, direct connection
- `headless` (query param) — `true`/`false`. Defaults to `ROD_HEADLESS` env var

### Success Response (200)

```json
{
  "code": 200,
  "status": "OK",
  "success": true,
  "data": {
    "logId": "...",
    "url": "...",
    "action": "CLICK_PLAY",
    "result": "SUCCESS",
    "message": "YouTube video play button clicked successfully"
  }
}
```

### Error Responses

| Scenario | HTTP | Body |
|---|---|---|
| Missing/invalid JSON | 400 | `{"code":400,"status":"BAD_REQUEST","success":false,"errors":{"message":"..."}}` |
| Non-YouTube URL | 400 | Same pattern |
| Invalid proxy URL/format | 400 | `{"code":400,"status":"BAD_REQUEST","success":false,"errors":{"message":"invalid proxy: ..."}}` |
| Proxy DNS unresolvable | 400 | Same pattern |
| Proxy TCP unreachable (5s dial) | 400 | Same pattern |
| Browser launch fail | 500 | `{"code":500,"status":"INTERNAL_SERVER_ERROR","success":false,"errors":{"message":"..."}}` |
| Navigation timeout (30s) | 500 | Same |
| Play button not found/click fail | 500 | Same |
| Overall request timeout (120s) | 500 | Same |

### MongoDB Write Failure

If click succeeds but MongoDB insert fails: response returns **200 success**. Insert error logged to stderr. The click result is what matters; log persistence is best-effort.

## Data Flow

```
POST /api/v1/scrape/youtube/play
  │
  ▼ handler.go
  ├─ Parse JSON body + query param (headless)
  ├─ Validate URL (YouTube patterns)
  │     └─ fail → 400
  ├─ Validate proxy (format, DNS resolve, TCP dial)
  │     └─ fail → 400
  │
  ▼ service.go (context with 120s deadline)
  ├─ browser.go: launch Rod (headless=param, proxy=param if provided)
  ├─ browser.go: navigate to URL (30s timeout)
  ├─ browser.go: click play button (button.ytp-large-play-button, visibility check)
  │     ├─ success → wait 7-15s randomly for video to start
  │     └─ fail    → result="FAILED"
  │
  ▼ repository.go: InsertLog(ctx, log)
  │     ├─ success → log_id
  │     └─ fail    → log error to stderr
  │
  ▼ handler.go: JSON response
  120s deadline triggers → context cancelled → browser killed
```

## MongoDB Model

Collection: `scrape_logs`
Index: `created_at` (descending)

```go
type ScrapeLog struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    URL        string             `bson:"url" json:"url"`
    Target     string             `bson:"target" json:"target"`       // "YOUTUBE"
    Action     string             `bson:"action" json:"action"`       // "CLICK_PLAY"
    Status     string             `bson:"status" json:"status"`       // "SUCCESS" | "FAILED"
    HTTPStatus int                `bson:"httpStatus" json:"httpStatus"`
    Message    string             `bson:"message" json:"message"`
    Error      *string            `bson:"error,omitempty" json:"error,omitempty"`
    StartedAt  time.Time          `bson:"startedAt" json:"startedAt"`
    FinishedAt time.Time          `bson:"finishedAt" json:"finishedAt"`
    DurationMs int64              `bson:"durationMs" json:"durationMs"`
    ProxyUsed  *string            `bson:"proxyUsed,omitempty" json:"proxyUsed,omitempty"`
    Headless   bool               `bson:"headless" json:"headless"`
    CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
```

## Environment Variables

```env
APP_ENV=development
APP_PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=scraping_service
MONGO_SCRAPE_LOG_COLLECTION=scrape_logs
ROD_HEADLESS=false
ROD_NAVIGATION_TIMEOUT_SECONDS=30
ROD_ACTION_TIMEOUT_SECONDS=15
```

All defaults declared in `.env.example`. Config loaded via `godotenv` (falls back to OS env vars).

## Conventions (from AGENTS.md + Go community)

- **Always** check errors — no unchecked errcheck violations
- **Tagged switch** over cascading `if` chains (QF1003)
- **Ponytail**: minimal scaffolding, flat package, one method per concern, standard library where possible
- **Caveman**: chat prose only, never code or docs
- **JSON keys**: `camelCase` (e.g., `logId`, `httpStatus`, `startedAt`)
- **BSON keys**: `camelCase` (matches JSON keys)
- **Database & collection names**: `snake_case` (e.g., `scraping_service`, `scrape_logs`)
- **URLs/endpoints**: `kebab-case` (e.g., `/api/v1/scrape/youtube/play`)
- **Go fields**: exported `PascalCase`

## Dependencies

```
github.com/gin-gonic/gin
github.com/go-rod/rod
go.mongodb.org/mongo-driver
github.com/joho/godotenv
```

## Proxy Validation

Proxies are validated BEFORE browser launch to fail fast:

1. **Format**: scheme must be `http`, `https`, or `socks5`; host required
2. **DNS resolve**: `net.LookupHost` — host must resolve to at least one address
3. **TCP dial**: `net.DialTimeout` with 5s timeout — port must be reachable

Invalid proxy → `400 BAD_REQUEST`. No browser launched.

## Browser Lifecycle

Per-request model:
- Launch browser → navigate → click → close (`defer close` for cleanup)
- Play button: selector `button.ytp-large-play-button`, visibility check before click
- Post-click: random 7-15s wait for video engine to start playing
- Proxy: validated (format/DNS/TCP) before launch; if valid, passed to `--proxy-server`
- Headless: from query param, falls back to env `ROD_HEADLESS`
- Overall timeout: 120s context — kills browser when deadline hits

## Future Growth

Flat package grows naturally:
- Second endpoint → extract `handler/` package
- Second browser target → extract `browser/` package
- Complex MongoDB operations → extract `repository/` package
- No premature scaffolding
