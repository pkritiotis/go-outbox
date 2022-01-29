package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/pkritiotis/outbox"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBroker_Send(t *testing.T) {
	tests := map[string]struct {
		broker *sarama.MockBroker
		config *sarama.Config
		event  outbox.Message
		expErr *outbox.BrokerError
	}{
		"Unsuccessful delivery should return error": {
			broker: func() *sarama.MockBroker {
				mp := sarama.NewMockBroker(t, 1)
				mp.SetHandlerByMap(map[string]sarama.MockResponse{
					"MetadataRequest": sarama.NewMockMetadataResponse(t).
						SetBroker(mp.Addr(), mp.BrokerID()).
						SetLeader("sampleTopic", 0, mp.BrokerID()),
					"ProduceRequest": sarama.NewMockProduceResponse(t),
				})

				return mp
			}(),
			config: func() *sarama.Config {
				mp := sarama.NewConfig()
				mp.Producer.Return.Successes = true
				mp.Producer.Partitioner = sarama.NewRandomPartitioner
				return mp
			}(),
			event: outbox.Message{
				Key:     "sampleKey",
				Headers: nil,
				Body:    sarama.ByteEncoder("testing"),
				Topic:   "sampleTopic",
			},
			expErr: &outbox.BrokerError{
				Error: sarama.ErrInsufficientData,
				Type:  outbox.Other,
			},
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			defer tt.broker.Close()
			b := Broker{
				brokers: []string{tt.broker.Addr()},
				config:  tt.config,
			}
			err := b.Send(tt.event)
			assert.Equal(t, tt.expErr, err)

		})
	}
}

func TestBroker_Send_ProducerError(t *testing.T) {
	b := NewBroker([]string{}, sarama.NewConfig())
	err := b.Send(outbox.Message{
		Key:     "etst",
		Headers: nil,
		Body:    []byte("test"),
		Topic:   "testTopic",
	})

	expErr := &outbox.BrokerError{
		Error: sarama.ConfigurationError("You must provide at least one broker address"),
		Type:  outbox.BrokerUnavailable,
	}

	assert.Equal(t, expErr, err)
}
func TestNewBroker(t *testing.T) {
	conf := sarama.NewConfig()
	brokers := []string{"localhost:9092"}

	b := NewBroker(brokers, conf)

	assert.NotNil(t, b)
	assert.Equal(t, brokers, b.brokers)
	assert.Equal(t, conf, b.config)
}
