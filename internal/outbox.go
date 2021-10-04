package internal

type Type string
type Topic string

type Event struct {
	ID    int
	Data  []byte
	Topic Topic
	Type  Type
}

type EventState int

const (
	Unprocessed EventState = iota
	Processing
	Processed
)

type OutboxEvent struct {
	ID    int
	Data  []byte
	Topic Topic
	Type  Type
	State EventState
}

type Store interface {
	Save(event OutboxEvent)
	GetUnprocessed() []OutboxEvent
}

type Outbox interface {
	Send(event Event)
}

type outbox struct {
	store Store
}

// NewOutbox Constructor
func NewOutbox(store Store) *outbox {
	return &outbox{store: store}
}

func (o outbox) Send(event Event) {
	o.store.Save(OutboxEvent{
		ID:    event.ID,
		Data:  event.Data,
		Topic: event.Topic,
		Type:  event.Type,
		State: Unprocessed,
	})
}
