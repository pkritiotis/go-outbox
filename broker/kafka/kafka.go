package kafka

import (
	"github.com/Shopify/sarama"

	"github.com/pkritiotis/outbox"
)

//Broker implements the MessageBroker interface
type Broker struct {
	brokers []string
	config  *sarama.Config
}

//NewBroker constructor
func NewBroker(brokers []string, config *sarama.Config) *Broker {
	config.Producer.Return.Successes = true
	return &Broker{brokers: brokers, config: config}
}

//Send delivers the message to kafka
func (b Broker) Send(event outbox.Message) *outbox.BrokerError {

	producer, err := sarama.NewSyncProducer(b.brokers, b.config)
	if err != nil {
		return &outbox.BrokerError{
			Error: err,
			Type:  outbox.BrokerUnavailable,
		}
	}
	var headers []sarama.RecordHeader

	for i := 0; i < len(event.Headers); i++ {
		headers = append(headers, sarama.RecordHeader{
			Key:   sarama.ByteEncoder(event.Headers[i].Key),
			Value: sarama.ByteEncoder(event.Headers[i].Value),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic: event.Topic,
		Key:   sarama.StringEncoder(event.Key),
		Value: sarama.StringEncoder(event.Body),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return &outbox.BrokerError{
			Error: err,
			Type:  outbox.Other,
		}
	}
	return nil
}
