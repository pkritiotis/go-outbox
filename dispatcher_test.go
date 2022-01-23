package outbox

import "testing"

func TestDispatcher_Run(t *testing.T) {
	tests := map[string]struct {
		recordProcessor Processor
		recordUnlocker  Unlocker
		settings        DispatcherSettings
		errChan         chan error
		doneChan        chan bool
	}{
		"Should execute processor and unlocker": {
			recordProcessor: func() mockRecordProcessor {
				mp := mockRecordProcessor{}
				mp.On("processRecords").Return(nil)
				return mp
			}(),
			recordUnlocker: func() mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("unlockExpiredMessages").Return(nil)
				return mp
			}(),
			settings: DispatcherSettings{
				ProcessIntervalSeconds:     1,
				LockCheckerIntervalSeconds: 1,
				MaxLockTimeDurationMins:    1,
				MaxSendAttempts:            12,
				TimeBetweenAttemptsSec:     12,
			},
			errChan:  make(chan error),
			doneChan: make(chan bool),
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			d := Dispatcher{
				recordProcessor: tt.recordProcessor,
				recordUnlocker:  tt.recordUnlocker,
				settings:        tt.settings,
			}
			d.Run(tt.errChan, tt.doneChan)
			tt.doneChan <- true
		})
	}
}
