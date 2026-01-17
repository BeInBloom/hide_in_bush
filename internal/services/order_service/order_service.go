package orderservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
)

type repo interface {
	GetUserByID(ctx context.Context, userID string) (models.User, error)
	CreateOrder(ctx context.Context, order models.Order) (string, error)
}

type OrderService struct {
	repo repo
}

func New(repo repo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (o *OrderService) UploadOrder(ctx context.Context, order models.Order) error {
	_, err := o.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, storage.ErrOrderAlreadyRegistered) {
			return fmt.Errorf("order already exists: %w", err)
		}

		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (o *OrderService) GetUserOrders(ctx context.Context, userID string) ([]models.Order, error) {
	user, err := o.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("user %s not found: %w", userID, err)
		}

		if errors.Is(err, storage.ErrNoOrders) {
			return nil, fmt.Errorf("no orders found for user %s: %w", userID, err)
		}

		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}

	return user.Orders, nil
}
