package ycq

//EventBus is the inteface that an event bus must implement.
type EventBus interface {
	PublishEvent(EventMessage)
	AddHandler(EventHandler, ...interface{})
	AddLocalHandler(EventHandler)
	AddGlobalHandler(EventHandler)
}

//InternalEventBus provides a lightweight in process event bus
type InternalEventBus struct {
	eventHandlers  map[string]map[EventHandler]struct{}
	localHandlers  map[EventHandler]struct{}
	globalHandlers map[EventHandler]struct{}
}

//NewInternalEventBus constructs a new InternalEventBus
func NewInternalEventBus() *InternalEventBus {
	b := &InternalEventBus{
		eventHandlers:  make(map[string]map[EventHandler]struct{}),
		localHandlers:  make(map[EventHandler]struct{}),
		globalHandlers: make(map[EventHandler]struct{}),
	}
	return b
}

//PublishEvent publishes events to all registered event handlers
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

//AddHandler registers an event handler for all of the events specified in the
//variadic events parameter.
func (b *InternalEventBus) AddHandler(handler EventHandler, events ...interface{}) {

	for _, event := range events {
		typeName := typeOf(event)
		if _, ok := b.eventHandlers[typeName]; !ok {
			b.eventHandlers[typeName] = make(map[EventHandler]struct{})
		}

		b.eventHandlers[typeName][handler] = struct{}{}
	}
}

//TODO
func (b *InternalEventBus) AddLocalHandler(handler EventHandler) {
	b.localHandlers[handler] = struct{}{}
}

//TODO
func (b *InternalEventBus) AddGlobalHandler(handler EventHandler) {
	b.globalHandlers[handler] = struct{}{}
}
