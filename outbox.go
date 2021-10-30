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

func (s Outbox) Add(msg Message, tx *sql.Tx) error {
	newID, _ := uuid.NewUUID()
	record := store.Record{
		ID: newID,
		Message: store.Message{
			Key:   msg.Key,
			Body:  msg.Body,
			Topic: msg.Topic,
		},
		State:       store.Unprocessed,
		CreatedOn:   time.Now().UTC(),
		LockID:      nil,
		LockedOn:    nil,
		ProcessedOn: nil,
	}
	for _, header := range msg.Headers {
		h := store.Header{
			Key:   header.Key,
			Value: header.Value,
		}
		record.Message.Headers = append(record.Message.Headers, h)
	}

	return s.store.SaveTx(record, tx)
}
