package outbox

import (
	"github.com/stretchr/testify/mock"
)

type mockRecordUnlocker struct {
	mock.Mock
}

func (m *mockRecordUnlocker) UnlockExpiredMessages() error {
	args := m.Called()
	return args.Error(0)
}
