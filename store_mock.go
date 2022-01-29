package outbox

import (
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

//MockStore mocks the Store
type MockStore struct {
	mock.Mock
}

//AddRecordTx method mock
func (m *MockStore) AddRecordTx(record Record, tx *sql.Tx) error {
	args := m.Called(record, tx)
	return args.Error(0)
}

//GetRecordsByLockID method mock
func (m *MockStore) GetRecordsByLockID(lockID string) ([]Record, error) {
	args := m.Called(lockID)
	return args.Get(0).([]Record), args.Error(1)
}

//UpdateRecordLockByState method mock
func (m *MockStore) UpdateRecordLockByState(lockID string, lockedOn time.Time, state RecordState) error {
	args := m.Called(lockID, lockedOn, state)
	return args.Error(0)
}

//UpdateRecordByID method mock
func (m *MockStore) UpdateRecordByID(message Record) error {
	args := m.Called(message)
	return args.Error(0)
}

//ClearLocksWithDurationBeforeDate method mock
func (m *MockStore) ClearLocksWithDurationBeforeDate(time time.Time) error {
	args := m.Called(time)
	return args.Error(0)
}

//ClearLocksByLockID method mock
func (m *MockStore) ClearLocksByLockID(lockID string) error {
	args := m.Called(lockID)
	return args.Error(0)
}
