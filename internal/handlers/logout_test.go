package handlers

import (
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_Logout(t *testing.T) {
	// Инициализация тестовых зависимостей
	store := newMockStorage()
	servConf := newMockServerConfig()
	logger, _ := logger.NewZapLogger("info")
	auth, _ := auth.Initialize(servConf, logger)

	// Создание хэндлера
	h, err := NewHandlers(store, servConf, logger, auth)
	require.NoError(t, err)

	t.Run("successful logout", func(t *testing.T) {
		// Создание тестового запроса
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()

		// Установка куки для теста
		cookie := &http.Cookie{
			Name:  "token",
			Value: "test_token",
		}
		req.AddCookie(cookie)

		// Вызов хэндлера
		h.Logout(w, req)

		// Проверки
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверка, что кука очищена
		cookies := resp.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "token", cookies[0].Name)
		assert.Equal(t, "", cookies[0].Value)
		assert.Equal(t, -1, cookies[0].MaxAge)

		// Проверка тела ответа
		var response resultMsg
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.IsError)
		assert.Equal(t, "Logged out successfully", response.ResultMessage)
	})

	t.Run("logout without cookie", func(t *testing.T) {
		// Создание тестового запроса без куки
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()

		// Вызов хэндлера
		h.Logout(w, req)

		// Проверки
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверка тела ответа
		var response resultMsg
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.IsError)
		assert.Equal(t, "Logged out successfully", response.ResultMessage)
	})
}
