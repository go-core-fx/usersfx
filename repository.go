package usersfx

import (
	"context"
)

type Repository interface {
	RegisterOrLogin(ctx context.Context, ident Identity) (*User, error)
	Login(ctx context.Context, ident Identity) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
}
