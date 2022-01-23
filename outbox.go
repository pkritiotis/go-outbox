package outbox

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

//Outbox encapsulates the save functionality of the outbox pattern
type Outbox struct {
	store Store
}

//New is the Outbox constructor
func New(store Store) Outbox {
	return Outbox{store: store}
}

//MessageHeader is the MessageHeader of the Message to be sent. It is used by Brokers
type MessageHeader struct {
	Key   string
	Value string
}

//Message encapsulates the contents of the message to be sent
type Message struct {
	Key     string
	Headers []MessageHeader
	Body    []byte
	Topic   string
}

//Add stores the msg Message within the provided SQL tx
func (s Outbox) Add(msg Message, tx *sql.Tx) error {
	newID, _ := uuid.NewUUID()
	record := Record{
		ID:          newID,
		Message:     msg,
		State:       PendingDelivery,
		CreatedOn:   time.Now().UTC(),
		LockID:      nil,
		LockedOn:    nil,
		ProcessedOn: nil,
	}

	return s.store.AddRecordTx(record, tx)
}
