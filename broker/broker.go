package broker

//Broker provides an interface for message brokers to send Message objects
type Broker interface {
	Send(message Message) error
}

type Header struct {
	Key   string
	Value string
}

type Message struct {
	Key     string
	Headers []Header
	Body    []byte
	Topic   string
}
