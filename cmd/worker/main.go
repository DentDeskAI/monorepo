package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/notifications"
	"github.com/dentdesk/dentdesk/internal/platform/config"
	"github.com/dentdesk/dentdesk/internal/platform/db"
	"github.com/dentdesk/dentdesk/internal/platform/logger"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)
	log.Info().Msg("worker starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("db connect")
	}
	defer database.Close()

	if migDir := firstExisting("./migrations", "/app/migrations"); migDir != "" {
		if err := db.RunMigrations(ctx, database, migDir); err != nil {
			log.Fatal().Err(err).Msg("migrations")
		}
		log.Info().Str("dir", migDir).Msg("migrations applied")
	} else {
		log.Warn().Msg("migrations dir not found, skipping")
	}

	sender := &notifications.Sender{
		DB:       database,
		Log:      log,
		WhatsApp: whatsapp.NewClient(cfg.WhatsAppToken, cfg.WhatsAppPhoneNumberID, cfg.WhatsAppAPIVersion),
		Repo:     appointments.NewRepo(database),
	}

	// One ticker is enough here: the worker checks reminders every 5 minutes and
	// uses wide enough windows for the 24h and 1h reminder jobs.
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()

	sender.RunTick(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tick.C:
			sender.RunTick(ctx)
		case <-stop:
			log.Info().Msg("worker stopping")
			return
		}
	}
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}
