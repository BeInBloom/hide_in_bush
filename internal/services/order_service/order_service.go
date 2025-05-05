package orderservice

import (
	"errors"
	"fmt"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
)

type repo interface {
	GetUserByID(userID string) (models.User, error)
	CreateOrder(userID string, order models.Order) (string, error)
}

type OrderService struct {
	repo repo
}

func New(repo repo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (o *OrderService) UploadOrder(userID string, order models.Order) error {
	_, err := o.repo.CreateOrder(userID, order)
	if err != nil {
		if errors.Is(err, storage.ErrOrderAlreadyExists) {
			return fmt.Errorf("order already exists: %w", err)
		}

		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (o *OrderService) GetUserOrders(userID string) ([]models.Order, error) {
	user, err := o.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.Orders, nil
}
