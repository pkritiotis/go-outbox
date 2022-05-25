package outbox

import (
	"errors"
	"github.com/pkritiotis/outbox/internal/time"
	"github.com/stretchr/testify/assert"
	"testing"
	time2 "time"
)

func Test_recordCleaner_removeExpiredMessages(t *testing.T) {
	sampleTime := time2.Now().UTC()
	timeProvider := &time.MockProvider{}
	timeProvider.On("Now").Return(sampleTime)

	tests := map[string]struct {
		store              Store
		time               time.Provider
		MaxMessageLifetime time2.Duration
		expErr             error
	}{
		"Successful removing should not return error": {
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("RemoveRecordsBeforeDatetime", sampleTime.Add(-2*time2.Minute)).Return(nil)
				return &mp
			}(),
			time:               timeProvider,
			MaxMessageLifetime: 2 * time2.Minute,
			expErr:             nil,
		},
		"Error in removing should return error": {
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("RemoveRecordsBeforeDatetime", sampleTime.Add(-2*time2.Minute)).Return(errors.New("test"))
				return &mp
			}(),
			time:               timeProvider,
			MaxMessageLifetime: 2 * time2.Minute,
			expErr:             errors.New("test"),
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			d := recordCleaner{
				store:             tt.store,
				time:              tt.time,
				MaxRecordLifetime: tt.MaxMessageLifetime,
			}
			err := d.RemoveExpiredMessages()
			assert.Equal(t, tt.expErr, err)
		})
	}
}

func Test_newrecordCleaner(t *testing.T) {
	mStore := &MockStore{}
	duration := time2.Duration(1) * time2.Second
	timeProvider := time.NewTimeProvider()
	exprecordCleaner := recordCleaner{
		store:             mStore,
		time:              timeProvider,
		MaxRecordLifetime: duration,
	}

	rc := newRecordCleaner(mStore, duration)

	assert.Equal(t, exprecordCleaner, rc)
}
