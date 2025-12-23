package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linkedin-automation-poc/internal/auth"
	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/connect"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/messaging"
	"linkedin-automation-poc/internal/search"
	"linkedin-automation-poc/internal/storage"
)

// main wires together config, logging, browser, storage and a small demo flow.
// This is an educational proof-of-concept only – DO NOT use it for production
// scraping or to violate LinkedIn's Terms of Service.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := logger.New()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	br, err := browser.New(ctx, cfg.Browser, log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialise browser")
	}
	defer br.MustClose()

	db, err := storage.New(cfg.Database.DSN, log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialise storage")
	}
	defer db.Close()

	email := os.Getenv("LINKEDIN_EMAIL")
	password := os.Getenv("LINKEDIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set in the environment")
	}

	if err := auth.Login(ctx, br, db, email, password, log); err != nil {
		log.WithError(err).Fatal("login failed")
	}

	// Simple demo: run a single search and attempt a few connection requests.
	profiles, err := search.SearchProfiles(ctx, br, cfg.Search, log)
	if err != nil {
		log.WithError(err).Error("search failed, but continuing with any profiles found")
		profiles = []string{} // Continue with empty list instead of crashing
	}

	if len(profiles) > 0 {
		if err := connect.SendConnectionRequests(ctx, br, db, cfg.Connect, profiles, log); err != nil {
			log.WithError(err).Error("connection workflow encountered errors, but continuing")
		}
	} else {
		log.Warn("no profiles found to connect with")
	}

	// Demo: send follow‑up messages to newly accepted connections.
	if err := messaging.SendFollowUps(ctx, br, db, cfg.Messaging, log); err != nil {
		log.WithError(err).Error("follow‑up messaging encountered errors, but continuing")
	}

	// Give async logging/browsing a moment to settle in this simple demo.
	time.Sleep(2 * time.Second)
}
