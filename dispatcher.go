package outbox

import (
	"github.com/pkritiotis/outbox/broker"
	"github.com/pkritiotis/outbox/store"
	"time"
)

type DispatcherSettings struct {
	MachineID                  string
	ProcessIntervalSeconds     int
	LockCheckerIntervalSeconds int
	MaxLockTimeDurationMins    int
}

type Dispatcher struct {
	store    store.Store
	broker   broker.Broker
	settings DispatcherSettings
}

// NewDispatcher initializes the dispatcher worker
func NewDispatcher(store store.Store, broker broker.Broker, settings DispatcherSettings) Dispatcher {
	return Dispatcher{
		store:    store,
		broker:   broker,
		settings: settings,
	}
}

//StartDispatcher periodically checks for new outbox messages from the Store, sends the messages through the Broker
//and updates the message status accordingly
func (d Dispatcher) Run() error {

	processErr := d.runMessageProcessor()
	if processErr != nil {
		return processErr
	}

	unlockerErr := d.runMessageUnlocker()
	if unlockerErr != nil {
		return unlockerErr
	}

	return nil
}

func (d Dispatcher) runMessageProcessor() error {
	lockID := d.settings.MachineID
	for {
		err := d.processMessages(lockID)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(d.settings.ProcessIntervalSeconds) * time.Second)
	}
	return nil
}

func (d Dispatcher) processMessages(lockID string) error {
	// 1. Lock unsent entities
	lockErr := lockUnprocessedEntities(d, lockID)
	if lockErr != nil {
		return lockErr
	}

	// 2. Fetch locked entities
	messages, msgErr := d.store.GetMessagesByLockID(lockID)
	if msgErr != nil {
		return msgErr
	}

	// 3. Process messages
	for _, msg := range messages {
		// 4a. Add message to kafka
		brokerErr := sendMessageToBroker(d, msg)
		if brokerErr != nil {
			msg.LockID = nil
			msg.LockedOn = nil
			_ = d.store.UpdateMessageByID(msg)
			continue
		}

		// 4b. Record sent successfully - Remove lock information and update state to processed
		removeLockErr := removeLockFromMessage(msg, d)
		if removeLockErr != nil {
			// if an update of a message failed try to clear the rest of the locks and return
			_ = d.store.ClearLocksByLockID(lockID)
			return removeLockErr
		}
	}
	return nil
}

func (d Dispatcher) runMessageUnlocker() error {
	for {
		err := d.unlockExpiredMessages()
		if err != nil {
			// won't do anything here
		}
		time.Sleep(time.Duration(d.settings.LockCheckerIntervalSeconds) * time.Second)
	}
}

func (d Dispatcher) unlockExpiredMessages() error {
	duration := time.Duration(d.settings.MaxLockTimeDurationMins) * time.Minute
	expiryTime := time.Now().UTC()
	clearErr := d.store.ClearLocksWithDurationBeforeDate(duration, expiryTime)
	if clearErr != nil {
		return clearErr
	}
	return nil
}

func removeLockFromMessage(msg store.Record, d Dispatcher) error {
	msg.LockedOn = nil
	msg.LockID = nil
	msg.State = store.Processed
	err := d.store.UpdateMessageByID(msg)
	return err
}

func sendMessageToBroker(d Dispatcher, rec store.Record) error {
	msg := broker.Message{
		Key:   rec.Message.Key,
		Body:  rec.Message.Body,
		Topic: rec.Message.Topic,
	}
	for _, h := range rec.Message.Headers {
		msg.Headers = append(msg.Headers, broker.Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	err := d.broker.Send(msg)
	return err
}

// lockUnprocessedEntities updates the messages with the current machine's lockID
func lockUnprocessedEntities(d Dispatcher, lockID string) error {
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateMessageLockByState(lockID, lockTime, store.Unprocessed)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
