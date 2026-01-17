package withdrawalservice

import "errors"

var (
	ErrWithdrawalNotFound     = errors.New("withdrawal not found")
	ErrFailedToGetWithdrawals = errors.New("failed to get withdrawals")
)
