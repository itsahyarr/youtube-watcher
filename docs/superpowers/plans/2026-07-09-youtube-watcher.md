# YouTube Watcher MVP — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-endpoint Go service that opens a YouTube URL in Rod, clicks play, and logs to MongoDB.

**Architecture:** Flat internal package — config → model/validator → repository/browser → service → handler → main. Per-request browser lifecycle. camelCase JSON/BSON keys, kebab-case URLs, snake_case DB names.

**Tech Stack:** Go 1.22+, Gin, Rod, MongoDB driver, godotenv

## Global Constraints

- JSON keys: camelCase (e.g., `logId`, `httpStatus`, `startedAt`)
- BSON keys: camelCase (matches JSON)
- DB/collection names: snake_case (`scraping_service`, `scrape_logs`)
- URLs/endpoints: kebab-case (`/api/v1/scrape/youtube/play`)
- Go fields: exported PascalCase
- Always check errors — no `_` discards
- Tagged `switch` over cascading `if` (QF1003)
- Dependencies: gin, rod, mongo-driver, godotenv only

---

### Task 1: Scaffold project

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `cmd/api/main.go`

**Interfaces:**
- Produces: Go module `github.com/itsahyarr/youtube-watcher`

- [ ] **Step 1: Initialize Go module and add dependencies**

```bash
cd /Users/itsahyarr/Documents/App/Scripts/youtube-watcher
go mod init github.com/itsahyarr/youtube-watcher
go get github.com/gin-gonic/gin github.com/go-rod/rod go.mongodb.org/mongo-driver/mongo github.com/joho/godotenv
```

- [ ] **Step 2: Run `go mod tidy` to clean up**

```bash
go mod tidy
```

- [ ] **Step 3: Create `.env.example`**

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

- [ ] **Step 4: Create `cmd/api/main.go` skeleton**

```go
package main

import "fmt"

func main() {
	fmt.Println("YouTube Watcher starting...")
}
```

- [ ] **Step 5: Verify builds**

```bash
go build ./cmd/api/
```
Expected: builds without errors (binary at `api` in cwd).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum .env.example cmd/api/main.go
git commit -m "feat: scaffold project with dependencies"
```

---

### Task 2: Config loading

**Files:**
- Create: `internal/config.go`
- Modify: `cmd/api/main.go` (import config)

**Interfaces:**
- Consumes: Go module, godotenv
- Produces: `func LoadConfig() (*Config, error)`, `type Config struct{...}`

- [ ] **Step 1: Write `internal/config.go`**

```go
package internal

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                  string
	AppPort                 string
	MongoURI                string
	MongoDatabase           string
	MongoScrapeLogCollection string
	RodHeadless             bool
	RodNavigationTimeout    int
	RodActionTimeout        int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                  getEnv("APP_ENV", "development"),
		AppPort:                 getEnv("APP_PORT", "8080"),
		MongoURI:                getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:           getEnv("MONGO_DATABASE", "scraping_service"),
		MongoScrapeLogCollection: getEnv("MONGO_SCRAPE_LOG_COLLECTION", "scrape_logs"),
		RodHeadless:             getEnvBool("ROD_HEADLESS", false),
		RodNavigationTimeout:    getEnvInt("ROD_NAVIGATION_TIMEOUT_SECONDS", 30),
		RodActionTimeout:        getEnvInt("ROD_ACTION_TIMEOUT_SECONDS", 15),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fallback
		}
		return b
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
```

- [ ] **Step 2: Update `cmd/api/main.go` to load config**

```go
package main

import (
	"fmt"
	"log"

	"github.com/itsahyarr/youtube-watcher/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("YouTube Watcher starting on port %s...\n", cfg.AppPort)
}
```

- [ ] **Step 3: Verify builds**

```bash
go build ./cmd/api/
```
Expected: builds without errors.

- [ ] **Step 4: Commit**

```bash
git add internal/config.go cmd/api/main.go
git commit -m "feat: add config loading with godotenv"
```

---

### Task 3: Model & response types

**Files:**
- Create: `internal/model.go`

**Interfaces:**
- Produces: `type ScrapeLog struct{...}`, `type SuccessResponse struct{...}`, `type ErrorResponse struct{...}`

- [ ] **Step 1: Write `internal/model.go`**

```go
package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScrapeLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL        string             `bson:"url" json:"url"`
	Target     string             `bson:"target" json:"target"`
	Action     string             `bson:"action" json:"action"`
	Status     string             `bson:"status" json:"status"`
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

type SuccessResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Code    int                    `json:"code"`
	Status  string                 `json:"status"`
	Success bool                   `json:"success"`
	Errors  map[string]interface{} `json:"errors"`
}
```

- [ ] **Step 2: Verify builds**

```bash
go build ./internal/
```
Expected: builds without errors.

- [ ] **Step 3: Commit**

```bash
git add internal/model.go
git commit -m "feat: add ScrapeLog model and response types"
```

---

### Task 4: URL validator

**Files:**
- Create: `internal/validator.go`

**Interfaces:**
- Produces: `func IsValidYouTubeURL(rawURL string) bool`

- [ ] **Step 1: Write `internal/validator.go`**

```go
package internal

import (
	"net/url"
	"strings"
)

func IsValidYouTubeURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Host)

	switch {
	case host == "www.youtube.com" || host == "youtube.com":
		return u.Path == "/watch" && u.Query().Get("v") != ""
	case host == "youtu.be":
		return u.Path != "" && u.Path != "/"
	default:
		return false
	}
}
```

- [ ] **Step 2: Write a quick self-check in `internal/validator_test.go`**

```go
package internal

import "testing"

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://www.youtube.com/watch?v=abc123", true},
		{"https://youtube.com/watch?v=abc123", true},
		{"https://youtu.be/abc123", true},
		{"https://www.google.com", false},
		{"", false},
		{"not-a-url", false},
		{"https://www.youtube.com/", false},
	}
	for _, tt := range tests {
		got := IsValidYouTubeURL(tt.url)
		if got != tt.want {
			t.Errorf("IsValidYouTubeURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/ -run TestIsValidYouTubeURL -v
```
Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add internal/validator.go internal/validator_test.go
git commit -m "feat: add YouTube URL validator"
```

---

### Task 5: MongoDB repository

**Files:**
- Create: `internal/repository.go`
- Modify: `cmd/api/main.go` (connect to MongoDB)

**Interfaces:**
- Consumes: `Config` (MongoURI, MongoDatabase, MongoScrapeLogCollection), `ScrapeLog`
- Produces: `func NewRepository(ctx context.Context, cfg *Config) (*Repository, error)`, `func (r *Repository) InsertLog(ctx context.Context, log *ScrapeLog) (string, error)`, `func (r *Repository) Close(ctx context.Context) error`

- [ ] **Step 1: Write `internal/repository.go`**

```go
package internal

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewRepository(ctx context.Context, cfg *Config) (*Repository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.MongoDatabase)
	collection := db.Collection(cfg.MongoScrapeLogCollection)

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"createdAt": -1},
		Options: options.Index().SetName("idx_created_at_desc"),
	})
	if err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	return &Repository{client: client, collection: collection}, nil
}

func (r *Repository) InsertLog(ctx context.Context, log *ScrapeLog) (string, error) {
	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return "", err
	}

	id := result.InsertedID.(primitive.ObjectID)
	return id.Hex(), nil
}

func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
```

- [ ] **Step 2: Update `cmd/api/main.go` to connect repository**

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/itsahyarr/youtube-watcher/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := internal.NewRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			log.Printf("error closing MongoDB: %v", err)
		}
	}()

	fmt.Printf("YouTube Watcher starting on port %s...\n", cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nshutting down...")
}
```

- [ ] **Step 3: Verify builds**

```bash
go build ./cmd/api/
```
Expected: builds without errors.

- [ ] **Step 4: Commit**

```bash
git add internal/repository.go cmd/api/main.go
git commit -m "feat: add MongoDB repository with InsertLog"
```

---

### Task 6: Rod browser client

**Files:**
- Create: `internal/browser.go`

**Interfaces:**
- Consumes: `Config` (RodNavigationTimeout, RodActionTimeout)
- Produces: `type BrowserClient struct{...}`, `func NewBrowserClient(cfg *Config) *BrowserClient`, `func (b *BrowserClient) Scrape(ctx context.Context, targetURL, proxyURL string, headless bool) (result *ScrapeResult, err error)`, `type ScrapeResult struct{...}`

- [ ] **Step 1: Write `internal/browser.go`**

```go
package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type ScrapeResult struct {
	Success bool
	Message string
	Error   error
}

type BrowserClient struct {
	navigationTimeout time.Duration
	actionTimeout     time.Duration
}

func NewBrowserClient(cfg *Config) *BrowserClient {
	return &BrowserClient{
		navigationTimeout: time.Duration(cfg.RodNavigationTimeout) * time.Second,
		actionTimeout:     time.Duration(cfg.RodActionTimeout) * time.Second,
	}
}

func (b *BrowserClient) Scrape(ctx context.Context, targetURL, proxyURL string, headless bool) (*ScrapeResult, error) {
	l := launcher.New().Headless(headless)

	if proxyURL != "" {
		l = l.Proxy(proxyURL)
	}

	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	navCtx, navCancel := context.WithTimeout(ctx, b.navigationTimeout)
	defer navCancel()

	if err := page.Context(navCtx).Navigate(targetURL); err != nil {
		return &ScrapeResult{
			Success: false,
			Message: "page navigation failed",
			Error:   err,
		}, nil
	}

	page.WaitLoad()

	result := b.clickPlay(page)
	return result, nil
}

func (b *BrowserClient) clickPlay(page *rod.Page) *ScrapeResult {
	actionCtx, actionCancel := context.WithTimeout(context.Background(), b.actionTimeout)
	defer actionCancel()

	selectors := []string{
		".ytp-large-play-button",
		".ytp-play-button",
		"button[aria-label*='Play']",
		"video",
	}

	for _, sel := range selectors {
		el, err := page.Context(actionCtx).Element(sel)
		if err != nil {
			continue
		}

		if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
			continue
		}

		return &ScrapeResult{
			Success: true,
			Message: "YouTube video play button clicked successfully",
		}
	}

	return &ScrapeResult{
		Success: false,
		Message: "play button not found",
		Error:   fmt.Errorf("no play button found with any known selector"),
	}
}
```

- [ ] **Step 2: Verify builds**

```bash
go build ./internal/
```
Expected: builds without errors.

- [ ] **Step 3: Commit**

```bash
git add internal/browser.go
git commit -m "feat: add Rod browser client with play button click"
```

---

### Task 7: Service orchestration

**Files:**
- Create: `internal/service.go`

**Interfaces:**
- Consumes: `BrowserClient.Scrape`, `Repository.InsertLog`, `IsValidYouTubeURL`, `ScrapeLog`, `ScrapeResult`
- Produces: `type Service struct{...}`, `func NewService(cfg *Config, repo *Repository, browser *BrowserClient) *Service`, `func (s *Service) ExecuteScrape(ctx context.Context, url, proxy string, headless bool) (*ScrapeLog, error)`

- [ ] **Step 1: Write `internal/service.go`**

```go
package internal

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Service struct {
	repo    *Repository
	browser *BrowserClient
}

