package internal

import (
	"context"
	"fmt"
	"math/rand"
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

	if err := page.Context(navCtx).WaitLoad(); err != nil {
		return &ScrapeResult{
			Success: false,
			Message: "page load failed",
			Error:   err,
		}, nil
	}

	result := b.clickPlay(page, ctx)
	return result, nil
}

func (b *BrowserClient) clickPlay(page *rod.Page, ctx context.Context) *ScrapeResult {
	actionCtx, actionCancel := context.WithTimeout(ctx, b.actionTimeout)
	defer actionCancel()

	// ponytail: wait for play button to appear; YouTube lazy-loads the player
	selector := "button.ytp-large-play-button"

	el, err := page.Context(actionCtx).Element(selector)
	if err != nil {
		return &ScrapeResult{
			Success: false,
			Message: "play button not found",
			Error:   fmt.Errorf("selector %q: %w", selector, err),
		}
	}

	visible, err := el.Visible()
	if err != nil || !visible {
		return &ScrapeResult{
			Success: false,
			Message: "play button not visible",
			Error:   fmt.Errorf("selector %q: visible=%v, err=%v", selector, visible, err),
		}
	}

	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return &ScrapeResult{
			Success: false,
			Message: "play button click failed",
			Error:   err,
		}
	}

	// ponytail: wait 7-15s randomly for video to start playing after click
	delay := 7 + rand.Intn(9)
	select {
	case <-time.After(time.Duration(delay) * time.Second):
	case <-ctx.Done():
		return &ScrapeResult{
			Success: false,
			Message: "context cancelled while waiting for video",
			Error:   ctx.Err(),
		}
	}

	return &ScrapeResult{
		Success: true,
		Message: fmt.Sprintf("YouTube video play button clicked successfully (waited %ds)", delay),
	}
}
