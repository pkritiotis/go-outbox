package outbox

import (
	"time"
)

type recordUnlocker struct {
	store                   Store
	MaxLockTimeDurationMins time.Duration
}

func newRecordUnlocker(store Store, maxLockTimeDurationMins time.Duration) *recordUnlocker {
	return &recordUnlocker{MaxLockTimeDurationMins: maxLockTimeDurationMins, store: store}
}

func (d recordUnlocker) unlockExpiredMessages() error {
	duration := d.MaxLockTimeDurationMins
	expiryTime := time.Now().UTC()
	clearErr := d.store.ClearLocksWithDurationBeforeDate(duration, expiryTime)
	if clearErr != nil {
		return clearErr
	}
	return nil
}
