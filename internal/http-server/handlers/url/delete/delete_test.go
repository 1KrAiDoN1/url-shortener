package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

// TestStorageWrapper - обертка для тестирования Storage
type TestStorageWrapper struct {
	*mocks.PostgresStorageInterface
}

func TestDeleteHandler(t *testing.T) {
	log := slogdiscard.NewDiscardLogger()
	ctx := context.Background()

	tests := []struct {
		name          string
		alias         string
		mockBehavior  func(mock *mocks.PostgresStorageInterface)
		expectedCode  int
		expectedError string
	}{
		{
			name:  "Success",
			alias: "test-alias",
			mockBehavior: func(mock *mocks.PostgresStorageInterface) {
				mock.On("DeleteURl", ctx, "test-alias").
					Return(nil).
					Once()
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "Empty alias",
			alias: "",
			mockBehavior: func(mock *mocks.PostgresStorageInterface) {
				// Нет вызовов к storage при пустом алиасе
			},
			expectedCode:  http.StatusOK,
			expectedError: "empty alias",
		},
		{
			name:  "URL not found",
			alias: "not-found",
			mockBehavior: func(mock *mocks.PostgresStorageInterface) {
				mock.On("DeleteURl", ctx, "not-found").
					Return(storage.ErrURLNotFound).
					Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "internal error",
		},
		{
			name:  "Internal error",
			alias: "error-case",
			mockBehavior: func(mock *mocks.PostgresStorageInterface) {
				mock.On("DeleteURl", ctx, "error-case").
					Return(errors.New("some db error")).
					Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage
			storageMock := &mocks.PostgresStorageInterface{}

			// Настраиваем ожидания мока
			tt.mockBehavior(storageMock)

			// Создаем хендлер с моком storage
			handler := delete.New(ctx, log, &storage.Storage{Postgres: storageMock})

			// Создаем тестовый запрос
			req, err := http.NewRequest(http.MethodDelete, "/"+tt.alias, nil)
			require.NoError(t, err)

			// Добавляем параметр маршрута
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tt.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Добавляем request_id в контекст
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id"))

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем хендлер
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			require.Equal(t, tt.expectedCode, rr.Code)

			// Декодируем ответ
			var resp response.Response
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			// Проверяем ответ
			if tt.expectedError == "" {
				require.Equal(t, "OK", resp.Status)
				require.Empty(t, resp.Error)
			} else {
				require.Equal(t, "Error", resp.Status)
				require.Equal(t, tt.expectedError, resp.Error)
			}

			// Проверяем что все ожидания мока выполнены
			storageMock.AssertExpectations(t)
		})
	}
}
