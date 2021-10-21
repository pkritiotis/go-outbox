package outbox

import (
	"github.com/denisbrodbeck/machineid"
	"time"
)

//MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(event Message) error
}

//Dispatcher periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
//and updates the message status accordingly
type Dispatcher struct {
	store  Store
	broker MessageBroker
}

func (d Dispatcher) Start() error {
	// 1. Get the unique machine id
	lockID, idErr := machineid.ID()
	if idErr != nil {
		return idErr
	}

	// 2. Lock unsent entities
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateMessageLockByState(lockID, lockTime, Unprocessed)
	if lockErr != nil {
		return lockErr
	}

	// 3. Fetch locked entities
	messages, msgErr := d.store.GetMessagesByLockID(lockID)
	if msgErr != nil {
		return msgErr
	}

	// 4. Process messages
	for _, msg := range messages {
		// 4a. Send message to kafka
		err := d.broker.Send(Message{
			ID:    msg.ID,
			Body:  msg.Body,
			Topic: msg.Topic,
			Type:  msg.Type,
		})
		if err != nil {
			msg.LockID = nil
			msg.LockedOn = nil
			_ = d.store.UpdateMessageByID(msg)
			continue
		}

		// 4b. Message sent successfully - Remove lock information and update state to processed
		msg.LockedOn = nil
		msg.LockID = nil
		msg.State = Processed
		err = d.store.UpdateMessageByID(msg)
		if err != nil {
			// if an update of a message failed try to clear the rest of the locks and return
			_ = d.store.ClearLocksByLockID(lockID)
			return err
		}
	}
	return nil
}
