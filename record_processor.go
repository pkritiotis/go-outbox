package outbox

import (
	"time"
)

//defaultRecordProcessor checks and dispatches new messages to be sent
type defaultRecordProcessor struct {
	messageBroker   MessageBroker
	store           Store
	machineID       string
	MaxSendAttempts int
}

//newProcessor constructs a new defaultRecordProcessor
func newProcessor(store Store, messageBroker MessageBroker, machineID string, maxSendAttempts int) *defaultRecordProcessor {
	return &defaultRecordProcessor{machineID: machineID, messageBroker: messageBroker, MaxSendAttempts: maxSendAttempts, store: store}
}

//ProcessRecords locks unprocessed messages, tries to deliver them and then unlocks them
func (d defaultRecordProcessor) processRecords() error {
	err := d.lockUnprocessedEntities()
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
	err = d.publishMessages(records)
	_ = d.store.ClearLocksByLockID(d.machineID)

	if err != nil {
		return err
	}
	return nil
}

func (d defaultRecordProcessor) publishMessages(records []Record) error {

	for _, rec := range records {
		// Publish message to message broker
		now := time.Now().UTC()
		rec.LastAttemptOn = &now
		rec.NumberOfAttempts++
		brokerErr := d.messageBroker.Send(rec.Message)

		// If an error occurs, remove lock information, update retrial times and continue
		if brokerErr != nil {
			// if the error is because of broker availability there is no reason to process the rest of the messages now - retry later
			if brokerErr.Type == BrokerUnavailable {
				return brokerErr.Error
			}

			rec.LockedOn = nil
			rec.LockID = nil
			errorMsg := brokerErr.Error.Error()
			rec.Error = &errorMsg
			if rec.NumberOfAttempts == d.MaxSendAttempts {
				rec.State = MaxAttemptsReached
			}
			_ = d.store.UpdateRecordByID(rec)

			continue
		}

		// Remove lock information and update state
		rec.State = Delivered
		rec.LockedOn = nil
		rec.LockID = nil
		rec.ProcessedOn = &now
		err := d.store.UpdateRecordByID(rec)

		if brokerErr != nil {
			return err
		}
	}
	return nil
}

// lockUnprocessedEntities updates the messages with the current machine's lockID
func (d defaultRecordProcessor) lockUnprocessedEntities() error {
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateRecordLockByState(d.machineID, lockTime, PendingDelivery)
	if lockErr != nil {
		return lockErr
	}
	return nil
}
