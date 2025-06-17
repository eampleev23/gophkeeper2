package handlers

import (
	"context"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Общие тестовые переменные.
var (
	testConfig    = server_config.NewServerConfig()
	testLogger, _ = logger.NewZapLogger("info")
	testAuth, _   = auth.Initialize(testConfig, testLogger)
)

// Общая реализация мока хранилища.
type mockStorage struct {
	users map[string]models.User
}

// Конструктор мока хранилища.
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

func (m *mockStorage) GetUserByLogin(ctx context.Context, userLoginReq models.UserLoginReq) (userModelResponse *models.User, err error) {
	return nil, nil
}

func (m *mockStorage) DBConnClose() error {
	return nil
}
