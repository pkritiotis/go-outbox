package outbox

const (
	BrokerUnavailable ErrorType = iota
	Other
)

type BrokerError struct {
	Error error
	Type  ErrorType
}

//MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(message Message) *BrokerError
}
