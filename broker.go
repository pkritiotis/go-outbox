// Package outbox provides an interface for message brokers to send Message objects
package outbox

// MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(message Message) error
}
