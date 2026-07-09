package internal

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// ponytail: a realistic desktop Chrome UA/headers combo; YouTube's bot check
// weighs an "automated" fingerprint (navigator.webdriver, missing UA, etc.)
// heavily on top of IP reputation, so we mask both.
const desktopUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

func isContextDestroyedErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "context was destroyed")
}

type ScrapeResult struct {
	Success bool
	Message string
	Error   error
	ExitIP  string
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
	l := launcher.New().
		Headless(headless).
		// ponytail: rod sets "enable-automation" by default, which flips
		// navigator.webdriver to true and shows the "controlled by automated
		// software" infobar -- both are classic bot-detection tells.
		Delete("enable-automation").
		Set("disable-blink-features", "AutomationControlled")

	var proxyUser, proxyPass string
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err == nil {
			if u.User != nil {
				proxyUser = u.User.Username()
				proxyPass, _ = u.User.Password()
			}
			// ponytail: strip auth from --proxy-server; MustHandleAuth handles it via CDP
			u.User = nil
			l = l.Set("proxy-server", u.String())
		} else {
			l = l.Set("proxy-server", proxyURL)
		}
	}

	debugURL := l.MustLaunch()
	browser := rod.New().ControlURL(debugURL).MustConnect()
	defer browser.MustClose()

	if proxyUser != "" {
		// ponytail: MustHandleAuth returns a blocking wait func that must run
		// as a goroutine, otherwise Chrome's Fetch.authRequired event is never
		// answered and every proxied navigation hangs until it times out.
		go browser.MustHandleAuth(proxyUser, proxyPass)()
	}

	// ponytail: stealth.Page injects JS shims (navigator.webdriver, chrome
	// runtime, plugins, permissions, etc.) that puppeteer-extra's stealth
	// evasions use to hide the automated fingerprint from bot checks.
	page, err := stealth.Page(browser)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer func() {
		if closeErr := page.Close(); closeErr != nil {
			fmt.Printf("WARN: failed to close page: %v\n", closeErr)
		}
	}()

	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      desktopUserAgent,
		AcceptLanguage: "en-US,en;q=0.9",
	}); err != nil {
		return nil, fmt.Errorf("failed to set user agent: %w", err)
	}

	navCtx, navCancel := context.WithTimeout(ctx, b.navigationTimeout)
	defer navCancel()

	// ponytail: capture the real exit IP the proxy assigned for the actual
	// document request, straight off CDP's Network.responseReceived event --
	// no extra request needed, and (unlike a separate probe request) it's
	// guaranteed to be the IP the target site actually saw, which matters
	// since our proxy rotates the exit IP per connection.
	var exitIP string
	waitIP := page.Context(navCtx).EachEvent(func(e *proto.NetworkResponseReceived) bool {
		if e.Type == proto.NetworkResourceTypeDocument && e.Response.RemoteIPAddress != "" {
			exitIP = e.Response.RemoteIPAddress
			return true
		}
		return false
	})

	if err := page.Context(navCtx).Navigate(targetURL); err != nil {
		return &ScrapeResult{
			Success: false,
			Message: "page navigation failed",
			Error:   err,
			ExitIP:  exitIP,
		}, nil
	}

	waitIP()

	// ponytail: youtu.be short links 30x-redirect to youtube.com/watch; WaitLoad's
	// polling eval can race that redirect and get "execution context was destroyed",
	// which is transient. Retry a few times instead of failing the whole scrape.
	var waitErr error
	for attempt := 0; attempt < 3; attempt++ {
		waitErr = page.Context(navCtx).WaitLoad()
		if waitErr == nil || !isContextDestroyedErr(waitErr) {
			break
		}
	}
	if waitErr != nil {
		return &ScrapeResult{
			Success: false,
			Message: "page load failed",
			Error:   waitErr,
			ExitIP:  exitIP,
		}, nil
	}

	if botResult := detectBotWall(page); botResult != nil {
		botResult.ExitIP = exitIP
		return botResult, nil
	}

	result := b.clickPlay(page, ctx)
	result.ExitIP = exitIP
	return result, nil
}

// ponytail: surfaces Google/YouTube's bot-check pages ("unusual traffic" /
// "confirm you're not a bot") as a distinct, actionable result instead of
// letting the caller misread it as a generic "play button not visible".
// This is almost always caused by the proxy's exit IP being flagged --
// most commonly because a *rotating* proxy hands out a different exit IP
// per TCP connection, so one page load looks like many different sessions.
func detectBotWall(page *rod.Page) *ScrapeResult {
	res, err := page.Eval(`() => document.body ? document.body.innerText.slice(0, 500) : ""`)
	if err != nil {
		return nil
	}

	body := strings.ToLower(res.Value.String())
	switch {
	case strings.Contains(body, "unusual traffic"):
		return &ScrapeResult{
			Success: false,
			Message: "blocked by Google bot-check (\"unusual traffic\"); likely caused by a rotating proxy assigning a different exit IP per connection -- use a sticky/dedicated proxy IP instead",
			Error:   fmt.Errorf("bot check triggered"),
		}
	case strings.Contains(body, "not a bot") || strings.Contains(body, "confirm you're not a bot"):
		return &ScrapeResult{
			Success: false,
			Message: "blocked by YouTube bot-check (\"sign in to confirm you're not a bot\"); likely caused by proxy IP reputation -- use a residential/sticky proxy instead of a rotating datacenter proxy",
			Error:   fmt.Errorf("bot check triggered"),
		}
	}

	return nil
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