func NewService(repo *Repository, browser *BrowserClient) *Service {
	return &Service{repo: repo, browser: browser}
}

func (s *Service) ExecuteScrape(ctx context.Context, targetURL, proxyURL string, headless bool) (*ScrapeLog, error) {
	if !IsValidYouTubeURL(targetURL) {
		return nil, fmt.Errorf("invalid YouTube URL")
	}

	startedAt := time.Now()

	log := &ScrapeLog{
		URL:       targetURL,
		Target:    "YOUTUBE",
		Action:    "CLICK_PLAY",
		Headless:  headless,
		StartedAt: startedAt,
		CreatedAt: time.Now(),
	}

	if proxyURL != "" {
		log.ProxyUsed = &proxyURL
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := s.browser.Scrape(ctx, targetURL, proxyURL, headless)
	finishedAt := time.Now()
	log.FinishedAt = finishedAt
	log.DurationMs = finishedAt.Sub(startedAt).Milliseconds()

	switch {
	case err != nil:
		log.Status = "FAILED"
		log.HTTPStatus = 500
		msg := err.Error()
		log.Message = msg
		log.Error = &msg
	case result.Success:
		log.Status = "SUCCESS"
		log.HTTPStatus = 200
		log.Message = result.Message
	default:
		log.Status = "FAILED"
		log.HTTPStatus = 500
		log.Message = result.Message
		if result.Error != nil {
			msg := result.Error.Error()
			log.Error = &msg
		}
	}

	logID, dbErr := s.repo.InsertLog(ctx, log)
	if dbErr != nil {
		log.Printf("ERROR: failed to insert scrape log to MongoDB: %v", dbErr)
	}

	if logID != "" {
		log.ID, _ = parseObjectID(logID)
	}

	if err != nil {
		return log, err
	}

	return log, nil
}
```

- [ ] **Step 2: Write `internal/objectid.go` helper**

```go
package internal

import "go.mongodb.org/mongo-driver/bson/primitive"

func parseObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}
```

- [ ] **Step 3: Verify builds**

```bash
go build ./internal/
```
Expected: builds without errors.

- [ ] **Step 4: Commit**

```bash
git add internal/service.go internal/objectid.go
git commit -m "feat: add service orchestration with 120s timeout"
```

---

### Task 8: HTTP handler

**Files:**
- Create: `internal/handler.go`

**Interfaces:**
- Consumes: `Service.ExecuteScrape`, `Config.RodHeadless`, `SuccessResponse`, `ErrorResponse`, `ScrapeLog`
- Produces: `type Handler struct{...}`, `func NewHandler(cfg *Config, svc *Service) *Handler`, `func (h *Handler) ScrapeYouTube(c *gin.Context)`

- [ ] **Step 1: Write `internal/handler.go`**

```go
package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScrapeRequest struct {
	URL   string  `json:"url" binding:"required"`
	Proxy *string `json:"proxy"`
}

