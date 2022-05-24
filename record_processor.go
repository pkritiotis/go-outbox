package outbox

import (
	"fmt"
	"github.com/pkritiotis/outbox/internal/time"
)

//defaultRecordProcessor checks and dispatches new messages to be sent
type defaultRecordProcessor struct {
	messageBroker MessageBroker
	store         Store
	time          time.Provider
	machineID     string
	retrialPolicy RetrialPolicy
}

//newProcessor constructs a new defaultRecordProcessor
func newProcessor(store Store, messageBroker MessageBroker, machineID string, retrialPolicy RetrialPolicy) *defaultRecordProcessor {
	return &defaultRecordProcessor{
		messageBroker: messageBroker,
		store:         store,
		time:          time.NewTimeProvider(),
		machineID:     machineID,
		retrialPolicy: retrialPolicy,
	}
}

//ProcessRecords locks unprocessed messages, tries to deliver them and then unlocks them
func (d defaultRecordProcessor) ProcessRecords() error {
	err := d.lockUnprocessedEntities()
	defer d.store.ClearLocksByLockID(d.machineID)
	if err != nil {
		return err
	}
	records, err := d.store.GetRecordsByLockID(d.machineID)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	return d.publishMessages(records)
}

func (d defaultRecordProcessor) publishMessages(records []Record) error {

	for _, rec := range records {
		// Send message to message broker
		now := d.time.Now().UTC()
		rec.LastAttemptOn = &now
		rec.NumberOfAttempts++
		err := d.messageBroker.Send(rec.Message)

		// If an error occurs, remove the lock information, update retrial times and continue
		if err != nil {
			rec.LockedOn = nil
			rec.LockID = nil
			errorMsg := err.Error()
			rec.Error = &errorMsg
			if d.retrialPolicy.MaxSendAttemptsEnabled && rec.NumberOfAttempts == d.retrialPolicy.MaxSendAttempts {
				rec.State = MaxAttemptsReached
			}
			dbErr := d.store.UpdateRecordByID(rec)
			if dbErr != nil {
				return fmt.Errorf("Could not update the record in the db: %w", dbErr)
			}

			return fmt.Errorf("An error occurred when trying to send the message to the broker: %w", err)
		}

		// Remove lock information and update state
		rec.State = Delivered
		rec.LockedOn = nil
		rec.LockID = nil
		rec.ProcessedOn = &now
		err = d.store.UpdateRecordByID(rec)

		if err != nil {
			return fmt.Errorf("Could not update the record in the db: %w", err)
		}
	}
	return nil
}

// lockUnprocessedEntities updates the messages with the current machine's lockID
func (d defaultRecordProcessor) lockUnprocessedEntities() error {
	lockTime := d.time.Now().UTC()
	lockErr := d.store.UpdateRecordLockByState(d.machineID, lockTime, PendingDelivery)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
