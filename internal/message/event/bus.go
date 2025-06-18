package event

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/ketMix/ebijam25/internal/log"
)

type Event interface {
	Type() string
}

type Bus struct {
	log         *slog.Logger
	debugName   string
	events      []Event
	handlers    map[string][]func(Event)
	eventToPipe map[string][]*Bus
}

func (b *Bus) Publish(event Event) {
	b.log.Debug("publish", "event", event.Type())
	b.events = append(b.events, event)
}

func (b *Bus) Subscribe(eventType string, handler func(Event)) {
	if b.handlers == nil {
		b.handlers = make(map[string][]func(Event))
	}
	b.log.Debug("subscribe", "event", eventType)
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) ProcessEvents() {
	for _, event := range b.events {
		if handlers, ok := b.handlers[event.Type()]; ok {
			for _, handler := range handlers {
				b.log.Debug("handle", "event", event.Type())
				handler(event)
			}
		}

		// Also pipe the event to other buses
		for key, buses := range b.eventToPipe {
			if strings.HasPrefix(event.Type(), key) {
				for _, otherBus := range buses {
					b.log.Debug("pipe", "event", event.Type(), "to", otherBus.debugName)
					otherBus.Publish(event)
				}
			}
		}
	}
	b.events = nil // Clear events after processing
}

func (b *Bus) Pipe(other *Bus, events []string) {
	for _, eventType := range events {
		if b.eventToPipe == nil {
			b.eventToPipe = make(map[string][]*Bus)
		}
		if slices.Contains(b.eventToPipe[eventType], other) {
			continue // Already piped to this bus for this event type
		}
		b.eventToPipe[eventType] = append(b.eventToPipe[eventType], other)
		b.log.Debug("pipe", "event", eventType, "to", other.debugName)
	}
}

func (b *Bus) Unpipe(other *Bus) {
	for eventType, buses := range b.eventToPipe {
		if idx := slices.Index(buses, other); idx != -1 {
			b.eventToPipe[eventType] = append(buses[:idx], buses[idx+1:]...)
		}
		if len(b.eventToPipe[eventType]) == 0 {
			delete(b.eventToPipe, eventType) // Remove the event type if no buses are left
		}
	}
}

func NewBus(name string) *Bus {
	return &Bus{
		log:       log.New("bus", name),
		debugName: name,
		events:    []Event{},
		handlers:  make(map[string][]func(Event)),
	}
}
