package outbox

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDispatcher_Run(t *testing.T) {
	tests := map[string]struct {
		recordProcessor processor
		recordUnlocker  unlocker
		settings        DispatcherSettings
		errChan         chan error
		doneChan        chan struct{}
		expError        error
	}{
		"Should execute processor and unlocker successfully": {
			recordProcessor: func() *mockRecordProcessor {
				mp := mockRecordProcessor{}
				mp.On("ProcessRecords").Return(nil)
				return &mp
			}(),
			recordUnlocker: func() *mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("UnlockExpiredMessages").Return(nil)
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:     1 * time.Minute,
				LockCheckerInterval: 1 * time.Minute,
				MaxLockTimeDuration: 1 * time.Minute,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: nil,
		},
		"Error in process records should return error": {
			recordProcessor: func() *mockRecordProcessor {
				mp := mockRecordProcessor{}
				mp.On("ProcessRecords").Return(errors.New("test"))
				return &mp
			}(),
			recordUnlocker: func() *mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("UnlockExpiredMessages").Return(nil)
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:     1,
				LockCheckerInterval: 1,
				MaxLockTimeDuration: 1,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: errors.New("test"),
		},
		"Error in unlock records should return error": {
			recordProcessor: func() *mockRecordProcessor {
				mp := mockRecordProcessor{}
				mp.On("ProcessRecords").Return(nil)
				return &mp
			}(),
			recordUnlocker: func() *mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("UnlockExpiredMessages").Return(errors.New("test"))
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:     1,
				LockCheckerInterval: 1,
				MaxLockTimeDuration: 1,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: errors.New("test"),
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
			var err error
			if tt.expError != nil {
				err = <-tt.errChan
			}
			tt.doneChan <- struct{}{}

			assert.Equal(t, tt.expError, err)
		})
	}
}
