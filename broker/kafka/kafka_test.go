package kafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/pkritiotis/outbox"
	"github.com/stretchr/testify/assert"
)

func TestBroker_Send(t *testing.T) {
	tests := map[string]struct {
		broker *sarama.MockBroker
		config *sarama.Config
		event  outbox.Message
		expErr error
	}{
		"Unsuccessful delivery should return error": {
			broker: func() *sarama.MockBroker {
				mp := sarama.NewMockBroker(t, 1)
				mp.SetHandlerByMap(map[string]sarama.MockResponse{
					"MetadataRequest": sarama.NewMockMetadataResponse(t).
						SetBroker(mp.Addr(), mp.BrokerID()).
						SetLeader("sampleTopic", 0, mp.BrokerID()),
					"ProduceRequest": sarama.NewMockProduceResponse(t).SetError("sampleTopic", 0, sarama.ErrBrokerNotAvailable),
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
				Key: "sampleKey",
				Headers: []outbox.MessageHeader{
					{
						Key:   "testKey",
						Value: "testValue",
					},
				},
				Body:  sarama.ByteEncoder("testing"),
				Topic: "sampleTopic",
			},
			expErr: sarama.KError(sarama.ErrBrokerNotAvailable),
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			defer tt.broker.Close()
			producer, err := sarama.NewSyncProducer([]string{tt.broker.Addr()}, tt.config)
			assert.Nil(t, err)
			b := Broker{
				producer: producer,
			}
			err = b.Send(tt.event)
			assert.Equal(t, tt.expErr, err)
		})
	}
}

func TestNewBroker_error(t *testing.T) {
	expErr := sarama.ConfigurationError("You must provide at least one broker address")

	b, err := NewBroker([]string{}, sarama.NewConfig())

	assert.Nil(t, b)
	assert.Equal(t, expErr, err)
}

func TestNewBroker_success(t *testing.T) {
	mp := sarama.NewMockBroker(t, 1)
	mp.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mp.Addr(), mp.BrokerID()).
			SetLeader("sampleTopic", 0, mp.BrokerID()),
		"ProduceRequest": sarama.NewMockProduceResponse(t),
	})
	conf := sarama.NewConfig()
	brokers := []string{mp.Addr()}

	b, err := NewBroker(brokers, conf)

	assert.Nil(t, err)
	assert.NotNil(t, b)
}
