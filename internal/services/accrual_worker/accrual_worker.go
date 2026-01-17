package accrualworker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
)

const (
	pollInterval = 5 * time.Second
	batchSize    = 100
	httpTimeout  = 5 * time.Second
)

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
)

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type repo interface {
	GetPendingOrders(ctx context.Context, limit int) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string, accrual float64) error
	AddAccrualToBalance(ctx context.Context, userID string, amount float64) error
}

type AccrualWorker struct {
	repo       repo
	accrualURL string
	client     http.Client
	logger     *slog.Logger
	done       chan struct{}
}

func New(accrualURL string, repo repo, logger *slog.Logger) *AccrualWorker {
	if !strings.HasPrefix(accrualURL, "http://") && !strings.HasPrefix(accrualURL, "https://") {
		accrualURL = "http://" + accrualURL
	}

	return &AccrualWorker{
		repo:       repo,
		accrualURL: accrualURL,
		client: http.Client{
			Timeout: httpTimeout,
		},
		logger: logger.With("component", "accrual_worker"),
		done:   make(chan struct{}),
	}
}

func (w *AccrualWorker) Run() error {
	w.logger.Info("Запуск accrual worker")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			w.logger.Info("Accrual worker остановлен")
			return nil
		case <-ticker.C:
			w.processOrders(context.Background())
		}
	}
}

func (w *AccrualWorker) Close() error {
	w.logger.Info("Остановка accrual worker")
	close(w.done)
	return nil
}

func (w *AccrualWorker) processOrders(ctx context.Context) {
	orders, err := w.repo.GetPendingOrders(ctx, batchSize)
	if err != nil {
		w.logger.Error("Ошибка получения заказов", "error", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	w.logger.Debug("Обрабатываем заказы", "count", len(orders))

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := w.processOrder(ctx, order); err != nil {
			w.logger.Error("Ошибка обработки заказа",
				"order_id", order.ID,
				"error", err,
			)
		}
	}
}

func (w *AccrualWorker) processOrder(ctx context.Context, order models.Order) error {
	resp, err := w.fetchAccrual(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("ошибка запроса к accrual: %w", err)
	}

	if resp == nil {
		return nil
	}

	if err := w.repo.UpdateOrderStatus(ctx, order.ID, resp.Status, resp.Accrual); err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	if resp.Status == StatusProcessed && resp.Accrual > 0 {
		if err := w.repo.AddAccrualToBalance(ctx, order.UserID, resp.Accrual); err != nil {
			return fmt.Errorf("ошибка начисления баллов: %w", err)
		}
		w.logger.Info("Начислены баллы",
			"order_id", order.ID,
			"user_id", order.UserID,
			"accrual", resp.Accrual,
		)
	}

	return nil
}

func (w *AccrualWorker) fetchAccrual(ctx context.Context, orderID string) (*accrualResponse, error) {
	reqURL, err := url.JoinPath(w.accrualURL, "/api/orders/", orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка формирования URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				w.logger.Warn("Rate limit, ожидание",
					"order_id", orderID,
					"retry_after", seconds,
				)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(seconds) * time.Second):
				}
				return w.fetchAccrual(ctx, orderID)
			}
		}
		return nil, fmt.Errorf("rate limit exceeded")

	default:
		return nil, fmt.Errorf("неожиданный код ответа: %d", resp.StatusCode)
	}
}
