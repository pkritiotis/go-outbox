package outbox

import (
	"log"
	"time"
)

//MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(message Message) error
}

// DispatcherSettings defines the set of configurations for the dispatcher
type DispatcherSettings struct {
	MachineID                  string
	ProcessIntervalSeconds     int
	LockCheckerIntervalSeconds int
	MaxLockTimeDurationMins    int
}

type Dispatcher struct {
	store         Store
	messageBroker MessageBroker
	settings      DispatcherSettings
}

// NewDispatcher initializes the dispatcher worker
func NewDispatcher(store Store, broker MessageBroker, settings DispatcherSettings) Dispatcher {
	return Dispatcher{
		store:         store,
		messageBroker: broker,
		settings:      settings,
	}
}

//Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
//and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error) {

	go d.runRecordProcessor(errChan)

	go d.runRecordUnlocker(errChan)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error) {
	log.Print("Record Processor Started")
	lockID := d.settings.MachineID
	for {
		log.Print("Record Processor Running")

		err := d.processRecords(lockID)
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processor Finished")

		time.Sleep(time.Duration(d.settings.ProcessIntervalSeconds) * time.Second)
	}
}

func (d Dispatcher) processRecords(lockID string) error {
	// 1. Lock unsent entities
	lockErr := lockUnprocessedEntities(d, lockID)
	if lockErr != nil {
		return lockErr
	}

	// 2. Fetch locked entities
	messages, msgErr := d.store.GetRecordByLockID(lockID)
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
			_ = d.store.UpdateRecordByID(msg)
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

func (d Dispatcher) runRecordUnlocker(errChan chan<- error) {
	log.Print("Record Unlocker Started")
	for {
		log.Print("Record Unlocker Running")

		err := d.unlockExpiredMessages()
		if err != nil {
			// won't do anything here
		}
		log.Print("Record Unlocker finished")

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

func removeLockFromMessage(rec Record, d Dispatcher) error {
	rec.LockedOn = nil
	rec.LockID = nil
	rec.State = Processed
	err := d.store.UpdateRecordByID(rec)
	return err
}

func sendMessageToBroker(d Dispatcher, rec Record) error {
	err := d.messageBroker.Send(rec.Message)
	return err
}

// lockUnprocessedEntities updates the messages with the current machine's lockID
func lockUnprocessedEntities(d Dispatcher, lockID string) error {
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateRecordLockByState(lockID, lockTime, Unprocessed)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
