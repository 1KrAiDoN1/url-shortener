package storage

import (
	"context"
	"errors"
	"url-shortener/internal/storage/postgres"
)

type Storage struct {
	Postgres postgres.PostgresStorageInterface
}

func NewStorage(pool *postgres.StoragePool) *Storage {
	return &Storage{
		Postgres: pool,
	}
}

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=StorageInterface
type StorageInterface interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURl(ctx context.Context, alias string) error
	URLExists(ctx context.Context, url string) (bool, error)
}

func (s *Storage) URLExists(ctx context.Context, url string) (bool, error) {
	return s.Postgres.URLExists(ctx, url)
}
func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	return s.Postgres.SaveURL(ctx, urlToSave, alias)
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	return s.Postgres.GetURL(ctx, alias)
}

func (s *Storage) DeleteURl(ctx context.Context, alias string) error {
	return s.Postgres.DeleteURl(ctx, alias)
}
