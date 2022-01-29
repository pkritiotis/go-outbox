package outbox

import (
	"errors"
	"github.com/pkritiotis/outbox/internal/time"
	"github.com/stretchr/testify/assert"
	"testing"
	time2 "time"
)

func Test_recordUnlocker_unlockExpiredMessages(t *testing.T) {
	sampleTime := time2.Now().UTC()
	timeProvider := customTimeProvider{sampleTime}

	tests := map[string]struct {
		store                   Store
		time                    time.Provider
		MaxLockTimeDurationMins time2.Duration
		expErr                  error
	}{
		"Successful unlocking should not return error": {
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("ClearLocksWithDurationBeforeDate", sampleTime.Add(-2*time2.Minute)).Return(nil)
				return &mp
			}(),
			time:                    timeProvider,
			MaxLockTimeDurationMins: 2 * time2.Minute,
			expErr:                  nil,
		},
		"Error in unlocking should return error": {
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("ClearLocksWithDurationBeforeDate", sampleTime.Add(-2*time2.Minute)).Return(errors.New("test"))
				return &mp
			}(),
			time:                    timeProvider,
			MaxLockTimeDurationMins: 2 * time2.Minute,
			expErr:                  errors.New("test"),
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			d := recordUnlocker{
				store:                   tt.store,
				time:                    tt.time,
				MaxLockTimeDurationMins: tt.MaxLockTimeDurationMins,
			}
			err := d.UnlockExpiredMessages()
			assert.Equal(t, tt.expErr, err)
		})
	}
}

func Test_newRecordUnlocker(t *testing.T) {
	mStore := &MockStore{}
	duration := time2.Duration(1) * time2.Second
	timeProvider := time.NewTimeProvider()
	expRecordUnlocker := recordUnlocker{
		store:                   mStore,
		time:                    timeProvider,
		MaxLockTimeDurationMins: duration,
	}

	rc := newRecordUnlocker(mStore, duration)

	assert.Equal(t, expRecordUnlocker, rc)
}
