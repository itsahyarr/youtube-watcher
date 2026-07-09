# Tech Stack

## Planned

- **Language**: Go (Golang)
- **HTTP framework**: Gin
- **Browser automation**: Rod (go-rod/rod)
- **Database**: MongoDB (mongo-driver)
- **Config**: env vars via `.env` file

## Module path (expected)

`github.com/<user>/youtube-watcher` or similar — TBD when go.mod is created

## Planned project structure

```
cmd/api/main.go
internal/config/
internal/handler/
internal/service/
internal/browser/
internal/repository/
internal/model/
internal/response/
pkg/validator/
```

## Key env vars

- `APP_PORT` (default 8080)
- `MONGO_URI`, `MONGO_DATABASE`, `MONGO_SCRAPE_LOG_COLLECTION`
- `ROD_HEADLESS=false`, `ROD_NAVIGATION_TIMEOUT_SECONDS`, `ROD_ACTION_TIMEOUT_SECONDS`
