package connect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// SendConnectionRequests iterates over profile URLs, enforcing a simple
// daily limit and recording sent requests in SQLite so duplicates are
// avoided.
func SendConnectionRequests(
	ctx context.Context,
	br *rod.Browser,
	store *storage.Storage,
	cfg config.ConnectConfig,
	profiles []string,
	log *logrus.Logger,
) error {
	page, err := br.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("open page: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	sentToday, err := store.CountRequestsSince(ctx, today)
	if err != nil {
		return err
	}

	for _, profileURL := range profiles {
		// Check if context was canceled (user closed browser, timeout, etc.)
		select {
		case <-ctx.Done():
			log.WithError(ctx.Err()).Warn("context canceled, stopping connection requests")
			return nil
		default:
		}

		if sentToday >= cfg.DailyLimit {
			log.WithField("limit", cfg.DailyLimit).Info("daily connect limit reached")
			return nil
		}

		already, err := store.HasSentRequest(ctx, profileURL)
		if err != nil {
			log.WithError(err).WithField("profile", profileURL).Warn("failed to check if request already sent, skipping")
			continue
		}
		if already {
			continue
		}

		log.WithField("profile", profileURL).Info("visiting profile to send connection request")
		
		// Add timeout for navigation
		navCtx, navCancel := context.WithTimeout(ctx, 30*time.Second)
		navDone := make(chan error, 1)
		go func() {
			if err := page.Navigate(profileURL); err != nil {
				navDone <- err
				return
			}
			navDone <- page.WaitLoad()
		}()
		
		select {
		case err := <-navDone:
			navCancel()
			if err != nil {
				log.WithError(err).WithField("profile", profileURL).Warn("failed to navigate to profile, skipping")
				continue
			}
		case <-navCtx.Done():
			navCancel()
			log.WithError(navCtx.Err()).WithField("profile", profileURL).Warn("navigation timeout, skipping")
			continue
		}

		// Attempt to locate a "Connect" button. LinkedIn may change its
		// markup; this is intentionally heuristic for a PoC.
		btn, err := page.ElementR("button", "Connect")
		if err != nil || btn == nil {
			log.WithField("profile", profileURL).Warn("no Connect button found")
			continue
		}

		if err := btn.Click("left", 1); err != nil {
			log.WithError(err).WithField("profile", profileURL).Warn("failed to click Connect")
			continue
		}

		// Some flows open a dialog with "Add a note".
		addNote, _ := page.ElementR("button", "Add a note")
		if addNote != nil {
			_ = addNote.Click("left", 1)
			if noteArea, _ := page.Element("textarea"); noteArea != nil {
				note := renderTemplate(cfg.NoteTemplate, map[string]string{
					"PROFILE_URL": profileURL,
				})
				if err := noteArea.Input(note); err != nil {
					log.WithError(err).Warn("failed to fill note textarea")
				}
			}
		}

		sendBtn, _ := page.ElementR("button", "Send")
		if sendBtn != nil {
			_ = sendBtn.Click("left", 1)
		}

		if err := store.RecordRequest(ctx, profileURL, time.Now()); err != nil {
			log.WithError(err).WithField("profile", profileURL).Warn("failed to record request in storage, continuing anyway")
		} else {
			sentToday++
			log.WithField("profile", profileURL).Info("connection request sent successfully")
		}

		// Think‑time between actions.
		stealth.RandomDelay(cfg.ActionDelayMin, cfg.ActionDelayMax)
	}
	
	log.WithField("total_sent", sentToday).Info("finished processing connection requests")
	return nil
}

// Simple variable replacement in templates like "Hi {{PROFILE_URL}}".
func renderTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}


