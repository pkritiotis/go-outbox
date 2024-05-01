package outbox

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	time2 "github.com/pkritiotis/outbox/internal/time"
	uuid2 "github.com/pkritiotis/outbox/internal/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	store := &MockStore{}
	expectedOutbox := Publisher{
		store: store,
		time:  time2.NewTimeProvider(),
		uuid:  uuid2.NewUUIDProvider(),
	}
	got := NewPublisher(store)
	assert.Equal(t, got, expectedOutbox)
}

func TestOutbox_Add(t *testing.T) {
	sampleTx := sql.Tx{}
	sampleUUID, _ := uuid.NewUUID()
	sampleTime := time.Now()

	timeProvider := &time2.MockProvider{}
	timeProvider.On("Now").Return(sampleTime)

	uuidProvider := &uuid2.MockProvider{}
	uuidProvider.On("NewUUID").Return(sampleUUID)

	sampleMessage := Message{
		Key: "testKey",
		Headers: map[string]string{
			"testHeader": "testValue",
		},
		Body:  []byte("testvalue"),
		Topic: "testTopic",
	}
	tests := map[string]struct {
		msg    Message
		store  Store
		tx     *sql.Tx
		expErr error
	}{
		"Successful Send Record should return without error": {
			msg: sampleMessage,
			store: func() *MockStore {
				mp := MockStore{}
				or := Record{
					ID:               sampleUUID,
					Message:          sampleMessage,
					State:            PendingDelivery,
					CreatedOn:        sampleTime.UTC(),
					LockID:           nil,
					LockedOn:         nil,
					ProcessedOn:      nil,
					NumberOfAttempts: 0,
					LastAttemptOn:    nil,
					Error:            nil,
				}
				mp.On("AddRecordTx", or, &sampleTx).Return(nil)
				return &mp
			}(),
			tx:     &sampleTx,
			expErr: nil,
		},
		"Failure in Send Record should return error": {
			msg: sampleMessage,
			store: func() *MockStore {
				mp := MockStore{}
				or := Record{
					ID:               sampleUUID,
					Message:          sampleMessage,
					State:            PendingDelivery,
					CreatedOn:        sampleTime.UTC(),
					LockID:           nil,
					LockedOn:         nil,
					ProcessedOn:      nil,
					NumberOfAttempts: 0,
					LastAttemptOn:    nil,
					Error:            nil,
				}
				mp.On("AddRecordTx", or, &sampleTx).Return(errors.New("error"))
				return &mp
			}(),
			tx:     &sampleTx,
			expErr: errors.New("error"),
		},
	}
	for name, test := range tests {
		tt := test
		t.Run(name, func(t *testing.T) {
			s := Publisher{
				store: tt.store,
				time:  timeProvider,
				uuid:  uuidProvider,
			}
			err := s.Send(tt.msg, tt.tx)
			assert.Equal(t, tt.expErr, err)
		})
	}
}
