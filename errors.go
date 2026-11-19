package usersfx

import "errors"

var (
	ErrValidationFailed = errors.New("validation failed")
	ErrNotFound         = errors.New("user not found")
	ErrIdentityExists   = errors.New("identity already exists")
)
