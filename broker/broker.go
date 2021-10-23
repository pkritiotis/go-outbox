package broker

//Broker provides an interface for message brokers to send Message objects
type Broker interface {
	Send(message Message) error
}

type Message struct {
	Key     string
	Headers map[string]string
	Body    interface{}
	Topic   string
}
