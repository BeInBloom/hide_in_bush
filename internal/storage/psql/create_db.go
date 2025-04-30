package psqlstorage

import (
	"database/sql"
	"fmt"
)

func createDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Инициализация схемы базы данных
	if err := initializeSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return db, nil
}

func initializeSchema(db *sql.DB) error {
	query := `
    -- Таблица пользователей
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
        login TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    -- Таблица заказов
    CREATE TABLE IF NOT EXISTS orders (
        id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
        user_id TEXT NOT NULL,
        status TEXT NOT NULL,
        accrual NUMERIC(10, 2),
        uploaded TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_user
            FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
    );

    -- Таблица баланса
    CREATE TABLE IF NOT EXISTS balances (
        user_id TEXT PRIMARY KEY,
        current_balance NUMERIC(10, 2) NOT NULL DEFAULT 0,
        withdrawn NUMERIC(10, 2) NOT NULL DEFAULT 0,
        CONSTRAINT fk_user_balance
            FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
    );
    
    -- Создание функции для триггера, если её нет
    CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    -- Создание триггера, если его нет
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
            CREATE TRIGGER update_users_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        END IF;
    END
    $$;

    -- Создание индексов, если их нет
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_orders_user_id') THEN
            CREATE INDEX idx_orders_user_id ON orders (user_id);
        END IF;
        
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_orders_status') THEN
            CREATE INDEX idx_orders_status ON orders (status);
        END IF;
    END
    $$;
    `

	_, err := db.Exec(query)
	return err
}
