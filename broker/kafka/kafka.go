package kafka

import (
	"github.com/pkritiotis/outbox/broker"
)

type Broker struct {
}

func (k Broker) Send(event broker.Message) error {
	panic("implement me")
}
