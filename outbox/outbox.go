package outbox

import (
	"database/sql"
)

type Outbox struct {
	store Store
}

// NewOutbox Constructor
func NewOutbox(store Store) Outbox {
	return Outbox{store: store}
}

func (o Outbox) SendTx(msg Message, tx *sql.Tx) error {
	return o.store.SaveTx(Message{
		ID:      msg.ID,
		Headers: msg.Headers,
		Key:     msg.Key,
		Body:    msg.Body,
		Topic:   msg.Topic,
		Type:    msg.Type,
	}, tx)
}
