package orderservice

import "errors"

var (
	ErrOrderBelongsToUser        = errors.New("order belongs to user")
	ErrOrderBelongsToAnotherUser = errors.New("order belongs to another user")
)
