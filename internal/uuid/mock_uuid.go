// Package uuid provides a UUID mock provider
package uuid

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProvider mocks UUID provider
type MockProvider struct {
	mock.Mock
}

// NewUUID returns the mocked UUID
func (m *MockProvider) NewUUID() uuid.UUID {
	args := m.Called()
	return args.Get(0).(uuid.UUID)
}
