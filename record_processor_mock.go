package outbox

import (
	"sync"

	"github.com/stretchr/testify/mock"
)

type mockRecordProcessor struct {
	wg sync.WaitGroup
	mock.Mock
}

func (m mockRecordProcessor) ProcessRecords() error {
	args := m.Called()
	return args.Error(0)
}
