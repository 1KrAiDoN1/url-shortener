package redirect_test

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

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestRedirectHandler(t *testing.T) {
	// Генерируем мок для storage
	// mockery --name=PostgresStorageInterface --dir=internal/storage/postgres --output=internal/storage/mocks --outpkg=mocks
	storageMock := mocks.NewPostgresStorageInterface(t)
	log := slogdiscard.NewDiscardLogger()
	ctx := context.Background()

	tests := []struct {
		name         string
		alias        string
		mockBehavior func()
		expectedCode int
		expectedURL  string
		expectedResp *response.Response
	}{
		{
			name:  "Success",
			alias: "test-alias",
			mockBehavior: func() {
				storageMock.On("GetURL", ctx, "test-alias").
					Return("https://google.com", nil).
					Once()
			},
			expectedCode: http.StatusFound,
			expectedURL:  "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			mockBehavior: func() {
				// Нет вызовов к storage при пустом алиасе
			},
			expectedCode: http.StatusOK,
			expectedResp: &response.Response{
				Status: "Error",
				Error:  "invalid request",
			},
		},
		{
			name:  "URL not found",
			alias: "not-found",
			mockBehavior: func() {
				storageMock.On("GetURL", ctx, "not-found").
					Return("", storage.ErrURLNotFound).
					Once()
			},
			expectedCode: http.StatusOK,
			expectedResp: &response.Response{
				Status: "Error",
				Error:  "not found",
			},
		},
		{
			name:  "Internal error",
			alias: "error-case",
			mockBehavior: func() {
				storageMock.On("GetURL", ctx, "error-case").
					Return("", errors.New("some db error")).
					Once()
			},
			expectedCode: http.StatusOK,
			expectedResp: &response.Response{
				Status: "Error",
				Error:  "internal error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.mockBehavior()

			// Создаем хендлер
			handler := redirect.New(ctx, log, &storage.Storage{Postgres: storageMock})

			// Создаем запрос
			req, err := http.NewRequest("GET", "/"+tt.alias, nil)
			require.NoError(t, err)

			// Добавляем alias в контекст роутера
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

			// Для успешного редиректа проверяем Location
			if tt.expectedCode == http.StatusFound {
				require.Equal(t, tt.expectedURL, rr.Header().Get("Location"))
			} else {
				// Для ошибок проверяем JSON ответ
				var resp response.Response
				err = json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp.Status, resp.Status)
				require.Equal(t, tt.expectedResp.Error, resp.Error)
			}

			// Проверяем что все ожидания по моку выполнены
			storageMock.AssertExpectations(t)
		})
	}
}
