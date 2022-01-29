package kafka

import (
	"fmt"
	"github.com/pkritiotis/outbox"
)

//Broker implements the MessageBroker interface
type Broker struct {
}

//Send delivers the message to kafka
func (k Broker) Send(event outbox.Message) *outbox.BrokerError {
	fmt.Printf("i'm here %v,%v\n", event, string(event.Body))
	return nil
}
