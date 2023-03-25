package outbox

import (
	"github.com/pkritiotis/outbox/logs"
	"time"
)

type processor interface {
	ProcessRecords() error
}

type unlocker interface {
	UnlockExpiredMessages() error
}

type cleaner interface {
	RemoveExpiredMessages() error
}

// RetrialPolicy contains the retrial settings
type RetrialPolicy struct {
	MaxSendAttemptsEnabled bool
	MaxSendAttempts        int
}

// DispatcherSettings defines the set of configurations for the dispatcher
type DispatcherSettings struct {
	ProcessInterval           time.Duration
	LockCheckerInterval       time.Duration
	MaxLockTimeDuration       time.Duration
	CleanupWorkerInterval     time.Duration
	RetrialPolicy             RetrialPolicy
	MessagesRetentionDuration time.Duration
}

// Dispatcher initializes and runs the outbox dispatcher
type Dispatcher struct {
	recordProcessor processor
	recordUnlocker  unlocker
	recordCleaner   cleaner
	settings        DispatcherSettings
	logger          logs.LoggerAdapter
}

// NewDispatcher constructor
func NewDispatcher(store Store, logger logs.LoggerAdapter, broker MessageBroker, settings DispatcherSettings, machineID string) *Dispatcher {
	return &Dispatcher{
		recordProcessor: newProcessor(
			store,
			broker,
			machineID,
			settings.RetrialPolicy,
		),
		recordUnlocker: newRecordUnlocker(
			store,
			settings.MaxLockTimeDuration,
		),
		recordCleaner: newRecordCleaner(
			store,
			settings.MessagesRetentionDuration,
		),
		settings: settings,
		logger:   logger,
	}
}

// Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
// and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error, doneChan <-chan struct{}) {
	doneProc := make(chan struct{}, 1)
	doneUnlock := make(chan struct{}, 1)
	doneClear := make(chan struct{}, 1)

	go func() {
		<-doneChan
		doneProc <- struct{}{}
		doneUnlock <- struct{}{}
		doneClear <- struct{}{}
	}()

	go d.runRecordProcessor(errChan, doneProc)
	go d.runRecordUnlocker(errChan, doneUnlock)
	go d.runRecordCleaner(errChan, doneClear)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.ProcessInterval)
	for {
		d.logger.Info("Record processor Running", nil)
		err := d.recordProcessor.ProcessRecords()
		if err != nil {
			errChan <- err
		}
		d.logger.Info("Record Processing Finished", nil)

		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			d.logger.Info("Stopping Record processor", nil)
			return
		}
	}
}

func (d Dispatcher) runRecordUnlocker(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.LockCheckerInterval)
	for {
		d.logger.Info("Record unlocker Running", nil)
		err := d.recordUnlocker.UnlockExpiredMessages()
		if err != nil {
			errChan <- err
		}
		d.logger.Info("Record unlocker Finished", nil)
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			d.logger.Info("Stopping Record unlocker", nil)
			return

		}
	}
}

func (d Dispatcher) runRecordCleaner(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.CleanupWorkerInterval)
	for {
		d.logger.Info("Record retention cleaner Running", nil)
		err := d.recordCleaner.RemoveExpiredMessages()
		if err != nil {
			errChan <- err
		}
		d.logger.Info("Record retention cleaner Finished", nil)
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			d.logger.Info("Stopping Record retention cleaner", nil)
			return

		}
	}
}
