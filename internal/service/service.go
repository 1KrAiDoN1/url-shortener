package service

import (
	"context"
	"url-shortener/internal/storage"
)

type Service struct {
	storage storage.StorageInterface
}

func NewService(storage storage.StorageInterface) *Service {
	return &Service{
		storage: storage,
	}
}

type ServiceInterface interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURl(ctx context.Context, alias string) error
	URLExists(ctx context.Context, url string) (bool, error)
}

func (s *Service) URLExists(ctx context.Context, url string) (bool, error) {
	return s.storage.URLExists(ctx, url)
}

func (s *Service) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	return s.storage.SaveURL(ctx, urlToSave, alias)
}
func (s *Service) GetURL(ctx context.Context, alias string) (string, error) {
	return s.storage.GetURL(ctx, alias)
}
func (s *Service) DeleteURl(ctx context.Context, alias string) error {
	return s.storage.DeleteURl(ctx, alias)
}
