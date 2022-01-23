package outbox

import (
	"github.com/stretchr/testify/mock"
	"sync"
)

type mockRecordUnlocker struct {
	wg sync.WaitGroup
	mock.Mock
}

func (m mockRecordUnlocker) unlockExpiredMessages() error {
	defer m.wg.Done()
	args := m.Called()
	return args.Error(0)
}
