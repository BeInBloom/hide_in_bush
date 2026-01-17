package orderservice

import (
	"context"
	"testing"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/services/order_service/mocks"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"go.uber.org/mock/gomock"
)

func TestUploadOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return("123", nil)

	err := service.UploadOrder(context.Background(), models.Order{ID: "123"})
	if err != nil {
		t.Errorf("UploadOrder() error = %v", err)
	}
}

func TestUploadOrder_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return("", storage.ErrOrderAlreadyRegistered)

	err := service.UploadOrder(context.Background(), models.Order{ID: "123"})
	if err == nil {
		t.Error("UploadOrder() expected error")
	}
}

func TestGetUserOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	now := time.Now()
	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user-1").
		Return(models.User{
			ID: "user-1",
			Orders: []models.Order{
				{ID: "123", Status: "PROCESSED", Uploaded: now},
			},
		}, nil)

	orders, err := service.GetUserOrders(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUserOrders() error = %v", err)
	}

	if len(orders) != 1 {
		t.Errorf("GetUserOrders() len = %d, want 1", len(orders))
	}
	if orders[0].ID != "123" {
		t.Errorf("GetUserOrders() number = %v, want 123", orders[0].ID)
	}
}

func TestGetUserOrders_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user-unknown").
		Return(models.User{}, storage.ErrUserNotFound)

	_, err := service.GetUserOrders(context.Background(), "user-unknown")
	if err == nil {
		t.Error("GetUserOrders() expected error for unknown user")
	}
}
