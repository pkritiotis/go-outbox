package outbox

type Message struct {
	ID    int
	Data  []byte
	Topic MessageTopic
	Type  MessageType
	state MessageState
}

type MessageType string

type MessageTopic string

type MessageState int

const (
	Unprocessed MessageState = iota
	Processing
	Processed
)
