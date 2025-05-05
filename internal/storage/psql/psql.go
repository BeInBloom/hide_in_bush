package psqlstorage

import (
	"database/sql"
	"errors"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PqsqlStorage struct {
	db *sql.DB
}

func New(s string) *PqsqlStorage {
	db, err := createDB(s)
	if err != nil {
		panic(err)
	}

	return &PqsqlStorage{
		db: db,
	}
}

func (p *PqsqlStorage) Close() error {
	return p.db.Close()
}

func (p *PqsqlStorage) CreateOrder(userID string, order models.Order) (string, error) {
	query := `
	INSERT INTO orders (user_id, status, price, uploaded)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	var orderID string
	err := p.db.QueryRow(
		query, userID, order.Status, order.Accrual, order.Uploaded,
	).Scan(&orderID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrOrderAlreadyExists
		}
		return "", storage.ErrCantCreateOrder
	}

	return orderID, nil
}

func (p *PqsqlStorage) GetUserByID(userID string) (models.User, error) {
	user, err := p.getUserByID(userID)
	if err != nil {
		return models.User{}, err
	}

	orders, err := p.GetOrdersByUserID(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Orders = orders

	balance, err := p.GetUserBalance(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Balance = balance

	return user, nil
}

func (p *PqsqlStorage) CreateUser(user models.User) (string, error) {
	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	var userID string
	err := p.db.QueryRow(query, user.Login, user.Password).Scan(&userID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrUserAlreadyExists
		}
		return "", storage.ErrCantCreateUser
	}

	return userID, nil
}

func (p *PqsqlStorage) GetUserByLogin(login string) (models.User, error) {
	user, err := p.getUserByLogin(login)
	if err != nil {
		return models.User{}, err
	}

	orders, err := p.GetOrdersByUserID(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Orders = orders

	balance, err := p.GetUserBalance(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Balance = balance

	return user, nil
}

func (p *PqsqlStorage) getUserByLogin(login string) (models.User, error) {
	userQuery := `
	SELECT id, login, password, created_at, updated_at
	FROM users
	WHERE login = $1`

	var user models.User
	err := p.db.QueryRow(userQuery, login).Scan(
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

func (p *PqsqlStorage) GetUserBalance(userID string) (models.Balance, error) {
	balanceQuery := `
	SELECT user_id, current_balance, withdrawn
	FROM balances
	WHERE user_id = $1`

	var balance models.Balance
	err := p.db.QueryRow(balanceQuery, userID).Scan(
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

func (p *PqsqlStorage) GetOrdersByUserID(userID string) ([]models.Order, error) {
	orderQuery := `
	SELECT id, user_id, status, price, uploaded
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded DESC`

	rows, err := p.db.Query(orderQuery, userID)
	if err != nil {
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

func (p *PqsqlStorage) getUserByID(userID string) (models.User, error) {
	userQuery := `
	SELECT id, login, password, created_at, updated_at
	FROM users
	WHERE id = $1`

	var user models.User
	err := p.db.QueryRow(userQuery, userID).Scan(
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

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
