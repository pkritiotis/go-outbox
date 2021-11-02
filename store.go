package outbox

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Record struct {
	ID               uuid.UUID
	Message          Message
	State            RecordState
	CreatedOn        time.Time
	LockID           *string
	LockedOn         *time.Time
	ProcessedOn      *time.Time
	NumberOfAttempts int
	LastAttemptOn    *time.Time
	Error            *string
}

type Store interface {
	AddRecordTx(message Record, tx *sql.Tx) error

	GetRecordsByLockID(lockID string, maxSendAttempts int) ([]Record, error)
	UpdateRecordLockByState(lockID string, lockedOn time.Time, state RecordState, numberOfRecords int) error
	UpdateRecordByID(message Record) error

	ClearLocksWithDurationBeforeDate(duration time.Duration, time time.Time) error
	ClearLocksByLockID(lockID string) error
}
