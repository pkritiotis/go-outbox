package store

import (
	"database/sql"
	"time"
)

type Store interface {
	SaveTx(message Record, tx *sql.Tx) error

	UpdateMessageLockByState(lockID string, lockedOn time.Time, state MessageState) error
	GetMessagesByLockID(lockID string) ([]Record, error)
	UpdateMessageByID(message Record) error

	ClearLocksWithDurationBeforeDate(duration time.Duration, time time.Time) error
	ClearLocksByLockID(lockID string) error
}
