package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Выносим создание конфига, логгера, авторизатора в глобальные переменные

var (
	testConfig    = server_config.NewServerConfig()
	testLogger, _ = logger.NewZapLogger("info")
	testAuth, _   = auth.Initialize(testConfig, testLogger)
)

// mockStorage реализует store.Store для тестов
type mockStorage struct {
	users map[string]models.User
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		users: make(map[string]models.User),
	}
}

func (m *mockStorage) CreateUser(ctx context.Context, userReq models.UserRegReq) (*models.User, error) {
	// Для теста BadRequest возвращаем ошибку при пустом логине
	if userReq.Login == "" {
		return nil, &pgconn.PgError{Code: pgerrcode.NotNullViolation}
	}
	if _, exists := m.users[userReq.Login]; exists {
		return &models.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	}
	newUser := models.User{
		ID:    len(m.users) + 1,
		Login: userReq.Login,
	}
	m.users[userReq.Login] = newUser
	return &newUser, nil
}

func (m *mockStorage) DBConnClose() error {
	return nil
}

func TestHandlers_Registration(t *testing.T) {

	t.Parallel() // Разрешаем параллельное выполнение тестов

	type want struct {
		statusCode   int
		jsonResponse resultMsg
	}

	tests := []struct {
		name        string
		requestUrl  string
		requestBody interface{}
		tableUsers  map[string]models.User
		want        want
	}{
		{
			name:       "Test status http.StatusOK",
			requestUrl: "/api/user/registration/",
			requestBody: models.UserRegReq{
				Login:    "Petr",
				Password: "petrPass",
			},
			tableUsers: map[string]models.User{
				"1": {Login: "Alex"},
				"2": {Login: "Drew"},
				"3": {Login: "Valery"},
			},
			want: want{
				statusCode: http.StatusOK,
				jsonResponse: resultMsg{
					IsError:       false,
					ResultMessage: "Success authenticate user after successful registration",
				},
			},
		},
		{
			name:        "Test status http.StatusBadRequest",
			requestUrl:  "/api/user/registration/",
			requestBody: "invalid request body",
			tableUsers: map[string]models.User{
				"1": {Login: "Alex"},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "Not a valid user registration request",
				},
			},
		},
		{
			name:        "Test status http.StatusBadRequest with fake struct",
			requestUrl:  "/api/user/registration/",
			requestBody: map[int]int{1: 1},
			tableUsers: map[string]models.User{
				"1": {Login: "Alex"},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "Login and password are required",
				},
			},
		},
		{
			name:       "Test status http.StatusConflict",
			requestUrl: "/api/user/registration/",
			requestBody: models.UserRegReq{
				Login:    "Petr",
				Password: "petrPass",
			},
			tableUsers: map[string]models.User{
				"1": {Login: "Petr"},
			},
			want: want{
				statusCode: http.StatusConflict,
				jsonResponse: resultMsg{
					IsError:       true,
					ResultMessage: "User with this login already exists",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
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
			for _, user := range tt.tableUsers {
				s.users[user.Login] = user // Заполняем мок данными из тест-кейса
			}

			handlers, err := NewHandlers(s, testConfig, testLogger, testAuth)
			require.NoError(t, err)
			h := http.HandlerFunc(handlers.Registration)
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
