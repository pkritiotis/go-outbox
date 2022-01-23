package outbox

import "github.com/stretchr/testify/mock"

type MockBroker struct {
	mock.Mock
}

func (m MockBroker) Send(message Message) *BrokerError {
	args := m.Called(message)
	return args.Get(0).(*BrokerError)
}
