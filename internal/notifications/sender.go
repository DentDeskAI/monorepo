// Package notifications — автоматические напоминания и follow-up сообщения.
package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/store"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

type Sender struct {
	DB       *sqlx.DB
	Log      zerolog.Logger
	WhatsApp *whatsapp.Client
	Repo     *store.AppointmentRepo
}

func (s *Sender) RunTick(ctx context.Context) {
	now := time.Now()
	s.send24h(ctx, now)
	s.send2h(ctx, now)
	s.sendFollowup(ctx, now)
}

func (s *Sender) send24h(ctx context.Context, now time.Time) {
	items, err := s.Repo.DueForReminder24h(ctx, now)
	if err != nil {
		s.Log.Error().Err(err).Msg("fetch 24h")
		return
	}
	for _, a := range items {
		if a.PatientPhone == nil {
			continue
		}
		body := format24h(a)
		if err := s.WhatsApp.SendText(ctx, *a.PatientPhone, body); err != nil {
			s.Log.Error().Err(err).Msg("send 24h")
			continue
		}
		if err := s.Repo.MarkReminder24hSent(ctx, a.ID); err != nil {
			s.Log.Error().Err(err).Msg("mark 24h")
		}
	}
}

func (s *Sender) send2h(ctx context.Context, now time.Time) {
	items, err := s.Repo.DueForReminder2h(ctx, now)
	if err != nil {
		s.Log.Error().Err(err).Msg("fetch 2h")
		return
	}
	for _, a := range items {
		if a.PatientPhone == nil {
			continue
		}
		body := format2h(a)
		if err := s.WhatsApp.SendText(ctx, *a.PatientPhone, body); err != nil {
			s.Log.Error().Err(err).Msg("send 2h")
			continue
		}
		_ = s.Repo.MarkReminder2hSent(ctx, a.ID)
	}
}

func (s *Sender) sendFollowup(ctx context.Context, now time.Time) {
	items, err := s.Repo.DueForFollowup(ctx, now)
	if err != nil {
		s.Log.Error().Err(err).Msg("fetch followup")
		return
	}
	for _, a := range items {
		if a.PatientPhone == nil {
			continue
		}
		body := "Спасибо, что посетили нас! Как самочувствие после приёма? Если будут вопросы — просто напишите в этот чат."
		if err := s.WhatsApp.SendText(ctx, *a.PatientPhone, body); err != nil {
			s.Log.Error().Err(err).Msg("send followup")
			continue
		}
		_ = s.Repo.MarkFollowupSent(ctx, a.ID)
	}
}

func format24h(a store.Appointment) string {
	doc := "врача"
	if a.DoctorName != nil {
		doc = "врача " + *a.DoctorName
	}
	return fmt.Sprintf(
		"Напоминаем о вашей записи завтра в %02d:%02d к %s. Если нужно перенести — просто напишите.",
		a.StartsAt.Hour(), a.StartsAt.Minute(), doc,
	)
}

func format2h(a store.Appointment) string {
	return fmt.Sprintf(
		"Ждём вас сегодня в %02d:%02d 🙂 Если планы изменились — дайте знать.",
		a.StartsAt.Hour(), a.StartsAt.Minute(),
	)
}