type Handler struct {
	cfg *Config
	svc *Service
}

func NewHandler(cfg *Config, svc *Service) *Handler {
	return &Handler{cfg: cfg, svc: svc}
}

func (h *Handler) ScrapeYouTube(c *gin.Context) {
	var req ScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Status:  "BAD_REQUEST",
			Success: false,
			Errors: map[string]interface{}{
				"message": "url is required and must be a valid YouTube URL",
			},
		})
		return
	}

	headless := parseHeadlessParam(c.Query("headless"), h.cfg.RodHeadless)

	if !IsValidYouTubeURL(req.URL) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Status:  "BAD_REQUEST",
			Success: false,
			Errors: map[string]interface{}{
				"message": "url is not a valid YouTube URL",
			},
		})
		return
	}

	proxyURL := ""
	if req.Proxy != nil {
		proxyURL = *req.Proxy
	}

	log, err := h.svc.ExecuteScrape(c.Request.Context(), req.URL, proxyURL, headless)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Status:  "INTERNAL_SERVER_ERROR",
			Success: false,
			Errors: map[string]interface{}{
				"message": err.Error(),
			},
		})
		return
	}

	result := "SUCCESS"
	if log.Status != "SUCCESS" {
		result = "FAILED"
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    200,
		Status:  "OK",
		Success: true,
		Data: map[string]interface{}{
			"logId":   log.ID.Hex(),
			"url":     log.URL,
			"action":  log.Action,
			"result":  result,
			"message": log.Message,
		},
	})
}

func parseHeadlessParam(param string, defaultVal bool) bool {
	switch param {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultVal
	}
}
```

- [ ] **Step 2: Verify builds**

```bash
go build ./internal/
```
Expected: builds without errors.

- [ ] **Step 3: Commit**

```bash
git add internal/handler.go
git commit -m "feat: add Gin HTTP handler for scrape endpoint"
```

---

### Task 9: Wire everything in main

**Files:**
- Modify: `cmd/api/main.go`

**Interfaces:**
- Consumes: `Config`, `Repository`, `BrowserClient`, `Service`, `Handler`

- [ ] **Step 1: Update `cmd/api/main.go` with full wiring**

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsahyarr/youtube-watcher/internal"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := internal.NewRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			log.Printf("error closing MongoDB: %v", err)
		}
	}()

	browser := internal.NewBrowserClient(cfg)
	svc := internal.NewService(repo, browser)
	handler := internal.NewHandler(cfg, svc)

	router := gin.Default()
	api := router.Group("/api/v1/scrape")
	{
		api.POST("/youtube/play", handler.ScrapeYouTube)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		fmt.Printf("YouTube Watcher listening on :%s\n", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nshutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	fmt.Println("server stopped")
}
```

- [ ] **Step 2: Verify builds**

```bash
go build ./cmd/api/
```
Expected: builds without errors.

- [ ] **Step 3: Quick compile check — run tests**

```bash
go test ./internal/ -v
```
Expected: validator test passes.

- [ ] **Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: wire all components, graceful shutdown"
```
