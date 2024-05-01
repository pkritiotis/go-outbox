package outbox

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Record represents the record that is stored and retrieved from the database
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

// RecordState is the State of the Record
type RecordState int

const (
	// PendingDelivery is the initial state of all records
	PendingDelivery RecordState = iota
	// Delivered indicates that the Records is already Delivered
	Delivered
	// MaxAttemptsReached indicates that the message is not Delivered but the max attempts are reached so it shouldn't be delivered
	MaxAttemptsReached
)

// Store is the interface that should be implemented by SQL-like database drivers to support the outbox functionality
type Store interface {
	// AddRecordTx stores the message within the provided database transaction
	AddRecordTx(record Record, tx *sql.Tx) error
	// GetRecordsByLockID returns the records by lockID
	GetRecordsByLockID(lockID string) ([]Record, error)
	// UpdateRecordLockByState updates the lock of all records with the provided state
	UpdateRecordLockByState(lockID string, lockedOn time.Time, state RecordState) error
	// UpdateRecordByID updates the provided the record
	UpdateRecordByID(message Record) error
	// ClearLocksWithDurationBeforeDate clears the locks of records with a lock time before the provided time
	ClearLocksWithDurationBeforeDate(time time.Time) error
	// ClearLocksByLockID clears all records locked by the provided lockID
	ClearLocksByLockID(lockID string) error
	// RemoveRecordsBeforeDatetime removes all records before the provided time
	RemoveRecordsBeforeDatetime(expiryTime time.Time) error
}
