package bun

import (
	"time"

	"github.com/go-core-fx/bunfx"
	"github.com/go-core-fx/usersfx"
	"github.com/uptrace/bun"
)

type user struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	bunfx.TimedModel

	ID     string         `bun:"id,pk"`
	Status usersfx.Status `bun:"status"`

	Identities []identity `bun:"rel:has-many,join:id=user_id"`
}

func newUser(id string) *user {
	return &user{
		BaseModel: bun.BaseModel{},
		TimedModel: bunfx.TimedModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},

		ID:     id,
		Status: usersfx.StatusActive,

		Identities: nil,
	}
}

type identity struct {
	bun.BaseModel `bun:"table:user_identities,alias:ui"`
	bunfx.TimedModel

	ID     int64  `bun:"id,pk,autoincrement"`
	UserID string `bun:"user_id"`

	Provider     string `bun:"provider"`
	ProviderID   string `bun:"provider_id"`
	ProviderData string `bun:"provider_data,nullzero"`
}

func newIdentity(userID string, provider string, providerID string, providerData string) *identity {
	return &identity{
		BaseModel: bun.BaseModel{},
		TimedModel: bunfx.TimedModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},

		ID:     0,
		UserID: userID,

		Provider:     provider,
		ProviderID:   providerID,
		ProviderData: providerData,
	}
}

func (u *user) toDomain() *usersfx.User {
	if u == nil {
		return nil
	}

	return &usersfx.User{
		ID:     u.ID,
		Status: u.Status,

		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
