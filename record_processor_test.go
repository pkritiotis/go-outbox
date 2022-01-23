package outbox

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_defaultRecordProcessor_processRecords(t *testing.T) {
	sampleTime := time.Now().UTC()
	timeProvider := customTimeProvider{sampleTime}

	sampleMessage := Message{
		Key: "testKey",
		Headers: []MessageHeader{{
			Key:   "testHeader",
			Value: "testValue",
		}},
		Body:  []byte("testvalue"),
		Topic: "testTopic",
	}
	machineID := "1"
	tests := map[string]struct {
		messageBroker   MessageBroker
		store           Store
		machineID       string
		MaxSendAttempts int
		expErr          error
	}{
		"Records should be processed correctly": {
			messageBroker: func() MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return((*BrokerError)(nil))
				return mp
			}(),
			store: func() MockStore {
				mp := MockStore{}
				mp.On("UpdateRecordLockByState", machineID, sampleTime, PendingDelivery).Return(nil)
				recordsToReturn := []Record{
					{
						ID:               uuid.New(),
						Message:          sampleMessage,
						State:            PendingDelivery,
						CreatedOn:        time.Now(),
						LockID:           &machineID,
						LockedOn:         nil,
						ProcessedOn:      nil,
						NumberOfAttempts: 0,
						LastAttemptOn:    nil,
						Error:            nil,
					},
				}
				mp.On("GetRecordsByLockID", machineID).Return(recordsToReturn, nil)
				recordToStore := recordsToReturn[0]
				recordToStore.State = Delivered
				recordToStore.LastAttemptOn = &sampleTime
				recordToStore.LockID = nil
				recordToStore.NumberOfAttempts++
				recordToStore.ProcessedOn = &sampleTime
				mp.On("UpdateRecordByID", recordToStore).Return(nil)
				mp.On("ClearLocksByLockID", machineID).Return(nil)
				return mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          nil,
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			d := defaultRecordProcessor{
				messageBroker:   tt.messageBroker,
				time:            timeProvider,
				store:           tt.store,
				machineID:       tt.machineID,
				MaxSendAttempts: tt.MaxSendAttempts,
			}
			err := d.processRecords()
			assert.Equal(t, tt.expErr, err)
		})
	}
}
