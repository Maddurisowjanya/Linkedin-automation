package search

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/stealth"
)

// SearchProfiles performs a simple LinkedIn people search for the configured
// keywords, walking through a few pages of results and returning unique
// profile URLs.
func SearchProfiles(ctx context.Context, br *rod.Browser, cfg config.SearchConfig, log *logrus.Logger) ([]string, error) {
	page, err := br.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("open page: %w", err)
	}

	unique := make(map[string]struct{})
keywordLoop:
	for _, kw := range cfg.Keywords {
		// Check if context was canceled
		select {
		case <-ctx.Done():
			log.WithError(ctx.Err()).Warn("context canceled, stopping search")
			return make([]string, 0), nil
		default:
		}

		log.WithField("keyword", kw).Info("running LinkedIn search")

		searchURL := fmt.Sprintf("https://www.linkedin.com/search/results/people/?keywords=%s", url.QueryEscape(kw))

		// Add timeout for navigation
		navCtx, navCancel := context.WithTimeout(ctx, 45*time.Second)
		navDone := make(chan error, 1)
		go func() {
			if err := page.Navigate(searchURL); err != nil {
				navDone <- err
				return
			}
			if err := page.WaitLoad(); err != nil {
				navDone <- err
				return
			}
			// Wait a bit for content to render
			time.Sleep(2 * time.Second)
			navDone <- nil
		}()

		select {
		case err := <-navDone:
			navCancel()
			if err != nil {
				log.WithError(err).WithField("keyword", kw).Warn("failed to navigate/search, skipping keyword")
				continue
			}
		case <-navCtx.Done():
			navCancel()
			log.WithError(navCtx.Err()).WithField("keyword", kw).Warn("search navigation timeout, skipping keyword")
			continue
		}

		// Check if we're on a checkpoint/challenge page
		info, err := page.Info()
		if err == nil && info.URL != "" {
			if strings.Contains(info.URL, "checkpoint") || strings.Contains(info.URL, "challenge") {
				log.WithField("url", info.URL).WithField("keyword", kw).Warn("⚠️  CHECKPOINT DETECTED - Please complete verification in the browser window, then wait 30 seconds")
				log.Warn("Waiting 30 seconds for you to complete the checkpoint...")

				// Wait for user to complete checkpoint
				checkpointCtx, checkpointCancel := context.WithTimeout(ctx, 30*time.Second)
				<-checkpointCtx.Done()
				checkpointCancel()

				// Check again if still on checkpoint
				if info2, err2 := page.Info(); err2 == nil && info2.URL != "" {
					if strings.Contains(info2.URL, "checkpoint") || strings.Contains(info2.URL, "challenge") {
						log.WithField("url", info2.URL).Warn("Still on checkpoint page - skipping this search keyword")
						continue
					}
					log.Info("Checkpoint appears resolved, continuing with search")
				}
			}
		}

		for pageIdx := 0; pageIdx < cfg.MaxPages; pageIdx++ {
			// Check context cancellation
			select {
			case <-ctx.Done():
				log.WithError(ctx.Err()).Warn("context canceled during pagination")
				break keywordLoop
			default:
			}

			// Let content load and scroll a bit to trigger lazy loading.
			stealth.RandomDelay(cfg.PageDelayMin, cfg.PageDelayMax)

			// Scroll to trigger lazy loading of profile cards
			if err := stealth.ScrollHumanLike(page, 3*time.Second); err != nil {
				log.WithError(err).Warn("failed to scroll page")
			}

			// Wait a bit more for content to load after scrolling
			time.Sleep(1 * time.Second)

			urls, err := extractProfileURLs(page)
			if err != nil {
				log.WithError(err).Warn("failed to extract some profile URLs")
			}
			if len(urls) == 0 && pageIdx == 0 {
				log.WithField("keyword", kw).Warn("no profile URLs found on first page - page may not have loaded correctly")
			}
			for _, u := range urls {
				unique[u] = struct{}{}
			}
			log.WithField("keyword", kw).WithField("page", pageIdx+1).WithField("found", len(urls)).Debug("extracted profile URLs from page")

			// Attempt to go to the "next" page if available.
			nextBtn, err := page.ElementR("button, a", "Next")
			if err != nil || nextBtn == nil {
				break
			}
			if err := nextBtn.Click("left", 1); err != nil {
				log.WithError(err).Warn("failed to click next button")
				break
			}

			// Add timeout for page load
			loadCtx, loadCancel := context.WithTimeout(ctx, 15*time.Second)
			loadDone := make(chan error, 1)
			go func() {
				loadDone <- page.WaitLoad()
			}()

			select {
			case err := <-loadDone:
				loadCancel()
				if err != nil {
					log.WithError(err).Warn("failed to wait for next page load")
					break
				}
			case <-loadCtx.Done():
				loadCancel()
				log.WithError(loadCtx.Err()).Warn("page load timeout")
				break keywordLoop
			}
		}
	}

	out := make([]string, 0, len(unique))
	for u := range unique {
		out = append(out, u)
	}
	log.WithField("total_profiles", len(out)).Info("search completed")
	return out, nil
}

func extractProfileURLs(page *rod.Page) ([]string, error) {
	links, err := page.Elements("a")
	if err != nil {
		return nil, err
	}
	var profiles []string
	for _, a := range links {
		href, err := a.Attribute("href")
		if err != nil || href == nil {
			continue
		}
		u := *href
		if strings.Contains(u, "/in/") { // basic heuristic for profile URLs
			if idx := strings.Index(u, "?"); idx > 0 {
				u = u[:idx]
			}
			profiles = append(profiles, u)
		}
	}
	return profiles, nil
}
