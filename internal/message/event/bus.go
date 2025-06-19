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
	nextEvents  []Event
	processing  bool
	handlers    map[string][]func(Event)
	eventToPipe map[string][]*Bus
	NoQueue     bool // If true, events are processed immediately without queuing
}

func (b *Bus) Publish(event Event) {
	b.log.Debug("publish", "event", event.Type())
	if b.NoQueue {
		b.ProcessEvent(event)
		return
	}

	if !b.processing {
		b.events = append(b.events, event)
	} else {
		b.nextEvents = append(b.nextEvents, event)
	}
}

func (b *Bus) Subscribe(eventType string, handler func(Event)) {
	if b.handlers == nil {
		b.handlers = make(map[string][]func(Event))
	}
	b.log.Debug("subscribe", "event", eventType)
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) ProcessEvent(event Event) {
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

func (b *Bus) ProcessEvents() {
	if b.NoQueue {
		return
	}
	b.processing = true
	for _, event := range b.events {
		b.ProcessEvent(event)
	}
	b.processing = false
	b.events = b.nextEvents
	b.nextEvents = nil
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
