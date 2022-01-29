package outbox

import "github.com/stretchr/testify/mock"

//MockBroker mocks the Broker interface
type MockBroker struct {
	mock.Mock
}

//Send method mock
func (m *MockBroker) Send(message Message) *BrokerError {
	args := m.Called(message)
	return args.Get(0).(*BrokerError)
}
