package outbox

import (
	"github.com/stretchr/testify/mock"
)

type mockRecordCleaner struct {
	mock.Mock
}

func (m *mockRecordCleaner) RemoveExpiredMessages() error {
	args := m.Called()
	return args.Error(0)
}
