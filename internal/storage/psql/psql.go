package psqlstorage

import (
	"database/sql"

	"github.com/BeInBloom/hide_in_bush/internal/models"
)

type PqsqlStorage struct {
	db *sql.DB
}

func New(s string) *PqsqlStorage {
	// db, err := sql.Open("postgres", s)
	// if err != nil {
	// 	// panic(err)
	// }

	// if err := db.Ping(); err != nil {
	// 	// panic(err)
	// }

	return &PqsqlStorage{
		db: nil,
	}
}

func (p *PqsqlStorage) Close() error {
	return p.db.Close()
}

func (p *PqsqlStorage) CreateOrder(userID string, order models.Order) (string, error) {
	panic("not implemented")
}

func (p *PqsqlStorage) GetUserByID(userID string) (models.User, error) {
	panic("not implemented")
}

func (p *PqsqlStorage) CreateUser(user models.User) (string, error) {
	panic("not implemented")
}

func (p *PqsqlStorage) GetUserByLogin(login string) (models.User, error) {
	panic("not implemented")
}
