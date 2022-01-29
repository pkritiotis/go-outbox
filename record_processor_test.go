package outbox

import (
	"errors"
	"github.com/google/uuid"
	time2 "github.com/pkritiotis/outbox/internal/time"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDefaultRecordProcessor_newProcessor(t *testing.T) {
	p := newProcessor(&MockStore{}, &MockBroker{}, "1", 2)
	assert.NotNil(t, p)
	assert.Equal(t, &MockStore{}, p.store)
	assert.Equal(t, &MockBroker{}, p.messageBroker)
	assert.Equal(t, "1", p.machineID)
	assert.Equal(t, 2, p.MaxSendAttempts)
}

func Test_defaultRecordProcessor_ProcessRecords(t *testing.T) {
	sampleTime := time.Now().UTC()
	timeProvider := &time2.MockProvider{}
	timeProvider.On("Now").Return(sampleTime)

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
		"Eligible records should be processed correctly": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return((*BrokerError)(nil))
				return &mp
			}(),
			store: func() *MockStore {
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
				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          nil,
		},
		"No eligible records should not return an error": {
			messageBroker: &MockBroker{},
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("UpdateRecordLockByState", machineID, sampleTime, PendingDelivery).Return(nil)
				recordsToReturn := []Record{}
				mp.On("GetRecordsByLockID", machineID).Return(recordsToReturn, nil)
				mp.On("ClearLocksByLockID", machineID).Return(nil)
				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          nil,
		},
		"Error in unlocking should return an error": {
			messageBroker: &MockBroker{},
			store: func() *MockStore {
				mp := MockStore{}
				mp.On("UpdateRecordLockByState", machineID, sampleTime, PendingDelivery).
					Return(errors.New("lock error"))
				mp.On("ClearLocksByLockID", machineID).Return(nil)

				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          errors.New("lock error"),
		},
		"Error in record fetching should return an error": {
			messageBroker: &MockBroker{},
			store: func() *MockStore {
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
				mp.On("GetRecordsByLockID", machineID).
					Return(recordsToReturn, errors.New("get error"))
				mp.On("ClearLocksByLockID", machineID).Return(nil)

				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          errors.New("get error"),
		},
		"Error in Update should return an error": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return((*BrokerError)(nil))
				return &mp
			}(),
			store: func() *MockStore {
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
				mp.On("UpdateRecordByID", recordToStore).
					Return(errors.New("update error"))
				mp.On("ClearLocksByLockID", machineID).Return(nil)

				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          errors.New("update error"),
		},
		"Error in Clear locks should not return an error": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return((*BrokerError)(nil))
				return &mp
			}(),
			store: func() *MockStore {
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
				mp.On("ClearLocksByLockID", machineID).
					Return(errors.New("clear locks error"))
				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          nil,
		},
		"Error in broker send should return an error": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return(&BrokerError{
					Error: errors.New("broker unavailable"),
					Type:  BrokerUnavailable,
				})
				return &mp
			}(),
			store: func() *MockStore {
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
				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          errors.New("broker unavailable"),
		},
		"Non availability error in broker should update the record retrial and set the message to pending": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return(&BrokerError{
					Error: errors.New("message broker error"),
					Type:  Other,
				})
				return &mp
			}(),
			store: func() *MockStore {
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
				recordToStore.State = PendingDelivery
				recordToStore.LastAttemptOn = &sampleTime
				errorMsg := errors.New("message broker error").Error()
				recordToStore.Error = &errorMsg
				recordToStore.LockID = nil
				recordToStore.NumberOfAttempts++
				mp.On("UpdateRecordByID", recordToStore).Return(nil)
				mp.On("ClearLocksByLockID", machineID).Return(nil)
				return &mp
			}(),
			machineID:       machineID,
			MaxSendAttempts: 3,
			expErr:          nil,
		},
		"Non availability error in broker should update the record retrial and set the message to max retrials reached": {
			messageBroker: func() *MockBroker {
				mp := MockBroker{}
				mp.On("Send", sampleMessage).Return(&BrokerError{
					Error: errors.New("message broker error"),
					Type:  Other,
				})
				return &mp
			}(),
			store: func() *MockStore {
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
						NumberOfAttempts: 2,
						LastAttemptOn:    nil,
						Error:            nil,
					},
				}
				mp.On("GetRecordsByLockID", machineID).Return(recordsToReturn, nil)
				recordToStore := recordsToReturn[0]
				recordToStore.State = MaxAttemptsReached
				recordToStore.LastAttemptOn = &sampleTime
				errorMsg := errors.New("message broker error").Error()
				recordToStore.Error = &errorMsg
				recordToStore.LockID = nil
				recordToStore.NumberOfAttempts++
				mp.On("UpdateRecordByID", recordToStore).Return(nil)
				mp.On("ClearLocksByLockID", machineID).Return(nil)
				return &mp
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
			err := d.ProcessRecords()
			assert.Equal(t, tt.expErr, err)
		})
	}
}
