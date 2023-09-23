package pubsub

import "sync"

type (
	Broker struct {
		listeners map[string][]*listener

		lock sync.RWMutex
	}

	ListenerFunc func(event string, data any)

	listener struct {
		fn ListenerFunc
	}
)

var (
	// DefaultBroker is the default pubsub broker.
	DefaultBroker = NewBroker()
)

// NewBroker creates a new pubsub broker.
func NewBroker() *Broker {
	return &Broker{
		listeners: make(map[string][]*listener),
	}
}

// Subscribe adds a listener for the given event. The returned
// function can be used to unsubscribe the listener.
func (b *Broker) Subscribe(event string, cb ListenerFunc) (unsubscribe func()) {
	newListener := &listener{fn: cb}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.listeners[event] = append(b.listeners[event], newListener)

	return func() {
		b.lock.Lock()
		defer b.lock.Unlock()

		oldL := b.listeners[event]
		oldIndex := -1
		for i, ol := range oldL {
			if ol == newListener {
				oldIndex = i
				break
			}
		}
		if oldIndex == -1 {
			return
		}

		b.listeners[event] = append(oldL[:oldIndex], oldL[oldIndex+1:]...)
	}
}

// Publish publishes the given event and data to all listeners.
func (b *Broker) Publish(event string, data any) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for _, l := range b.listeners[event] {
		l.fn(event, data)
	}
}

// Subscribe subscribes to the given event on the default broker.
func Subscribe(event string, cb ListenerFunc) (unsubscribe func()) {
	return DefaultBroker.Subscribe(event, cb)
}

// Publish publishes the given event and data to all listeners on the
// default broker.
func Publish(event string, data any) {
	DefaultBroker.Publish(event, data)
}
