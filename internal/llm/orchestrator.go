package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/scheduling"
	"github.com/dentdesk/dentdesk/internal/store"
)

// Intent — что пациент хочет.
type Intent string

const (
	IntentBooking    Intent = "booking"
	IntentReschedule Intent = "reschedule"
	IntentCancel     Intent = "cancel"
	IntentUrgentPain Intent = "urgent_pain"
	IntentPricing    Intent = "pricing"
	IntentGreeting   Intent = "greeting"
	IntentOther      Intent = "other"
)

// Extracted — структурированный результат классификации сообщения.
type Extracted struct {
	Intent   Intent  `json:"intent"`
	Service  *string `json:"service,omitempty"`
	Doctor   *string `json:"doctor,omitempty"`
	When     *string `json:"when,omitempty"` // free-form: "завтра утром", "после обеда"
	Language string  `json:"language"`       // "ru" | "kk"
}

// Orchestrator связывает LLM с расписанием и диалогом.
type Orchestrator struct {
	llm   Client
	sched scheduling.Scheduler
	log   zerolog.Logger
}

func NewOrchestrator(llm Client, sched scheduling.Scheduler, log zerolog.Logger) *Orchestrator {
	return &Orchestrator{llm: llm, sched: sched, log: log}
}

// ConvState — состояние диалога, хранится в conversations.context.
type ConvState struct {
	LastIntent     Intent              `json:"last_intent,omitempty"`
	PendingSlots   []scheduling.Slot    `json:"pending_slots,omitempty"`
	PendingService string              `json:"pending_service,omitempty"`
}

// Reply — что вернуть пациенту + возможное действие для нашего кода.
type Reply struct {
	Text         string
	NewState     ConvState
	ActionTaken  string // "booked" | "slots_offered" | "handoff" | ""
	Appointment  *scheduling.BookResult
	ChosenSlot   *scheduling.Slot
	Meta         map[string]any
}

// Handle — главная точка входа.
// history — последние ~10 сообщений в хронологическом порядке.
// state — текущее состояние диалога (может быть нулевым).
func (o *Orchestrator) Handle(
	ctx context.Context,
	clinicID, patientID uuid.UUID,
	incoming string,
	history []store.Message,
	state ConvState,
) (*Reply, error) {
	// 1) Если бот только что предложил слоты — сначала пытаемся распознать выбор.
	if len(state.PendingSlots) > 0 {
		if chosen := matchSlotChoice(incoming, state.PendingSlots); chosen != nil {
			book, err := o.sched.CreateAppointment(ctx, scheduling.BookRequest{
				ClinicID:  clinicID,
				PatientID: patientID,
				DoctorID:  chosen.DoctorID,
				StartsAt:  chosen.StartsAt,
				EndsAt:    chosen.EndsAt,
				Service:   state.PendingService,
			})
			if err != nil {
				o.log.Error().Err(err).Msg("book failed")
				return &Reply{
					Text:     "Ой, не получилось закрепить время. Попробуем ещё раз — какое время удобно?",
					NewState: ConvState{LastIntent: IntentBooking},
				}, nil
			}
			text := fmt.Sprintf(
				"Записала вас на %s к врачу %s 🙂 Если планы поменяются — напишите, перенесём.",
				formatSlotHuman(*chosen), chosen.Doctor,
			)
			return &Reply{
				Text:        text,
				NewState:    ConvState{},
				ActionTaken: "booked",
				Appointment: book,
				ChosenSlot:  chosen,
			}, nil
		}
		// Не распознали выбор — падаем в обычный флоу, но слоты оставляем.
	}

	// 2) Классификация сообщения.
	extracted, err := o.classify(ctx, incoming)
	if err != nil {
		o.log.Warn().Err(err).Msg("classify failed, falling back")
		extracted = &Extracted{Intent: IntentOther, Language: "ru"}
	}

	// 3) Если это запись или срочная боль — ищем слоты.
	newState := state
	newState.LastIntent = extracted.Intent
	var slotHint string

	if extracted.Intent == IntentBooking || extracted.Intent == IntentUrgentPain {
		from := time.Now().Add(30 * time.Minute)
		to := from.Add(72 * time.Hour)
		var specialty string
		if extracted.Doctor != nil {
			specialty = guessSpecialty(*extracted.Doctor)
		}
		slots, err := o.sched.GetFreeSlots(ctx, clinicID, from, to, specialty)
		if err != nil {
			o.log.Error().Err(err).Msg("get slots failed")
		}
		if len(slots) > 0 {
			// берём первые 3 разнесённые по времени
			picked := pickDiverseSlots(slots, 3)
			newState.PendingSlots = picked
			if extracted.Service != nil {
				newState.PendingService = *extracted.Service
			}
			slotHint = "[СВОБОДНЫЕ СЛОТЫ: " + slotsToString(picked) + "]"
		} else {
			slotHint = "[НЕТ СЛОТОВ]"
		}
	}

	// 4) Генерируем ответ через LLM с persona + контекст.
	text, err := o.generateReply(ctx, incoming, history, slotHint)
	if err != nil {
		o.log.Error().Err(err).Msg("generate failed")
		text = "Минутку, сейчас помогу. Подскажите, хотите записаться на приём?"
	}
	safe, intervened := ApplyGuardrails(text)
	if intervened {
		o.log.Warn().Msg("guardrails intervened")
	}

	action := ""
	if len(newState.PendingSlots) > 0 {
		action = "slots_offered"
	}

	return &Reply{
		Text:        safe,
		NewState:    newState,
		ActionTaken: action,
		Meta: map[string]any{
			"intent":   extracted.Intent,
			"language": extracted.Language,
		},
	}, nil
}

