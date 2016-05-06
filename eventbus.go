package ycq

type EventBus interface {
	PublishEvent(EventMessage)
	AddHandler(EventHandler, ...interface{})
	AddLocalHandler(EventHandler)
	AddGlobalHandler(EventHandler)
}

type InternalEventBus struct {
	eventHandlers  map[string]map[EventHandler]struct{}
	localHandlers  map[EventHandler]struct{}
	globalHandlers map[EventHandler]struct{}
}

func NewInternalEventBus() *InternalEventBus {
	b := &InternalEventBus{
		eventHandlers:  make(map[string]map[EventHandler]struct{}),
		localHandlers:  make(map[EventHandler]struct{}),
		globalHandlers: make(map[EventHandler]struct{}),
	}
	return b
}

func (b *InternalEventBus) PublishEvent(event EventMessage) {
	if handlers, ok := b.eventHandlers[event.EventType()]; ok {
		for handler := range handlers {
			handler.Handle(event)
		}
	}

	for handler := range b.localHandlers {
		handler.Handle(event)
	}
	for handler := range b.globalHandlers {
		handler.Handle(event)
	}
}
func (b *InternalEventBus) AddHandler(handler EventHandler, events ...interface{}) {

	for _, event := range events {
		typeName := typeOf(event)
		if _, ok := b.eventHandlers[typeName]; !ok {
			b.eventHandlers[typeName] = make(map[EventHandler]struct{})
		}

		b.eventHandlers[typeName][handler] = struct{}{}
	}
}

func (b *InternalEventBus) AddLocalHandler(handler EventHandler) {
	b.localHandlers[handler] = struct{}{}
}

func (b *InternalEventBus) AddGlobalHandler(handler EventHandler) {
	b.globalHandlers[handler] = struct{}{}
}
