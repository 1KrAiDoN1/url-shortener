package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoragePool struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*StoragePool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &StoragePool{
			pool: pool},
		nil

}

func (d *StoragePool) GetPool() *pgxpool.Pool {
	return d.pool
}
func (d *StoragePool) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}

type PostgresStorageInterface interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURl(ctx context.Context, alias string) error
	URLExists(ctx context.Context, url string) (bool, error)
}

func (d *StoragePool) URLExists(ctx context.Context, url string) (bool, error) {
	const op = "postgres.storage.AliasExists"
	var exists bool
	err := d.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM url WHERE url = $1)`, url).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check url existence: %w", op, err)
	}
	return exists, nil

}
func (s *StoragePool) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const op = "postgres.storage.SaveURL"
	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO url (url, alias) VALUES ($1, $2) RETURNING id`, urlToSave, alias).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s failed to save url: %w", op, err)
	}
	return id, nil
}

func (s *StoragePool) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "postgres.storage.GetURL"
	var url string
	err := s.pool.QueryRow(ctx, `SELECT url FROM url WHERE alias = $1`, alias).Scan(&url)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%s url not found: %w", op, err)
	}
	if err != nil {
		return "", fmt.Errorf("%s failed to get url: %w", op, err)
	}
	return url, nil
}

func (s *StoragePool) DeleteURl(ctx context.Context, alias string) error {
	const op = "postgres.storage.DeleteURl"
	res, err := s.pool.Exec(ctx, `DELETE FROM url WHERE alias = $1`, alias)
	if err != nil {
		return fmt.Errorf("%s failed to delete url: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("url not found or access denied")
	}

	return nil
}
