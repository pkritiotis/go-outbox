package outbox

import (
	"log"
	"time"
)

type ErrorType int

const (
	BrokerUnavailable ErrorType = iota
	Other
)

type BrokerError struct {
	Error error
	Type  ErrorType
}

//MessageBroker provides an interface for message brokers to send Message objects
type MessageBroker interface {
	Send(message Message) *BrokerError
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
	// While records are unprocessed, publish messages in batches
	for {
		err := d.lockUnprocessedEntities()
		if err != nil {
			return err
		}

		// Fetch next batch locked entities
		records, err := d.store.GetRecordsByLockID(d.machineID, d.settings.MaxSendAttempts)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			break
		}
		err = d.publishMessages(records)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d Dispatcher) publishMessages(records []Record) error {

	for _, rec := range records {
		// Publish message to message broker
		now := time.Now().UTC()
		rec.LastAttemptOn = &now
		rec.NumberOfAttempts++
		err := d.messageBroker.Send(rec.Message)

		// If an error occurs, remove lock information, update retrial times and continue
		if err != nil {
			// if the error is because of broker availability there is no reason to process the rest of the messages now - retry later
			if err.Type == BrokerUnavailable {
				return nil
			}

			rec.LockedOn = nil
			rec.LockID = nil
			errorMsg := err.Error.Error()
			rec.Error = &errorMsg
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
		err = d.store.UpdateRecordByID(rec)

		if err != nil {
			return err
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
	lockErr := d.store.UpdateRecordLockByState(d.machineID, lockTime, Unprocessed, d.settings.RecordProcessBatchNumber)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
