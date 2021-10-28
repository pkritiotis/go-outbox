package kafka

import (
	"fmt"
	"github.com/pkritiotis/outbox/broker"
)

type Broker struct {
}

func (k Broker) Send(event broker.Message) error {

	fmt.Printf("i'm here %v,%v\n", event, string(event.Body))
	return nil
}
