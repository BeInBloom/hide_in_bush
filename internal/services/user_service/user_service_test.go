package userservice

import (
	"context"
	"testing"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/services/user_service/mocks"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return("user-123", nil)

	userID, err := service.Register(context.Background(), models.UserCredentials{
		Login:    "testuser",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if userID != "user-123" {
		t.Errorf("Register() = %v, want %v", userID, "user-123")
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return("", storage.ErrUserAlreadyExists)

	_, err := service.Register(context.Background(), models.UserCredentials{
		Login:    "existinguser",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("Register() expected error")
	}
}

func TestValidateCredentials_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "testuser").
		Return(models.User{
			ID:       "user-123",
			Login:    "testuser",
			Password: string(hashedPassword),
		}, nil)

	userID, err := service.ValidateCredentials(context.Background(), models.UserCredentials{
		Login:    "testuser",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("ValidateCredentials() error = %v", err)
	}

	if userID != "user-123" {
		t.Errorf("ValidateCredentials() = %v, want %v", userID, "user-123")
	}
}

func TestValidateCredentials_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "testuser").
		Return(models.User{
			ID:       "user-123",
			Login:    "testuser",
			Password: string(hashedPassword),
		}, nil)

	_, err := service.ValidateCredentials(context.Background(), models.UserCredentials{
		Login:    "testuser",
		Password: "wrongpassword",
	})

	if err == nil {
		t.Fatal("ValidateCredentials() expected error for wrong password")
	}
}

func TestValidateCredentials_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "nonexistent").
		Return(models.User{}, storage.ErrUserNotFound)

	_, err := service.ValidateCredentials(context.Background(), models.UserCredentials{
		Login:    "nonexistent",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("ValidateCredentials() expected error for non-existent user")
	}
}

func TestUserBalance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	service := New(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user-123").
		Return(models.User{
			ID: "user-123",
			Balance: models.Balance{
				CurrentBalance: 500.5,
				Withdrawn:      42.0,
			},
		}, nil)

	balance, err := service.UserBalance(context.Background(), "user-123")

	if err != nil {
		t.Fatalf("UserBalance() error = %v", err)
	}

	if balance.CurrentBalance != 500.5 {
		t.Errorf("UserBalance().CurrentBalance = %v, want %v", balance.CurrentBalance, 500.5)
	}

	if balance.Withdrawn != 42.0 {
		t.Errorf("UserBalance().Withdrawn = %v, want %v", balance.Withdrawn, 42.0)
	}
}
