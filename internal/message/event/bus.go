package event

type Event interface {
	Type() string
}

type Bus struct {
	events   []Event
	handlers map[string][]func(Event)
}

func (b *Bus) Publish(event Event) {
	b.events = append(b.events, event)
}

func (b *Bus) Subscribe(eventType string, handler func(Event)) {
	if b.handlers == nil {
		b.handlers = make(map[string][]func(Event))
	}
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) ProcessEvents() {
	for _, event := range b.events {
		if handlers, ok := b.handlers[event.Type()]; ok {
			for _, handler := range handlers {
				handler(event)
			}
		}
	}
	b.events = nil // Clear events after processing
}

func NewBus() *Bus {
	return &Bus{
		events:   []Event{},
		handlers: make(map[string][]func(Event)),
	}
}

var eventBus = NewBus()
