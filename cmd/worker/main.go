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

	sender := &notifications.Sender{
		DB:       database,
		Log:      log,
		WhatsApp: whatsapp.NewClient(cfg.WhatsAppToken, cfg.WhatsAppPhoneNumberID, cfg.WhatsAppAPIVersion),
		Repo:     appointments.NewRepo(database),
	}

	// Единый тикер — реминдеры проверяются каждые 5 минут, этого достаточно
	// для окон 24ч±30м и 2ч±30м. Прод-вариант — разные тикеры.
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()

	// сразу первый прогон
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
