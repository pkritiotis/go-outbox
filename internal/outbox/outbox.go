package outbox

import "errors"

type Outbox struct {
	store Store
}

// NewOutbox Constructor
func NewOutbox(store Store) Outbox {
	return Outbox{store: store}
}

func (o Outbox) SendTx(msg Message, tx interface{}) error {
	switch o.store.Type() {
	case SQL:
		err := o.store.SaveTx(Message{
			ID:    msg.ID,
			Data:  msg.Data,
			Topic: msg.Topic,
			Type:  msg.Type,
			state: Unprocessed,
		}, tx)
		if err != nil {
			return err
		}
	default:
		return errors.New("store provider not supporter")
	}
	return nil
}
