package withdrawalservice

import (
	"context"
	"testing"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/services/withdrawal_service/mocks"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"go.uber.org/mock/gomock"
)

func TestGetUserWithdrawals_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	now := time.Now()
	mockRepo.EXPECT().
		GetUserWithdrawals(gomock.Any(), "user-1").
		Return([]models.Withdrawal{
			{Order: "123", Sum: 500, ProcessedAt: now},
		}, nil)

	withdrawals, err := service.GetUserWithdrawals(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUserWithdrawals() error = %v", err)
	}

	if len(withdrawals) != 1 {
		t.Errorf("GetUserWithdrawals() len = %d, want 1", len(withdrawals))
	}
}

func TestPostWithdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateWithdrawal(gomock.Any(), "user-1", gomock.Any()).
		Return(nil)

	err := service.PostWithdraw(context.Background(), "user-1", models.Withdrawal{Order: "123", Sum: 100})
	if err != nil {
		t.Errorf("PostWithdraw() error = %v", err)
	}
}

func TestPostWithdraw_InsufficientBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateWithdrawal(gomock.Any(), "user-1", gomock.Any()).
		Return(storage.ErrInsufficientBalance)

	err := service.PostWithdraw(context.Background(), "user-1", models.Withdrawal{Order: "123", Sum: 1000})
	if err != ErrInsufficientBalance {
		t.Errorf("PostWithdraw() error = %v, want %v", err, ErrInsufficientBalance)
	}
}
