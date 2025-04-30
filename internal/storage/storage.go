package storage

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrCantCreateUser     = errors.New("can't create user")
	ErrCantGetUser        = errors.New("can't get user")
	ErrCantCreateOrder    = errors.New("can't create order")
)
