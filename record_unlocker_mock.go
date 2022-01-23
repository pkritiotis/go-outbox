package outbox

import (
	"github.com/stretchr/testify/mock"
	"sync"
)

type MockRecordUnlocker struct {
	wg sync.WaitGroup
	mock.Mock
}

func (m MockRecordUnlocker) unlockExpiredMessages() error {
	defer m.wg.Done()
	args := m.Called()
	return args.Error(0)
}
