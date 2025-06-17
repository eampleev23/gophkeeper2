package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_Login(t *testing.T) {
	t.Parallel()

	type want struct {
		statusCode   int
		jsonResponse resultMsg
	}

	hashedPassword := "password"

	tests := []struct {
		name        string
		requestUrl  string
		requestBody interface{}
		tableUsers  map[string]models.User
		want        want
	}{
		{
			name:       "Test successful login",
			requestUrl: "/api/user/login/",
			requestBody: models.UserLoginReq{
				Login:    "Petr",
				Password: "correctPassword",
			},
			tableUsers: map[string]models.User{
				"Petr": {
					ID:           1,
					Login:        "Petr",
					PasswordHash: hashedPassword,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				jsonResponse: resultMsg{
					IsError:       false,
					ResultMessage: "Successfully logged in",
				},
			},
		},
		{
			name:        "Test invalid request body",
			requestUrl:  "/api/user/login/",
			requestBody: "invalid request body",
			tableUsers:  map[string]models.User{},
			want: want{
				statusCode: http.StatusBadRequest,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "Not a valid user login request",
				},
			},
		},
		{
			name:        "Test empty login or password",
			requestUrl:  "/api/user/login/",
			requestBody: map[string]string{"login": "", "password": ""},
			tableUsers:  map[string]models.User{},
			want: want{
				statusCode: http.StatusBadRequest,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "Login and password are required",
				},
			},
		},
		{
			name:       "Test user not found",
			requestUrl: "/api/user/login/",
			requestBody: models.UserLoginReq{
				Login:    "NonExistent",
				Password: "anyPassword",
			},
			tableUsers: map[string]models.User{
				"Petr": {
					ID:           1,
					Login:        "Petr",
					PasswordHash: hashedPassword,
				},
			},
			want: want{
				statusCode: http.StatusNotFound,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "User with this login does not exist",
				},
			},
		},
		{
			name:       "Test wrong password",
			requestUrl: "/api/user/login/",
			requestBody: models.UserLoginReq{
				Login:    "Petr",
				Password: "wrongPassword",
			},
			tableUsers: map[string]models.User{
				"Petr": {
					ID:           1,
					Login:        "Petr",
					PasswordHash: hashedPassword,
				},
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "The passwords don't match",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			request := httptest.NewRequest(http.MethodPost, tt.requestUrl, bytes.NewBuffer(jsonBody))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Создаем мок хранилище и заполняем его тестовыми данными
			s := newMockStorage()
			s.users = tt.tableUsers

			// Мокаем функцию VerifyPassword
			originalVerifyPassword := store.VerifyPassword
			defer func() { store.VerifyPassword = originalVerifyPassword }()
			store.VerifyPassword = func(password, hash string) (bool, error) {
				// В реальном приложении здесь должна быть правильная проверка хэша
				if password == "correctPassword" {
					return true, nil
				}
				return false, nil
			}

			handlers, err := NewHandlers(s, testConfig, testLogger, testAuth)
			require.NoError(t, err)
			h := http.HandlerFunc(handlers.Login)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			bytesJsonResponse, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			var jsonResponse resultMsg
			err = json.Unmarshal(bytesJsonResponse, &jsonResponse)
			require.NoError(t, err)

			assert.Equal(t, tt.want.jsonResponse, jsonResponse)
		})
	}
}
