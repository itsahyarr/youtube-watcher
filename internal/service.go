package internal

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo    *Repository
	browser *BrowserClient
}

func NewService(repo *Repository, browser *BrowserClient) *Service {
	return &Service{repo: repo, browser: browser}
}

func (s *Service) ExecuteScrape(ctx context.Context, targetURL, proxyURL string, headless bool) (*ScrapeLog, error) {
	startedAt := time.Now()

	logEntry := &ScrapeLog{
		URL:       targetURL,
		Target:    "YOUTUBE",
		Action:    "CLICK_PLAY",
		Headless:  headless,
		StartedAt: startedAt,
		CreatedAt: time.Now(),
	}

	if proxyURL != "" {
		logEntry.ProxyUsed = &proxyURL
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := s.browser.Scrape(ctx, targetURL, proxyURL, headless)
	finishedAt := time.Now()
	logEntry.FinishedAt = finishedAt
	logEntry.DurationMs = finishedAt.Sub(startedAt).Milliseconds()

	switch {
	case err != nil:
		logEntry.Status = "FAILED"
		logEntry.HTTPStatus = 500
		msg := err.Error()
		logEntry.Message = msg
		logEntry.Error = &msg
	case result.Success:
		logEntry.Status = "SUCCESS"
		logEntry.HTTPStatus = 200
		logEntry.Message = result.Message
	default:
		logEntry.Status = "FAILED"
		logEntry.HTTPStatus = 500
		logEntry.Message = result.Message
		if result.Error != nil {
			msg := result.Error.Error()
			logEntry.Error = &msg
		}
	}

	if result != nil && result.ExitIP != "" {
		logEntry.ExitIP = &result.ExitIP
	}

	logID, dbErr := s.repo.InsertLog(ctx, logEntry)
	if dbErr != nil {
		log.Printf("ERROR: failed to insert scrape log to MongoDB: %v", dbErr)
	}

	if logID != "" {
		id, parseErr := primitive.ObjectIDFromHex(logID)
		if parseErr != nil {
			log.Printf("ERROR: failed to parse log ID %q: %v", logID, parseErr)
		} else {
			logEntry.ID = id
		}
	}

	if err != nil {
		return logEntry, err
	}

	return logEntry, nil
}
