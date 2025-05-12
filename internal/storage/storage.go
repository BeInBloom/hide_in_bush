package storage

import "errors"

var (
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrUserNotFound               = errors.New("user not found")
	ErrCantCreateUser             = errors.New("can't create user")
	ErrCantGetUser                = errors.New("can't get user")
	ErrCantCreateOrder            = errors.New("can't create order")
	ErrNoOrders                   = errors.New("no orders")
	ErrCantGetOrders              = errors.New("can't get orders")
	ErrCantGetUserBalance         = errors.New("can't get user balance")
	ErrOrderRegisteredToOtherUser = errors.New("order registered to other user")
	ErrOrderAlreadyRegistered     = errors.New("order already registered")
)
