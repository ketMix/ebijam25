package world

import (
	"github.com/ketMix/ebijam25/internal/message/event"
)

// EventBus is the event bus used for events.
var EventBus = event.NewBus()
