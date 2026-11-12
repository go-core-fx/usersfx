package usersfx

import (
	"fmt"
	"time"
)

type Status string
type Provider string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"

	ProviderTelegram Provider = "telegram"
)

type User struct {
	ID     string
	Status Status

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Identity struct {
	Provider     Provider
	ProviderID   string
	ProviderData string
}

func (i Identity) Validate() error {
	if i.ProviderID == "" {
		return fmt.Errorf("%w: provider id is required", ErrValidationFailed)
	}

	switch i.Provider {
	case ProviderTelegram:
	default:
		return fmt.Errorf("%w: invalid provider", ErrValidationFailed)
	}

	return nil
}
