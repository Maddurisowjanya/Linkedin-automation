package browser

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"

	"linkedin-automation-poc/internal/config"
)

// New creates a new Rod browser instance configured with stealth techniques
// that try to mimic a real user. This is still only a best‑effort mitigation
// and must NOT be used for abusive automation.
func New(ctx context.Context, cfg config.BrowserConfig, log *logrus.Logger) (*rod.Browser, error) {
	rand.Seed(time.Now().UnixNano())

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = randomUserAgent()
	}

	l := launcher.New().Leakless(false).
		// Configurable headless mode – running with a visible window generally
		// looks more like a real user.
		Headless(cfg.Headless).
		// Randomised user‑agent to avoid always advertising the same browser
		// fingerprint.
		Append("user-agent", userAgent).
		// Custom viewport (window size) to avoid a fixed, easily detectable
		// automation resolution.
		Append("window-size", fmt.Sprintf("%dx%d", cfg.ViewportWidth, cfg.ViewportHeight)).
		// Disable some obvious automation blink features.
		Append("disable-blink-features", "AutomationControlled")

	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	br := rod.New().ControlURL(url).MustConnect()

	// 1) Use the official rod/stealth helper, which creates a new page and
	// patches a wide range of automation flags (languages, plugins,
	// webdriver flag, etc).
	page := stealth.MustPage(br)

	// 2) Explicitly ensure navigator.webdriver is undefined. Some sites check
	// this property directly; returning undefined mirrors a typical real user.
	_, err = page.Eval(`
		() => {
			Object.defineProperty(navigator, 'webdriver', {
				get: () => undefined,
			});
		}
	`, nil)
	if err != nil {
		log.WithError(err).Warn("failed to override navigator.webdriver")
	}

	// 3) Inject small JS snippet to mask other simple automation hints,
	// such as overly small plugin/language lists, which some detection
	// scripts look for.
	_, err = page.Eval(`
		() => {
			try {
				// Pretend we have some common plugins installed.
				Object.defineProperty(navigator, 'plugins', {
					get: () => [1, 2, 3],
				});

				// Provide a realistic language list.
				Object.defineProperty(navigator, 'languages', {
					get: () => ['en-US', 'en'],
				});
			} catch (e) {
				// Best effort – failures here are non‑fatal.
			}
		}
	`, nil)
	if err != nil {
		log.WithError(err).Warn("failed to inject additional stealth JS")
	}

	return br, nil
}

// randomUserAgent returns a pseudo‑random modern desktop Chrome user‑agent.
// Rotating user‑agents slightly changes the fingerprint presented to servers.
func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
	return agents[rand.Intn(len(agents))]
}



