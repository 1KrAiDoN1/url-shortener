package save_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestHandlers_New(t *testing.T) {
	type mockBehavior func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error)

	const testAlias = "test_alias"
	const testURL = "https://google.com"

	tests := []struct {
		name          string
		inputBody     string
		mockBehavior  mockBehavior
		expectedCode  int
		checkResponse func(t *testing.T, response save.Response)
	}{
		{
			name:      "Success with alias",
			inputBody: fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, testURL, testAlias),
			mockBehavior: func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error) {
				s.On("URLExists", mock.Anything, url).Return(exists, nil)
				s.On("SaveURL", mock.Anything, url, alias).Return(int64(1), saveErr)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp save.Response) {
				require.Equal(t, int64(1), resp.Id)
				require.Equal(t, testAlias, resp.Alias)
				require.Equal(t, "OK", resp.Status)
			},
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"url": "https://google.com", "alias": "test",}`,
			mockBehavior: func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error) {},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp save.Response) {
				require.Equal(t, "Error", resp.Status)
				require.Contains(t, resp.Error, "failed to decode")
			},
		},
		{
			name:         "Empty URL",
			inputBody:    `{"url": "", "alias": "test"}`,
			mockBehavior: func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error) {},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp save.Response) {
				require.Equal(t, "Error", resp.Status)
				require.Contains(t, resp.Error, "required field")
			},
		},
		{
			name:      "URL already exists",
			inputBody: fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, testURL, testAlias),
			mockBehavior: func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error) {
				s.On("URLExists", mock.Anything, url).Return(true, nil)
			},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp save.Response) {
				require.Equal(t, "Error", resp.Status)
				require.Equal(t, "url already exists", resp.Error)
			},
		},
		{
			name:      "Save URL error",
			inputBody: fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, testURL, testAlias),
			mockBehavior: func(s *mocks.ServiceInterface, url, alias string, exists bool, saveErr error) {
				s.On("URLExists", mock.Anything, url).Return(exists, nil)
				s.On("SaveURL", mock.Anything, url, alias).Return(int64(0), errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, resp save.Response) {
				require.Equal(t, "Error", resp.Status)
				require.Equal(t, "failed to add url", resp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewServiceInterface(t)
			tt.mockBehavior(serviceMock, testURL, testAlias, false, nil)

			h := save.NewHandlers(serviceMock)
			handler := h.New(context.Background(), slogdiscard.NewDiscardLogger())

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewBufferString(tt.inputBody))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)

			var resp save.Response
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			tt.checkResponse(t, resp)
		})
	}
}
