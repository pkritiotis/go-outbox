package outbox

import (
	"github.com/stretchr/testify/mock"
	"sync"
)

type mockRecordProcessor struct {
	wg sync.WaitGroup
	mock.Mock
}

func (m mockRecordProcessor) ProcessRecords() error {
	args := m.Called()
	return args.Error(0)
}
