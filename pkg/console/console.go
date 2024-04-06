// Package console provides I/O with the console.
package console

import "github.com/jfhamlin/muscrat/pkg/pubsub"

type (
	Level string

	Message struct {
		Level   Level  `json:"level"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
)

const (
	// Debug level
	Debug Level = "debug"
	// Info level
	Info Level = "info"
	// Warn level
	Warn Level = "warn"
	// Error level
	Error Level = "error"
)

// Log logs a message to the console.
func Log(level Level, message string, data any) {
	pubsub.Publish("console.log", Message{
		Level:   level,
		Message: message,
		Data:    data,
	})
}
