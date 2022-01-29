package outbox

const (
	//BrokerUnavailable represents broker errors that are related to availability
	BrokerUnavailable BrokerErrorType = iota
	//Other represents non recoverable borker errors
	Other
)

//BrokerErrorType contains the type of broker errors
type BrokerErrorType int

//BrokerError encapsulates the error and type of broker errors
type BrokerError struct {
	Error error
	Type  BrokerErrorType
}

//MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(message Message) *BrokerError
}
