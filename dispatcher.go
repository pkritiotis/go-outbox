package outbox

import (
	"log"
	"time"
)

type ErrorType int

type Processor interface {
	processRecords() error
}

type Unlocker interface {
	unlockExpiredMessages() error
}

// DispatcherSettings defines the set of configurations for the dispatcher
type DispatcherSettings struct {
	ProcessIntervalSeconds     int
	LockCheckerIntervalSeconds int
	MaxLockTimeDurationMins    int
	MaxSendAttempts            int
	TimeBetweenAttemptsSec     int
}

type Dispatcher struct {
	recordProcessor Processor
	recordUnlocker  Unlocker
	settings        DispatcherSettings
}

func NewDispatcher(store Store, broker MessageBroker, settings DispatcherSettings, machineID string) *Dispatcher {
	return &Dispatcher{
		recordProcessor: newProcessor(
			store,
			broker,
			machineID,
			settings.MaxSendAttempts,
		),
		recordUnlocker: newRecordUnlocker(
			store,
			time.Duration(settings.MaxLockTimeDurationMins)*time.Minute,
		),
		settings: settings,
	}
}

//Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
//and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error, doneChan <-chan bool) {
	go d.runRecordProcessor(errChan, doneChan)
	go d.runRecordUnlocker(errChan, doneChan)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error, doneChan <-chan bool) {
	ticker := time.NewTicker(time.Duration(d.settings.ProcessIntervalSeconds) * time.Second)
	for {
		log.Print("Record Processor Running")
		err := d.recordProcessor.processRecords()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processing Finished")

		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record Processor")
			return
		}
	}
}

func (d Dispatcher) runRecordUnlocker(errChan chan<- error, doneChan <-chan bool) {
	ticker := time.NewTicker(time.Duration(d.settings.LockCheckerIntervalSeconds) * time.Second)
	for {
		log.Print("Record Unlocker Running")
		err := d.recordUnlocker.unlockExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Unlocker Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record Unlocker")
			return

		}
	}
}
