package outbox

import (
	"database/sql"
	"time"
)

type StoreType int

type Store interface {
	SaveTx(message Message, tx *sql.Tx) error
	UpdateMessageLockByState(lockID string, lockedOn time.Time, state MessageState) error
	UpdateMessages(messages []Message) error
	UpdateMessageByID(message Message) error
	GetMessagesByLockID(lockID string) ([]Message, error)
	ClearLocksBeforeDate(time time.Time) error
	ClearLocksByLockID(lockID string) error
}
