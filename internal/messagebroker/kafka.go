package messagebroker

import "github.com/pkritiotis/go-outbox/internal/outbox"

type Kafka struct {
}

func (k Kafka) Send(event outbox.Message) error {
	panic("implement me")
}
