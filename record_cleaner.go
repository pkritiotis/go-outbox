package outbox

import (
	"github.com/pkritiotis/outbox/internal/time"
	time2 "time"
)

type recordCleaner struct {
	store             Store
	time              time.Provider
	MaxRecordLifetime time2.Duration
}

func newRecordCleaner(store Store, maxRecordLifetime time2.Duration) recordCleaner {
	return recordCleaner{MaxRecordLifetime: maxRecordLifetime, store: store, time: time.NewTimeProvider()}
}

func (d recordCleaner) RemoveExpiredMessages() error {
	expiryTime := d.time.Now().UTC().Add(-d.MaxRecordLifetime)
	err := d.store.RemoveRecordsBeforeDatetime(expiryTime)
	if err != nil {
		return err
	}
	return nil
}
