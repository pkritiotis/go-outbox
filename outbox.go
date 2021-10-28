package outbox

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/pkritiotis/outbox/store"
	"time"
)

type Outbox struct {
	store store.Store
}

func New(store store.Store) Outbox {
	return Outbox{store: store}
}

type Message struct {
	Key     string
	Headers map[string]string
	Body    []byte
	Topic   string
}

func (s Outbox) Add(msg Message, tx *sql.Tx) error {
	newID, _ := uuid.NewUUID()
	return s.store.SaveTx(store.Record{
		ID: newID,
		Message: store.Message{
			Headers: msg.Headers,
			Key:     msg.Key,
			Body:    msg.Body,
			Topic:   msg.Topic,
		},
		State:       store.Unprocessed,
		CreatedOn:   time.Now().UTC(),
		LockID:      nil,
		LockedOn:    nil,
		ProcessedOn: nil,
	}, tx)
}
