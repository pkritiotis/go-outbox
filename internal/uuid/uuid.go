package uuid

import "github.com/google/uuid"

// Provider abstracts the UUID generation
type Provider interface {
	NewUUID() uuid.UUID
}

// NewUUIDProvider constructor that returns default UUID generation
func NewUUIDProvider() Provider {
	return uuidProvider{}
}

type uuidProvider struct{}

// NewUUID generates a new UUID
func (u uuidProvider) NewUUID() uuid.UUID {
	return uuid.New()
}
