package kafka

import (
	"github.com/Shopify/sarama"

	"github.com/pkritiotis/outbox"
)

// Broker implements the MessageBroker interface
type Broker struct {
	producer sarama.SyncProducer
}

// NewBroker constructor
func NewBroker(brokers []string, config *sarama.Config) (*Broker, error) {
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Broker{producer: producer}, nil
}

// Send delivers the message to kafka
func (b Broker) Send(event outbox.Message) error {
	var headers []sarama.RecordHeader

	for k, v := range event.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   sarama.ByteEncoder(k),
			Value: sarama.ByteEncoder(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic: event.Topic,
		Key:   sarama.StringEncoder(event.Key),
		Value: sarama.StringEncoder(event.Body),
	}
	_, _, err := b.producer.SendMessage(msg)

	return err
}
