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

func NewOutbox(store store.Store) Outbox {
	return Outbox{store: store}
}

type Message struct {
	Key     string
	Headers map[string]string
	Body    interface{}
	Topic   string
}

func (s Outbox) Add(msg Message, tx *sql.Tx) error {
	newID, _ := uuid.NewUUID()
	return s.store.SaveTx(store.Message{
		ID:          newID,
		Headers:     msg.Headers,
		Key:         msg.Key,
		Body:        msg.Body,
		Topic:       msg.Topic,
		State:       store.Unprocessed,
		CreatedOn:   time.Now().UTC(),
		LockID:      nil,
		LockedOn:    nil,
		ProcessedOn: nil,
	}, tx)
}
