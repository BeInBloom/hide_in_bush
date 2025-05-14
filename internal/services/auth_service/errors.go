package authservice

import "errors"

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token expired")
	ErrTokenNotActive = errors.New("token not active")
)