// ---- вспомогательные ----

func (o *Orchestrator) classify(ctx context.Context, incoming string) (*Extracted, error) {
	system := `Ты — классификатор намерений пациента стоматологии.
Верни СТРОГО JSON: {"intent": "...", "service": null|"...", "doctor": null|"...", "when": null|"...", "language": "ru"|"kk"}.
Возможные intent: booking, reschedule, cancel, urgent_pain, pricing, greeting, other.
Не добавляй текст вокруг JSON, никаких пояснений.`
	resp, err := o.llm.Chat(ctx, ChatRequest{
		System: system,
		Messages: []Message{
			{Role: RoleUser, Content: incoming},
		},
		Temperature: 0,
		MaxTokens:   200,
		JSONOnly:    true,
	})
	if err != nil {
		return nil, err
	}
	raw := extractJSON(resp.Text)
	var out Extracted
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("classify parse: %w; raw=%q", err, resp.Text)
	}
	if out.Language == "" {
		out.Language = "ru"
	}
	return &out, nil
}

func (o *Orchestrator) generateReply(
	ctx context.Context,
	incoming string,
	history []store.Message,
	slotHint string,
) (string, error) {
	msgs := make([]Message, 0, len(history)+2)
	// последние 8 сообщений диалога
	limit := len(history)
	if limit > 8 {
		history = history[limit-8:]
	}
	for _, h := range history {
		role := RoleUser
		if h.Direction == "outbound" {
			role = RoleAssistant
		}
		msgs = append(msgs, Message{Role: role, Content: h.Body})
	}
	// актуальное сообщение + служебный хинт
	content := incoming
	if slotHint != "" {
		content = content + "\n\n" + slotHint
	}
	msgs = append(msgs, Message{Role: RoleUser, Content: content})

	resp, err := o.llm.Chat(ctx, ChatRequest{
		System:      SystemPromptAigul,
		Messages:    msgs,
		Temperature: 0.6,
		MaxTokens:   300,
	})
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

// matchSlotChoice — пытается понять, какой слот выбрал пациент в свободной форме.
func matchSlotChoice(msg string, slots []scheduling.Slot) *scheduling.Slot {
	m := strings.ToLower(strings.TrimSpace(msg))
	if m == "" {
		return nil
	}
	// 1) по номеру: "1", "первый", "1."
	if num := detectNumber(m); num >= 1 && num <= len(slots) {
		s := slots[num-1]
		return &s
	}
	// 2) по времени "10:00", "17"
	hh, mm, ok := detectTime(m)
	if ok {
		for _, s := range slots {
			if s.StartsAt.Hour() == hh && (mm < 0 || s.StartsAt.Minute() == mm) {
				cp := s
				return &cp
			}
		}
	}
	// 3) прямое "да" при единственном слоте
	if len(slots) == 1 && (strings.Contains(m, "да") || strings.Contains(m, "ок") || strings.Contains(m, "хорошо") || strings.Contains(m, "ия")) {
		s := slots[0]
		return &s
	}
	return nil
}

func detectNumber(m string) int {
	keywords := map[string]int{
		"первый": 1, "первое": 1, "1": 1, "бірінші": 1,
		"второй": 2, "второе": 2, "2": 2, "екінші": 2,
		"третий": 3, "третье": 3, "3": 3, "үшінші": 3,
	}
	for k, v := range keywords {
		if strings.Contains(m, k) {
			return v
		}
	}
	return 0
}

// detectTime возвращает (час, минута, ok). Минута = -1 если не указана.
func detectTime(m string) (int, int, bool) {
	// ищем HH:MM
	for i := 0; i+4 <= len(m); i++ {
		if i+5 <= len(m) && m[i+2] == ':' && isDigit(m[i]) && isDigit(m[i+1]) && isDigit(m[i+3]) && isDigit(m[i+4]) {
			h := int(m[i]-'0')*10 + int(m[i+1]-'0')
			mi := int(m[i+3]-'0')*10 + int(m[i+4]-'0')
			if h >= 0 && h < 24 && mi >= 0 && mi < 60 {
				return h, mi, true
			}
		}
	}
	// ищем одиночное двузначное число 8..21
	for i := 0; i+2 <= len(m); i++ {
		if isDigit(m[i]) && isDigit(m[i+1]) && (i+2 == len(m) || !isDigit(m[i+2])) {
			if i > 0 && isDigit(m[i-1]) {
				continue
			}
			h := int(m[i]-'0')*10 + int(m[i+1]-'0')
			if h >= 8 && h <= 21 {
				return h, -1, true
			}
		}
	}
	return 0, 0, false
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func pickDiverseSlots(slots []scheduling.Slot, n int) []scheduling.Slot {
	if len(slots) <= n {
		return slots
	}
	// Выбираем равномерно: первый, середина, последний из топа.
	out := []scheduling.Slot{slots[0]}
	if n >= 2 {
		out = append(out, slots[len(slots)/2])
	}
	if n >= 3 {
		out = append(out, slots[len(slots)-1])
	}
	return out
}

func slotsToString(slots []scheduling.Slot) string {
	parts := make([]string, 0, len(slots))
	for i, s := range slots {
		parts = append(parts, fmt.Sprintf("%d) %s к %s", i+1, formatSlotHuman(s), s.Doctor))
	}
	return strings.Join(parts, "; ")
}

func formatSlotHuman(s scheduling.Slot) string {
	now := time.Now()
	day := ""
	d := s.StartsAt
	switch {
	case sameDay(d, now):
		day = "сегодня"
	case sameDay(d, now.Add(24*time.Hour)):
		day = "завтра"
	case sameDay(d, now.Add(48*time.Hour)):
		day = "послезавтра"
	default:
		day = d.Format("02.01")
	}
	return fmt.Sprintf("%s в %02d:%02d", day, d.Hour(), d.Minute())
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func guessSpecialty(doctor string) string {
	l := strings.ToLower(doctor)
	switch {
	case strings.Contains(l, "хирург") || strings.Contains(l, "удалит"):
		return "surgeon"
	case strings.Contains(l, "ортодонт") || strings.Contains(l, "брекет"):
		return "orthodontist"
	default:
		return ""
	}
}

// extractJSON — достаёт первый {...} блок из текста (LLM иногда оборачивает в ```).
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
