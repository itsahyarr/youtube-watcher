# Product Requirements Document (PRD)

## Product Name
YouTube Scraping Service MVP

## Overview
The YouTube Scraping Service is a backend API service that receives a YouTube URL, opens it using a browser automation engine, and performs a basic interaction by clicking the video play button. The service is built with Golang using the Gin framework, Rod for browser automation, and MongoDB for storing scrape execution logs.

For the MVP, the service is intended for development and experimentation only. The browser should run with `headless=false` so developers can visually inspect the automation flow.

## Goals
- Provide a simple HTTP API to trigger a YouTube page automation task.
- Open a given YouTube video URL in a browser using Rod.
- Click the YouTube play button successfully.
- Store a scrape execution log in MongoDB after each scrape attempt.
- Support local development with visible browser automation.

## Non-Goals
- Full YouTube data extraction.
- Downloading videos, audio, captions, or private content.
- Bypassing authentication, paywalls, bot protection, or access restrictions.
- Large-scale scraping or distributed scraping infrastructure.
- User account login automation.
- Scheduling recurring scrape jobs.
- Production-grade anti-detection or proxy rotation.

## Target Users
- Backend developers testing browser automation workflows.
- Internal engineering team validating Rod-based scraping experiments.
- Developers building a future scraping or rendering service.

## Tech Stack

### Backend
- Language: Golang
- Framework: Gin

### Browser Automation
- Rod
- Browser mode for development: `headless=false`

### Database
- MongoDB
- Used to store scrape execution logs after each scrape attempt.

## MVP Scope
The MVP contains one primary workflow:

1. Client sends a YouTube URL to the API.
2. API validates the request body.
3. Service launches or connects to a browser using Rod.
4. Service navigates to the provided YouTube URL.
5. Service waits for the page to load.
6. Service clicks the video play button.
7. Service records the scrape result into MongoDB.
8. API returns a structured success or error response.

## API Contract

### Base URL
```text
http://localhost:<PORT>
```

### Endpoint
```http
POST /api/v1/scrape/youtube/play
```

### Request Body
```json
{
  "url": "https://www.youtube.com/watch?v=VIDEO_ID"
}
```

### Success Response
```json
{
  "code": 200,
  "status": "OK",
  "success": true,
  "data": {
    "log_id": "64f000000000000000000000",
    "url": "https://www.youtube.com/watch?v=VIDEO_ID",
    "action": "CLICK_PLAY",
    "result": "SUCCESS",
    "message": "YouTube video play button clicked successfully"
  }
}
```

### Error Response
```json
{
  "code": 400,
  "status": "BAD_REQUEST",
  "success": false,
  "errors": {
    "url": "URL is required and must be a valid YouTube URL"
  }
}
```

### Internal Server Error Response
```json
{
  "code": 500,
  "status": "INTERNAL_SERVER_ERROR",
  "success": false,
  "errors": {
    "message": "Failed to execute scrape task"
  }
}
```

## Functional Requirements

### FR-001: Receive Scrape Request
The service must expose an HTTP endpoint that accepts a JSON request body containing a YouTube URL.

Acceptance Criteria:
- Request body must contain `url`.
- Empty URL must return `400 BAD_REQUEST`.
- Invalid URL format must return `400 BAD_REQUEST`.
- Non-YouTube URL must return `400 BAD_REQUEST`.

### FR-002: Validate YouTube URL
The service must validate that the submitted URL belongs to YouTube.

Allowed URL formats for MVP:
- `https://www.youtube.com/watch?v=<video_id>`
- `https://youtube.com/watch?v=<video_id>`
- `https://youtu.be/<video_id>`

Acceptance Criteria:
- Valid YouTube URLs are accepted.
- URLs from other domains are rejected.
- Malformed URLs are rejected.

### FR-003: Open Browser with Rod
The service must use Rod to open a browser session.

Acceptance Criteria:
- Browser launches successfully during local development.
- Browser runs with `headless=false` in development mode.
- Browser page is created for each scrape request or reused safely based on implementation choice.

### FR-004: Navigate to YouTube URL
The service must navigate to the provided YouTube URL.

Acceptance Criteria:
- The page loads without immediate navigation errors.
- The service waits until the page is reasonably ready for interaction.
- Navigation timeout is handled gracefully.

### FR-005: Click Play Button
The service must click the YouTube video play button.

Acceptance Criteria:
- The service detects the video player or play button.
- The service clicks the play button.
- If the video is already playing, the service should treat the action as successful where possible.
- If the play button cannot be found, the service returns an error and logs the failed attempt.

