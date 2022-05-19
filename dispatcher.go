package outbox

import (
	"log"
	"time"
)

type processor interface {
	ProcessRecords() error
}

type unlocker interface {
	UnlockExpiredMessages() error
}

// RetrialPolicy contains the retrial settings
type RetrialPolicy struct {
	MaxSendAttemptsEnabled bool
	MaxSendAttempts        int
}

// DispatcherSettings defines the set of configurations for the dispatcher
type DispatcherSettings struct {
	ProcessInterval     time.Duration
	LockCheckerInterval time.Duration
	MaxLockTimeDuration time.Duration
	RetrialPolicy       RetrialPolicy
}

//Dispatcher initializes and runs the outbox dispatcher
type Dispatcher struct {
	recordProcessor processor
	recordUnlocker  unlocker
	settings        DispatcherSettings
}

//NewDispatcher constructor
func NewDispatcher(store Store, broker MessageBroker, settings DispatcherSettings, machineID string) *Dispatcher {
	return &Dispatcher{
		recordProcessor: newProcessor(
			store,
			broker,
			machineID,
			settings.RetrialPolicy,
		),
		recordUnlocker: newRecordUnlocker(
			store,
			time.Duration(settings.MaxLockTimeDuration)*time.Minute,
		),
		settings: settings,
	}
}

//Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
//and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error, doneChan <-chan struct{}) {
	doneProc := make(chan struct{}, 1)
	doneUnlock := make(chan struct{}, 1)

	go func() {
		<-doneChan
		doneProc <- struct{}{}
		doneUnlock <- struct{}{}
	}()

	go d.runRecordProcessor(errChan, doneProc)
	go d.runRecordUnlocker(errChan, doneUnlock)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(d.settings.ProcessInterval) * time.Second)
	for {
		log.Print("Record processor Running")
		err := d.recordProcessor.ProcessRecords()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processing Finished")

		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record processor")
			return
		}
	}
}

func (d Dispatcher) runRecordUnlocker(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(d.settings.LockCheckerInterval) * time.Second)
	for {
		log.Print("Record unlocker Running")
		err := d.recordUnlocker.UnlockExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record unlocker Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record unlocker")
			return

		}
	}
}
