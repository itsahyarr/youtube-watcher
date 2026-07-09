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
