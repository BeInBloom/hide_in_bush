package psqlstorage

import (
	"database/sql"
	"errors"

	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	sq "github.com/Masterminds/squirrel"
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
	query := sq.Insert("orders").
		Columns("user_id", "price", "status").
		Values(userID, order.Accrual, order.Status).
		Suffix("RETURNING id")

	var orderID string
	err := query.RunWith(p.db).QueryRow().Scan(&orderID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrOrderAlreadyExists
		}
		return "", storage.ErrCantCreateOrder
	}

	return orderID, nil
}

func (p *PqsqlStorage) GetUserByID(userID string) (models.User, error) {
	query := sq.Select("login", "password").
		From("users").
		Where(sq.Eq{"id": userID})

	var user models.User
	err := query.RunWith(p.db).
		QueryRow().
		Scan(&user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, storage.ErrCantGetUser
	}

	return user, nil
}

func (p *PqsqlStorage) CreateUser(user models.User) (string, error) {
	query := sq.Insert("users").
		Columns("login", "password").
		Values(user.Login, user.Password).
		Suffix("RETURNING id")

	var userID string
	err := query.RunWith(p.db).QueryRow().Scan(&userID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return "", storage.ErrUserAlreadyExists
		}
		return "", storage.ErrCantCreateUser
	}

	return userID, nil
}

func (p *PqsqlStorage) GetUserByLogin(login string) (models.User, error) {
	query := sq.Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	var user models.User
	err := query.RunWith(p.db).
		QueryRow().
		Scan(&user.ID, &user.Login, &user.Password)
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
