package outbox

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Outbox struct {
	store Store
}

func New(store Store) Outbox {
	return Outbox{store: store}
}

type Header struct {
	Key   string
	Value string
}

type Message struct {
	Key     string
	Headers []Header
	Body    []byte
	Topic   string
}

type RecordState int

const (
	Unprocessed RecordState = iota
	Processed
)

func (s Outbox) Add(msg Message, tx *sql.Tx) error {
	newID, _ := uuid.NewUUID()
	record := Record{
		ID:          newID,
		Message:     msg,
		State:       Unprocessed,
		CreatedOn:   time.Now().UTC(),
		LockID:      nil,
		LockedOn:    nil,
		ProcessedOn: nil,
	}

	return s.store.AddRecordTx(record, tx)
}
