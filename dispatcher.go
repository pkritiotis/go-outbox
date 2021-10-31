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
	ProcessIntervalSeconds     int
	LockCheckerIntervalSeconds int
	MaxLockTimeDurationMins    int
	RecordProcessBatchNumber   int
	MaxSendAttempts            int
	TimeBetweenAttemptsSec     int
}

type Dispatcher struct {
	machineID     string
	store         Store
	messageBroker MessageBroker
	settings      DispatcherSettings
}

// NewDispatcher initializes the dispatcher worker
func NewDispatcher(store Store, broker MessageBroker, settings DispatcherSettings, machineID string) Dispatcher {
	return Dispatcher{
		store:         store,
		messageBroker: broker,
		settings:      settings,
		machineID:     machineID,
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
	for {
		log.Print("Record Processor Running")

		err := d.processRecords()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processor Finished")

		time.Sleep(time.Duration(d.settings.ProcessIntervalSeconds) * time.Second)
	}
}

func (d Dispatcher) processRecords() error {
	lockErr := d.lockUnprocessedEntities()
	if lockErr != nil {
		return lockErr
	}

	// While records are unprocessed, publish messages in batches
	for {
		// Fetch next batch locked entities
		messages, msgErr := d.store.GetRecordsByLockID(d.machineID, d.settings.RecordProcessBatchNumber, d.settings.MaxSendAttempts)
		if msgErr != nil {
			return msgErr
		}
		if len(messages) == 0 {
			break
		}
		err := d.publishMessages(messages)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d Dispatcher) publishMessages(records []Record) error {

	for _, rec := range records {
		// Publish message to message broker
		brokerErr := d.messageBroker.Send(rec.Message)
		now := time.Now().UTC()
		rec.NumberOfAttempts++
		rec.LastAttemptOn = &now

		// If an error occurs, remove lock information and update retrial times
		if brokerErr != nil {
			rec.LockedOn = nil
			rec.LockID = nil
			if rec.NumberOfAttempts == d.settings.MaxSendAttempts {
				rec.State = MaxAttemptsReached
			}
			_ = d.store.UpdateRecordByID(rec)
			continue
		}

		// Remove lock information and update state
		rec.State = Processed
		rec.LockedOn = nil
		rec.LockID = nil
		rec.ProcessedOn = &now
		removeLockErr := d.store.UpdateRecordByID(rec)

		if removeLockErr != nil {
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

// lockUnprocessedEntities updates the messages with the current machine's lockID
func (d Dispatcher) lockUnprocessedEntities() error {
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateRecordLockByState(d.machineID, lockTime, Unprocessed)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
