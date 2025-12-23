package messaging

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

// SendFollowUps is a high‑level demo that navigates to the "My Network"
// area, identifies recently accepted connections (heuristically) and sends
// them a follow‑up message using simple templates.
func SendFollowUps(
	ctx context.Context,
	br *rod.Browser,
	store *storage.Storage,
	cfg config.MessagingConfig,
	log *logrus.Logger,
) error {
	if len(cfg.Templates) == 0 {
		log.Warn("no messaging templates configured – skipping follow‑ups")
		return nil
	}

	page, err := br.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return err
	}

	if err := page.Navigate("https://www.linkedin.com/mynetwork/"); err != nil {
		return err
	}
	if err := page.WaitLoad(); err != nil {
		return err
	}

	// This PoC does not implement a full "newly accepted only" detector.
	// Instead it demonstrates how one might iterate over a small number of
	// connection cards and open the message UI.
	connections, _ := page.Elements("a")

	today := time.Now().Truncate(24 * time.Hour)
	sentToday, err := store.CountMessagesSince(ctx, "followup", today)
	if err != nil {
		return err
	}

	for _, c := range connections {
		// Check context cancellation
		select {
		case <-ctx.Done():
			log.WithError(ctx.Err()).Warn("context canceled, stopping messaging")
			return nil
		default:
		}

		if sentToday >= cfg.DailyLimit {
			log.WithField("limit", cfg.DailyLimit).Info("daily messaging limit reached")
			return nil
		}

		href, _ := c.Attribute("href")
		if href == nil || !strings.Contains(*href, "/in/") {
			continue
		}
		profileURL := *href

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

		msgBtn, _ := page.ElementR("button", "Message")
		if msgBtn == nil {
			continue
		}
		if err := msgBtn.Click("left", 1); err != nil {
			continue
		}

		// Locate the message textarea / editor – highly simplified.
		editor, _ := page.Element("div[role=textbox], textarea")
		if editor == nil {
			continue
		}

		tpl := cfg.Templates[rand.Intn(len(cfg.Templates))]
		body := renderTemplate(tpl, map[string]string{
			"PROFILE_URL": profileURL,
		})
		if err := editor.Input(body); err != nil {
			continue
		}

		sendBtn, _ := page.ElementR("button", "Send")
		if sendBtn != nil {
			_ = sendBtn.Click("left", 1)
		}

		if err := store.RecordMessage(ctx, profileURL, "followup", time.Now()); err != nil {
			log.WithError(err).WithField("profile", profileURL).Warn("failed to record message in storage, continuing anyway")
		} else {
			sentToday++
			log.WithField("profile", profileURL).Info("follow-up message sent successfully")
		}

		stealth.RandomDelay(cfg.ActionDelayMin, cfg.ActionDelayMax)
	}
	
	log.WithField("total_sent", sentToday).Info("finished processing follow-up messages")
	return nil
}

// shared simple templating helper
func renderTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}



