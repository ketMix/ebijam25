package main

type Event interface {
	Type() string
}

type EventProduce struct {
	structure  *Structure
	individual *Individual
}

func (e *EventProduce) Type() string {
	return "produce"
}

type EventMerge struct {
	from int
	to   int
}

func (e *EventMerge) Type() string {
	return "merge"
}

type EventResourceDepleted struct {
	id int
}

func (e *EventResourceDepleted) Type() string {
	return "deplete"
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
