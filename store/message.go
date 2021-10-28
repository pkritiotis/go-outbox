package store

import (
	"github.com/google/uuid"
	"time"
)

/*
CREATE TABLE outbox (
	id varchar(100) NOT NULL,
	data varchar(100) NOT NULL,
	state INT NOT NULL,
	created_on DATETIME NOT NULL,
	lock_id varchar(100) NULL,
	locked_on DATETIME NULL,
	processed_on DATETIME NULL
)
ENGINE=InnoDB
DEFAULT CHARSET=latin1
COLLATE=latin1_general_ci;

*/
type Message struct {
	Headers map[string]string
	Key     string
	Body    []byte
	Topic   string
}
type Record struct {
	ID          uuid.UUID
	Message     Message
	State       MessageState
	CreatedOn   time.Time
	LockID      *string
	LockedOn    *time.Time
	ProcessedOn *time.Time
}

type MessageState int

const (
	Unprocessed MessageState = iota
	Processing
	Processed
)
