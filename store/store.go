package store

import (
	"database/sql"
	"time"
)

type Store interface {
	SaveTx(message Message, tx *sql.Tx) error

	UpdateMessageLockByState(lockID string, lockedOn time.Time, state MessageState) error
	GetMessagesByLockID(lockID string) ([]Message, error)
	UpdateMessageByID(message Message) error

	ClearLocksWithDurationBeforeDate(duration time.Duration, time time.Time) error
	ClearLocksByLockID(lockID string) error
}
