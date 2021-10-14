package outbox

type MessageBroker interface {
	Send(event Message) error
}

type Dispatcher struct {
	store  Store
	broker MessageBroker
}

func (d Dispatcher) Start() {
	messages := d.store.GetUnprocessed()
	for _, msg := range messages {
		err := d.broker.Send(Message{
			ID:    msg.ID,
			Data:  msg.Data,
			Topic: msg.Topic,
			Type:  msg.Type,
		})
		if err != nil {
			//retry logic
			continue
		}
		msg.state = Processed
		d.store.Update(msg)
	}
}
