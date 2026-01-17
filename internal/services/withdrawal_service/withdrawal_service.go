package withdrawalservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"golang.org/x/sync/errgroup"
)

const (
	queryLimit      = 100
	maxIdleConns    = 10
	idleConnTimeout = 30 * time.Second
	maxTryCount     = 5
)

type repo interface {
	GetUserByID(ctx context.Context, userID string) (models.User, error)
}

type WithdrawalService struct {
	client http.Client
	url    string
	repo   repo
}

func New(baseURL string, repo repo) *WithdrawalService {
	// Добавляем схему только если её нет
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	client := http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    maxIdleConns,
			IdleConnTimeout: idleConnTimeout,
		},
	}

	return &WithdrawalService{
		client: client,
		url:    baseURL,
		repo:   repo,
	}
}

func (w *WithdrawalService) GetUserWithdrawals(
	ctx context.Context,
	userID string,
) ([]models.Withdrawal, error) {
	user, err := w.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	withdrawal, err := w.getWithdrawalByOrders(ctx, user.Orders)
	if err != nil {
		return nil, ErrFailedToGetWithdrawals
	}

	return withdrawal, nil
}

func (w *WithdrawalService) PostWithdraw(
	ctx context.Context,
	withdrawwal models.Withdrawal,
) error {
	return nil
}

func (w *WithdrawalService) getWithdrawalByOrders(
	ctx context.Context,
	orders []models.Order,
) ([]models.Withdrawal, error) {
	g, ctx := errgroup.WithContext(ctx)
	withdrawals := make([]models.Withdrawal, 0, len(orders))
	var mu sync.Mutex

	sem := make(chan struct{}, queryLimit)

	for _, order := range orders {
		order := order
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()

			withdrawal, err := w.getWithdrawalByOrderID(ctx, order.ID)
			if err != nil {
				if errors.Is(err, ErrWithdrawalNotFound) {
					return nil
				}
				return err
			}

			mu.Lock()
			withdrawals = append(withdrawals, withdrawal)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (w *WithdrawalService) getWithdrawalByOrderID(
	ctx context.Context,
	orderID string,
) (models.Withdrawal, error) {
	var counter int
	for counter < maxTryCount {
		select {
		case <-ctx.Done():
			return models.Withdrawal{}, ctx.Err()
		default:
		}

		request, err := w.makeReqByOrderID(ctx, orderID)
		if err != nil {
			return models.Withdrawal{}, fmt.Errorf("failed to make request: %w", err)
		}

		resp, err := w.client.Do(request)
		if err != nil {
			return models.Withdrawal{}, fmt.Errorf("failed to do request: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			return w.handleStatusOk(resp)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if err := w.handleToManyRequests(ctx, resp); err != nil {
				return models.Withdrawal{}, fmt.Errorf("failed to handle too many requests: %w", err)
			}
			counter++
			continue
		}

		if resp.StatusCode == http.StatusNoContent {
			return models.Withdrawal{}, ErrWithdrawalNotFound
		}

		break
	}

	return models.Withdrawal{}, ErrFailedToGetWithdrawals
}

func (w *WithdrawalService) makeReqByOrderID(
	ctx context.Context,
	orderID string,
) (*http.Request, error) {
	const withdrawalPath = "/api/orders/"
	qery, err := url.JoinPath(w.url, withdrawalPath, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		qery,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "text/plain")

	return request, nil
}

func (w *WithdrawalService) handleStatusOk(
	r *http.Response,
) (models.Withdrawal, error) {
	defer r.Body.Close()
	var withdrawal models.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		return models.Withdrawal{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return withdrawal, nil
}

func (w *WithdrawalService) handleToManyRequests(
	ctx context.Context,
	r *http.Response,
) error {
	retryAfter := r.Header.Get("Retry-After")
	if retryAfter == "" {
		return fmt.Errorf("retry-after header is empty")
	}

	retryAfterInt, err := strconv.Atoi(retryAfter)
	if err != nil {
		return fmt.Errorf("failed to parse retry-after not int: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(retryAfterInt) * time.Second):
	}
	return nil
}
