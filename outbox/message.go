package outbox

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID          uuid.UUID
	Headers     map[string]string
	Key         string
	Body        interface{}
	Topic       MessageTopic
	Type        MessageType
	State       MessageState
	CreatedOn   *time.Time
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

type MessageType string

type MessageTopic string
