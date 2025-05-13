package withdrawalservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	GetUserByID(userID string) (models.User, error)
}

type WithdrawalService struct {
	client http.Client
	url    string
	repo   repo
}

func New(url string, repo repo) *WithdrawalService {
	client := http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    maxIdleConns,
			IdleConnTimeout: idleConnTimeout,
		},
	}

	return &WithdrawalService{
		client: client,
		url:    "http://" + url,
		repo:   repo,
	}
}

func (w *WithdrawalService) GetUserWithdrawals(
	userID string,
) ([]models.Withdrawal, error) {
	user, err := w.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	withdrawal, err := w.getWithdrawalByOrders(user.Orders)
	if err != nil {
		return nil, ErrFailedToGetWithdrawals
	}

	return withdrawal, nil
}

func (w *WithdrawalService) PostWithdraw(
	withdrawwal models.Withdrawal,
) error {
	return nil
}

func (w *WithdrawalService) getWithdrawalByOrders(
	orders []models.Order,
) ([]models.Withdrawal, error) {
	g := errgroup.Group{}
	withdrawals := make([]models.Withdrawal, 0, len(orders))
	var mu sync.Mutex

	sem := make(chan struct{}, queryLimit)

	for _, order := range orders {
		order := order
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			withdrawal, err := w.getWithdrawalByOrderID(order.ID)
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
	orderID string,
) (models.Withdrawal, error) {
	var counter int
	for counter < maxTryCount {
		request, err := w.makeReqByOrderID(orderID)
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
			if err := w.handleToManyRequests(resp); err != nil {
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
	orderID string,
) (*http.Request, error) {
	const withdrawalPath = "/api/orders/"
	qery, err := url.JoinPath(w.url, withdrawalPath, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}

	request, err := http.NewRequest(
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

	time.Sleep(time.Duration(retryAfterInt) * time.Second)
	return nil
}
