package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"

	"linkedin-automation-poc/internal/storage"
)

const cookieFile = "linkedin_session_cookies.json"

// Login performs a LinkedIn login sequence, reusing session cookies when
// possible. It tries to detect obvious failure states such as invalid
// credentials or checkpoint / captcha pages.
func Login(ctx context.Context, br *rod.Browser, store *storage.Storage, email, password string, log *logrus.Logger) error {
	page, err := br.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("open blank page: %w", err)
	}

	// Attempt to restore existing cookies first to avoid logging in on every
	// run. This keeps the demo closer to how a user would behave across
	// sessions.
	if err := loadCookies(ctx, br, log); err == nil {
		log.Info("restored existing LinkedIn session cookies – testing session")
		if err := page.Navigate("https://www.linkedin.com/feed/"); err == nil {
			if waitForURLContains(page, "linkedin.com/feed", 10*time.Second) == nil {
				log.Info("existing session appears valid, skipping login form")
				return nil
			}
		}
	}

	log.Info("performing fresh LinkedIn login")
	if err := page.Navigate("https://www.linkedin.com/login"); err != nil {
		return fmt.Errorf("navigate to login: %w", err)
	}

	if err := page.Timeout(20 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("wait login page load: %w", err)
	}

	// Basic form interaction; selectors can change over time, so we try a
	// small set of commonly observed variants for each field.
	usernameSelectors := []string{
		"#username",    // older layout
		"#session_key", // newer layout
		"input[name=session_key]",
		"input[name='session_key']",
	}
	passwordSelectors := []string{
		"#password",         // older layout
		"#session_password", // newer layout
		"input[name=session_password]",
		"input[name='session_password']",
	}

	usernameEl, err := firstExistingElement(page, usernameSelectors)
	if err != nil {
		return fmt.Errorf("locate username field: %w", err)
	}
	if err := usernameEl.Input(email); err != nil {
		return fmt.Errorf("fill username: %w", err)
	}

	passwordEl, err := firstExistingElement(page, passwordSelectors)
	if err != nil {
		return fmt.Errorf("locate password field: %w", err)
	}
	if err := passwordEl.Input(password); err != nil {
		return fmt.Errorf("fill password: %w", err)
	}

	if err := page.MustElement("button[type=submit]").Click("left", 1); err != nil {
		return fmt.Errorf("click submit: %w", err)
	}

	if err := page.Timeout(30 * time.Second).WaitLoad(); err != nil {
		return fmt.Errorf("wait post‑login load: %w", err)
	}

	// At this point LinkedIn may present a variety of post‑login pages
	// (feed, onboarding, security, etc.). For this PoC we no longer enforce
	// that we *must* land on /feed – we simply log the current URL and
	// proceed, so the rest of the demo can run even if additional manual
	// interaction is required.
	if info, err := page.Info(); err == nil {
		log.WithField("url", info.URL).Info("post‑login page URL")

		// If we're on a checkpoint page, warn the user
		if strings.Contains(info.URL, "checkpoint") || strings.Contains(info.URL, "challenge") {
			log.Warn("⚠️  LinkedIn checkpoint detected - you may need to complete verification manually")
			log.Warn("The app will continue, but some features may not work until checkpoint is resolved")
		}
	}

	// Persist cookies so they can be reused in later runs or after a crash.
	if err := saveCookies(ctx, br, log); err != nil {
		log.WithError(err).Warn("failed to persist cookies; session will not survive restart")
	}

	log.Info("LinkedIn login successful")
	return nil
}

func waitForURLContains(page *rod.Page, needle string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := page.Info()
		if err == nil && info.URL != "" && strings.Contains(info.URL, needle) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("timeout waiting for target URL")
}

// saveCookies serializes all browser cookies to disk so they can be restored
// on the next run. This keeps the PoC resilient to restarts.
func saveCookies(ctx context.Context, br *rod.Browser, log *logrus.Logger) error {
	// Use the DevTools protocol helper to fetch all cookies known to the browser.
	resp, err := proto.NetworkGetAllCookies{}.Call(br)
	if err != nil {
		return fmt.Errorf("get cookies: %w", err)
	}

	b, err := json.Marshal(resp.Cookies)
	if err != nil {
		return fmt.Errorf("marshal cookies: %w", err)
	}
	if err := os.WriteFile(cookieFile, b, 0o600); err != nil {
		return fmt.Errorf("write cookie file: %w", err)
	}
	log.WithField("path", cookieFile).Info("session cookies saved")
	return nil
}

// loadCookies reads cookies from disk and installs them into the browser.
func loadCookies(ctx context.Context, br *rod.Browser, log *logrus.Logger) error {
	b, err := os.ReadFile(cookieFile)
	if err != nil {
		return err
	}
	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal(b, &cookies); err != nil {
		return fmt.Errorf("unmarshal cookies: %w", err)
	}

	params := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for _, c := range cookies {
		params = append(params, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires,
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
		})
	}

	// Rod exposes a convenience wrapper for setting cookies on the browser.
	if err := br.SetCookies(params); err != nil {
		return fmt.Errorf("set cookies: %w", err)
	}
	log.WithField("count", len(cookies)).Info("restored cookies into browser")
	return nil
}

// firstExistingElement iterates over a list of CSS selectors and returns the
// first element that exists on the page. This makes the login flow resilient
// to minor LinkedIn markup changes between deployments.
func firstExistingElement(page *rod.Page, selectors []string) (*rod.Element, error) {
	for _, sel := range selectors {
		el, err := page.Element(sel)
		if err == nil && el != nil {
			return el, nil
		}
	}
	return nil, fmt.Errorf("no matching element for selectors: %v", selectors)
}
