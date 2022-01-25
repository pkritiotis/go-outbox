package outbox

import (
	"github.com/stretchr/testify/mock"
)

type mockRecordUnlocker struct {
	mock.Mock
}

func (m *mockRecordUnlocker) unlockExpiredMessages() error {
	args := m.Called()
	return args.Error(0)
}
