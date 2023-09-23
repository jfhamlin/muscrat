package pubsub

import "testing"

func TestBroker(t *testing.T) {
	b := NewBroker()

	received := make(chan any, 1)

	unsub := b.Subscribe("test", func(event string, data any) {
		received <- data
	})

	b.Publish("test", "hello")
	if <-received != "hello" {
		t.Fatal("expected 'hello'")
	}

	unsub()

	b.Publish("test", "world")

	select {
	case <-received:
		t.Fatal("expected no message")
	default:
	}
}
