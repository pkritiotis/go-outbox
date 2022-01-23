package outbox

import (
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m MockStore) AddRecordTx(record Record, tx *sql.Tx) error {
	args := m.Called(record, tx)
	return args.Error(0)
}

func (m MockStore) GetRecordsByLockID(lockID string) ([]Record, error) {
	args := m.Called(lockID)
	return args.Get(0).([]Record), args.Error(1)
}

func (m MockStore) UpdateRecordLockByState(lockID string, lockedOn time.Time, state RecordState) error {
	args := m.Called(lockID, lockedOn, state)
	return args.Error(0)
}

func (m MockStore) UpdateRecordByID(message Record) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m MockStore) ClearLocksWithDurationBeforeDate(duration time.Duration, time time.Time) error {
	args := m.Called(duration, time)
	return args.Error(0)
}

func (m MockStore) ClearLocksByLockID(lockID string) error {
	args := m.Called(lockID)
	return args.Error(0)
}