### FR-006: Create Scrape Log in MongoDB
After each scrape attempt, the service must create a log document in MongoDB.

Acceptance Criteria:
- A log is created for successful attempts.
- A log is created for failed attempts.
- The API response includes the log ID if the log is created successfully.
- Log creation failure should be handled explicitly.

## MongoDB Log Schema

Collection name:
```text
scrape_logs
```

Example document:
```json
{
  "_id": "ObjectId",
  "url": "https://www.youtube.com/watch?v=VIDEO_ID",
  "target": "YOUTUBE",
  "action": "CLICK_PLAY",
  "status": "SUCCESS",
  "http_status": 200,
  "message": "YouTube video play button clicked successfully",
  "error": null,
  "started_at": "2026-07-09T07:00:00Z",
  "finished_at": "2026-07-09T07:00:05Z",
  "duration_ms": 5000,
  "metadata": {
    "headless": false,
    "environment": "development"
  },
  "created_at": "2026-07-09T07:00:05Z"
}
```

### Log Status Values
- `SUCCESS`
- `FAILED`

### Target Values
- `YOUTUBE`

### Action Values
- `CLICK_PLAY`

## Environment Configuration

Example `.env` values:

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

## Suggested Project Structure

```text
scraping-service/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   └── youtube_handler.go
│   ├── service/
│   │   └── youtube_scrape_service.go
│   ├── browser/
│   │   └── rod_client.go
│   ├── repository/
│   │   └── scrape_log_repository.go
│   ├── model/
│   │   └── scrape_log.go
│   └── response/
│       └── response.go
├── pkg/
│   └── validator/
│       └── url_validator.go
├── .env
├── go.mod
└── README.md
```

## Service Flow

```text
Client
  ↓
Gin HTTP Handler
  ↓
Request Validation
  ↓
YouTube Scrape Service
  ↓
Rod Browser Client
  ↓
Open YouTube URL
  ↓
Click Play Button
  ↓
Create MongoDB Scrape Log
  ↓
Return API Response
```

## Error Handling

The service should handle these common errors:

| Scenario | HTTP Code | Status |
|---|---:|---|
| Missing URL | 400 | BAD_REQUEST |
| Invalid URL | 400 | BAD_REQUEST |
| Non-YouTube URL | 400 | BAD_REQUEST |
| Browser launch failed | 500 | INTERNAL_SERVER_ERROR |
| Page navigation failed | 500 | INTERNAL_SERVER_ERROR |
| Play button not found | 500 | INTERNAL_SERVER_ERROR |
| Play button click failed | 500 | INTERNAL_SERVER_ERROR |
| MongoDB log creation failed | 500 | INTERNAL_SERVER_ERROR |

## Development Requirements

- Browser must be visible during development.
- `ROD_HEADLESS=false` must be the default for local development.
- Logs should be printed to the console for debugging.
- Each request should produce one MongoDB log document.
- The service should be runnable locally with MongoDB.

## Security and Compliance Considerations

- The service should only accept valid YouTube URLs.
- The service should not support arbitrary websites in the MVP.
- The service should not bypass YouTube access restrictions.
- The service should not download or redistribute YouTube content.
- The service should avoid storing sensitive user data.
- The service should include request timeout protection to prevent hanging browser sessions.

## Performance Considerations

For MVP, performance is not the main priority. However, the service should:

- Apply navigation and action timeouts.
- Avoid leaking browser pages or processes.
- Close pages after each scrape attempt if using a per-request page model.
- Keep MongoDB writes simple and indexed by `created_at` if logs grow.

## Future Enhancements

Potential post-MVP improvements:

- Support headless mode for staging or production.
- Add screenshot capture after clicking play.
- Add video title extraction.
- Add duration and player state detection.
- Add job queue for asynchronous scraping.
- Add retry mechanism.
- Add authentication for API clients.
- Add rate limiting.
- Add structured observability with metrics and tracing.
- Add support for other target websites.
- Add Docker Compose setup for API and MongoDB.

## Success Metrics

The MVP is considered successful when:

- The API accepts a valid YouTube URL.
- Rod opens the YouTube page in a visible browser.
- The service clicks the play button successfully.
- MongoDB stores a scrape log for every attempt.
- The API returns a consistent structured response.

## MVP Acceptance Criteria

- `POST /api/v1/scrape/youtube/play` is implemented.
- Request validation is implemented.
- Rod browser automation is implemented with `headless=false` for development.
- YouTube page navigation works for supported URL formats.
- Play button click is attempted and handled.
- MongoDB scrape log is created after each attempt.
- Success and error responses follow the defined API contract.
