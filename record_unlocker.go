package outbox

import (
	time2 "time"

	"github.com/pkritiotis/outbox/internal/time"
)

type recordUnlocker struct {
	store                   Store
	time                    time.Provider
	MaxLockTimeDurationMins time2.Duration
}

func newRecordUnlocker(store Store, maxLockTimeDurationMins time2.Duration) recordUnlocker {
	return recordUnlocker{MaxLockTimeDurationMins: maxLockTimeDurationMins, store: store, time: time.NewTimeProvider()}
}

func (d recordUnlocker) UnlockExpiredMessages() error {
	expiryTime := d.time.Now().UTC().Add(-d.MaxLockTimeDurationMins)
	clearErr := d.store.ClearLocksWithDurationBeforeDate(expiryTime)
	if clearErr != nil {
		return clearErr
	}
	return nil
}
