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
		recordCleaner   cleaner
		settings        DispatcherSettings
		errChan         chan error
		doneChan        chan struct{}
		expError        error
	}{
		"Should execute processor, unlocker, and cleaner successfully": {
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
			recordCleaner: func() *mockRecordCleaner {
				mp := mockRecordCleaner{}
				mp.On("RemoveExpiredMessages").Return(nil)
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:           1,
				LockCheckerInterval:       1,
				CleanupWorkerInterval:     1,
				MaxLockTimeDuration:       1 * time.Minute,
				MessagesRetentionDuration: 10 * time.Minute,
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
				mp.On("ProcessRecords").Return(errors.New("process error"))
				return &mp
			}(),
			recordUnlocker: func() *mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("UnlockExpiredMessages").Return(nil)
				return &mp
			}(),
			recordCleaner: func() *mockRecordCleaner {
				mp := mockRecordCleaner{}
				mp.On("RemoveExpiredMessages").Return(nil)
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:           1,
				LockCheckerInterval:       1,
				CleanupWorkerInterval:     1,
				MaxLockTimeDuration:       1,
				MessagesRetentionDuration: 10 * time.Minute,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: errors.New("process error"),
		},
		"Error in unlock records should return error": {
			recordProcessor: func() *mockRecordProcessor {
				mp := mockRecordProcessor{}
				mp.On("ProcessRecords").Return(nil)
				return &mp
			}(),
			recordUnlocker: func() *mockRecordUnlocker {
				mp := mockRecordUnlocker{}
				mp.On("UnlockExpiredMessages").Return(errors.New("unlocker error"))
				return &mp
			}(),
			recordCleaner: func() *mockRecordCleaner {
				mp := mockRecordCleaner{}
				mp.On("RemoveExpiredMessages").Return(nil)
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:           1,
				LockCheckerInterval:       1,
				CleanupWorkerInterval:     1,
				MaxLockTimeDuration:       1,
				MessagesRetentionDuration: 10 * time.Minute,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: errors.New("unlocker error"),
		},
		"Error in clean records should return error": {
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
			recordCleaner: func() *mockRecordCleaner {
				mp := mockRecordCleaner{}
				mp.On("RemoveExpiredMessages").Return(errors.New("cleaner error"))
				return &mp
			}(),
			settings: DispatcherSettings{
				ProcessInterval:           1,
				LockCheckerInterval:       1,
				CleanupWorkerInterval:     1,
				MaxLockTimeDuration:       1,
				MessagesRetentionDuration: 10 * time.Minute,
				RetrialPolicy: RetrialPolicy{
					MaxSendAttemptsEnabled: true,
					MaxSendAttempts:        12,
				},
			},
			errChan:  make(chan error),
			doneChan: make(chan struct{}),
			expError: errors.New("cleaner error"),
		},
	}

	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			d := Dispatcher{
				recordProcessor: tt.recordProcessor,
				recordUnlocker:  tt.recordUnlocker,
				recordCleaner:   tt.recordCleaner,
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

func TestNewDispatcher(t *testing.T) {

	store := MockStore{}
	broker := MockBroker{}
	settings := DispatcherSettings{}
	machineID := "1"
	expectedDispatcher := &Dispatcher{
		recordProcessor: newProcessor(
			&store,
			&broker,
			machineID,
			RetrialPolicy{},
		),
		recordUnlocker: newRecordUnlocker(
			&store,
			time.Duration(0),
		),
		recordCleaner: newRecordCleaner(
			&store,
			time.Duration(0),
		),
		settings: DispatcherSettings{},
	}

	d := NewDispatcher(&store, &broker, settings, machineID)

	assert.Equal(t, expectedDispatcher, d)

}
