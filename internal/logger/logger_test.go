package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestLogger(t *testing.T) {
	// Создаем тестовый логгер
	zl, err := NewZapLogger("debug")
	assert.NoError(t, err)

	// Создаем тестовый обработчик, который будет проверять наличие логгера в контексте
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что логгер есть в контексте
		loggerFromCtx := r.Context().Value(KeyLoggerCtx)
		assert.NotNil(t, loggerFromCtx)
		assert.IsType(t, &ZapLog{}, loggerFromCtx)

		// Записываем тестовый ответ
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test response"))
		assert.NoError(t, err)
	})

	// Создаем middleware
	middleware := zl.RequestLogger(testHandler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Content-Type", "application/json")

	// Создаем Recorder для записи ответа
	rec := httptest.NewRecorder()

	// Выполняем запрос через middleware
	middleware.ServeHTTP(rec, req)

	// Проверяем, что ответ записан правильно
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test response", rec.Body.String())
}

func TestLoggingResponseWriter(t *testing.T) {
	// Создаем тестовый ResponseWriter
	rec := httptest.NewRecorder()

	// Создаем наш responseData
	respData := &responseData{}

	// Создаем наш loggingResponseWriter
	lrw := loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   respData,
	}

	// Проверяем начальные значения
	assert.Equal(t, 0, respData.status)
	assert.Equal(t, 0, respData.size)

	// Тестируем WriteHeader
	testStatus := http.StatusNotFound
	lrw.WriteHeader(testStatus)
	assert.Equal(t, testStatus, respData.status)
	assert.Equal(t, testStatus, rec.Code)

	// Тестируем Write
	testBody := "test body"
	size, err := lrw.Write([]byte(testBody))
	assert.NoError(t, err)
	assert.Equal(t, len(testBody), size)
	assert.Equal(t, len(testBody), respData.size)
	assert.Equal(t, testBody, rec.Body.String())
}

func TestShortDur(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 5 * time.Second, "5s"},
		{"minutes", 2 * time.Minute, "2m"},
		{"minutes with seconds", 2*time.Minute + 15*time.Second, "2m15s"},
		{"hours", 3 * time.Hour, "3h"},
		{"hours with minutes", 3*time.Hour + 30*time.Minute, "3h30m"},
		{"complex", 3*time.Hour + 30*time.Minute + 15*time.Second, "3h30m15s"},
		{"milliseconds", 500 * time.Millisecond, "500ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, shortDur(tt.duration))
		})
	}
}

func TestInitializeLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{"debug level", "debug", false},
		{"info level", "info", false},
		{"warn level", "warn", false},
		{"error level", "error", false},
		{"invalid level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl := &ZapLog{ZL: zap.NewNop()}
			_, err := Initialize(tt.level, zl)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, zap.NewNop(), zl.ZL)
			}
		})
	}
}

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{"valid level", "debug", false},
		{"invalid level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zl, err := NewZapLogger(tt.level)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, zl)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, zl)
				assert.NotEqual(t, zap.NewNop(), zl.ZL)
			}
		})
	}
}
