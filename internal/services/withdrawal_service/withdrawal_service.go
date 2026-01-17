package withdrawalservice

import (
	"context"
	"errors"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type repo interface {
	CreateWithdrawal(ctx context.Context, userID string, withdrawal models.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, userID string) ([]models.Withdrawal, error)
}

type WithdrawalService struct {
	repo repo
}

func New(repo repo) *WithdrawalService {
	return &WithdrawalService{
		repo: repo,
	}
}

func (w *WithdrawalService) GetUserWithdrawals(
	ctx context.Context,
	userID string,
) ([]models.Withdrawal, error) {
	return w.repo.GetUserWithdrawals(ctx, userID)
}

func (w *WithdrawalService) PostWithdraw(
	ctx context.Context,
	userID string,
	withdrawal models.Withdrawal,
) error {
	err := w.repo.CreateWithdrawal(ctx, userID, withdrawal)
	if err != nil {
		if errors.Is(err, storage.ErrInsufficientBalance) {
			return ErrInsufficientBalance
		}
		return err
	}
	return nil
}
