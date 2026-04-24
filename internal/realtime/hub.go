// Package realtime — простой SSE hub для push'а новых сообщений в CRM операторам.
package realtime

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

type Event struct {
	Type    string          `json:"type"` // "message" | "appointment"
	Payload json.RawMessage `json:"payload"`
}

type subscriber struct {
	clinicID uuid.UUID
	ch       chan Event
}

type Hub struct {
	mu   sync.RWMutex
	subs map[*subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: map[*subscriber]struct{}{}}
}

func (h *Hub) Subscribe(clinicID uuid.UUID) (ch <-chan Event, unsubscribe func()) {
	s := &subscriber{clinicID: clinicID, ch: make(chan Event, 16)}
	h.mu.Lock()
	h.subs[s] = struct{}{}
	h.mu.Unlock()
	return s.ch, func() {
		h.mu.Lock()
		delete(h.subs, s)
		h.mu.Unlock()
		close(s.ch)
	}
}

func (h *Hub) Publish(clinicID uuid.UUID, eventType string, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	ev := Event{Type: eventType, Payload: b}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for s := range h.subs {
		if s.clinicID != clinicID {
			continue
		}
		select {
		case s.ch <- ev:
		default:
			// переполнение — дропаем, оператор просто перечитает ленту
		}
	}
}
