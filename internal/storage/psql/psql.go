package psqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PqsqlStorage struct {
	db *sql.DB
}

func New(s string, logger *slog.Logger) *PqsqlStorage {
	db, err := createDB(s)
	if err != nil {
		logger.Error("Не удалось подключиться к базе данных", "error", err)
		os.Exit(1)
	}

	return &PqsqlStorage{
		db: db,
	}
}

func (p *PqsqlStorage) Close() error {
	return p.db.Close()
}

func (p *PqsqlStorage) CreateOrder(ctx context.Context, order models.Order) (string, error) {
	existingOrder, err := p.getOrderByID(ctx, order.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p.createOrderIfNotExists(ctx, order)
		} else {
			return "", fmt.Errorf("failed to get order by ID: %w", err)
		}
	}

	if existingOrder.UserID != order.UserID {
		return "", storage.ErrOrderRegisteredToOtherUser
	}

	return "", storage.ErrOrderAlreadyRegistered
}

func (p *PqsqlStorage) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	user, err := p.getUserByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	orders, err := p.GetOrdersByUserID(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, storage.ErrNoOrders) {
			return models.User{}, err
		}
	}
	user.Orders = orders

	balance, err := p.GetUserBalance(ctx, user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Balance = balance

	return user, nil
}

func (p *PqsqlStorage) CreateUser(ctx context.Context, user models.User) (string, error) {
	query := "INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id"

	now := time.Now()
	var userID string
	err := p.db.QueryRowContext(ctx, query, user.Login, user.Password, now, now).Scan(&userID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrUserAlreadyExists
		}
		return "", storage.ErrCantCreateUser
	}

	return userID, nil
}

func (p *PqsqlStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	user, err := p.getUserByLogin(ctx, login)
	if err != nil {
		return models.User{}, err
	}

	orders, err := p.GetOrdersByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNoOrders) {
			user.Orders = []models.Order{}
		} else {
			return models.User{}, err
		}
	}
	user.Orders = orders

	balance, err := p.GetUserBalance(ctx, user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Balance = balance

	return user, nil
}

func (p *PqsqlStorage) GetUserBalance(ctx context.Context, userID string) (models.Balance, error) {
	balanceQuery := `
	SELECT user_id, current_balance, withdrawn
	FROM balances
	WHERE user_id = $1`

	var balance models.Balance
	err := p.db.QueryRowContext(ctx, balanceQuery, userID).Scan(
		&balance.UserID,
		&balance.CurrentBalance,
		&balance.Withdrawn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Balance{}, storage.ErrUserNotFound
		}

		return models.Balance{}, storage.ErrCantGetUserBalance
	}

	return balance, nil
}

func (p *PqsqlStorage) GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	orderQuery := `
	SELECT id, user_id, status, accrual, uploaded
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded DESC`

	rows, err := p.db.QueryContext(ctx, orderQuery, userID)
	if err != nil {
		if isNoRowsError(err) {
			return nil, storage.ErrNoOrders
		}
		return nil, storage.ErrCantGetOrders
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.Uploaded,
		)
		if err != nil {
			return nil, storage.ErrCantGetOrders
		}

		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, storage.ErrCantGetOrders
	}

	return orders, nil
}

func (p *PqsqlStorage) getOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	query := `
	SELECT id, user_id, status, accrual, uploaded
	FROM orders
	WHERE id = $1`

	var order models.Order
	err := p.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.Uploaded,
	)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return order, nil
}

func (p *PqsqlStorage) getUserByLogin(ctx context.Context, login string) (models.User, error) {
	userQuery := `
	SELECT id, login, password, created_at, updated_at
	FROM users
	WHERE login = $1`

	var user models.User
	err := p.db.QueryRowContext(ctx, userQuery, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, storage.ErrCantGetUser
	}

	return user, nil
}

func (p *PqsqlStorage) getUserByID(ctx context.Context, userID string) (models.User, error) {
	userQuery := `
	SELECT id, login, password, created_at, updated_at
	FROM users
	WHERE id = $1`

	var user models.User
	err := p.db.QueryRowContext(ctx, userQuery, userID).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, storage.ErrCantGetUser
	}

	return user, nil
}

func (p *PqsqlStorage) createOrderIfNotExists(ctx context.Context, order models.Order) (string, error) {
	insertQuery := `
		INSERT INTO orders (id, user_id, status, accrual, uploaded)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var orderID string
	err := p.db.QueryRowContext(
		ctx, insertQuery, order.ID, order.UserID, order.Status, order.Accrual, order.Uploaded,
	).Scan(&orderID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrOrderAlreadyRegistered
		}
		return "", storage.ErrCantCreateOrder
	}

	return orderID, nil
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isNoRowsError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42703"
}

func (p *PqsqlStorage) GetPendingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	query := `
	SELECT id, user_id, status, accrual, uploaded
	FROM orders
	WHERE status IN ('NEW', 'PROCESSING')
	ORDER BY uploaded ASC
	LIMIT $1`

	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.Uploaded,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

func (p *PqsqlStorage) UpdateOrderStatus(ctx context.Context, orderID, status string, accrual float64) error {
	query := `
	UPDATE orders
	SET status = $2, accrual = $3
	WHERE id = $1`

	_, err := p.db.ExecContext(ctx, query, orderID, status, accrual)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (p *PqsqlStorage) AddAccrualToBalance(ctx context.Context, userID string, amount float64) error {
	query := `
	UPDATE balances
	SET current_balance = current_balance + $2
	WHERE user_id = $1`

	_, err := p.db.ExecContext(ctx, query, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to add accrual to balance: %w", err)
	}

	return nil
}

func (p *PqsqlStorage) CreateWithdrawal(ctx context.Context, userID string, withdrawal models.Withdrawal) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRowContext(ctx, `SELECT current_balance FROM balances WHERE user_id = $1`, userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if currentBalance < withdrawal.Sum {
		return storage.ErrInsufficientBalance
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO withdrawals (user_id, order_id, sum, processed_at)
		VALUES ($1, $2, $3, $4)`,
		userID, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE balances 
		SET current_balance = current_balance - $2, withdrawn = withdrawn + $2
		WHERE user_id = $1`,
		userID, withdrawal.Sum)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return tx.Commit()
}

func (p *PqsqlStorage) GetUserWithdrawals(ctx context.Context, userID string) ([]models.Withdrawal, error) {
	query := `
	SELECT order_id, sum, processed_at
	FROM withdrawals
	WHERE user_id = $1
	ORDER BY processed_at DESC`

	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return withdrawals, nil
}
