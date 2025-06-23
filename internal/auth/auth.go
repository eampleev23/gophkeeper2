package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// resultMessage - структура для возврата json ответа более детализированного, чем просто статус.
type resultMsg struct {
	IsError       bool
	ResultMessage string `json:"result_message"`
}

type Authorizer struct {
	logger   *logger.ZapLog
	servConf *server_config.ServerConfig
}

var keyLogger logger.Key = logger.KeyLoggerCtx

// Initialize инициализирует синглтон авторизовывальщика с секретным ключом.
func Initialize(c *server_config.ServerConfig, l *logger.ZapLog) (*Authorizer, error) {
	au := &Authorizer{
		servConf: c,
		logger:   l,
	}
	return au, nil
}

type Key string

const (
	KeyUserIDCtx Key = "user_id_ctx"
)

// MiddleCheckAuth мидлвар, который проверяет авторизацию.
func (au *Authorizer) MiddleCheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, gotRequest *http.Request) {
		// 1. Получаем куку.
		cookie, err := gotRequest.Cookie("token")
		if err != nil {
			au.logger.ZL.Debug("No token cookie", zap.Error(err))
			resultMsg := resultMsg{IsError: true, ResultMessage: "Authentication required"}
			msg, _ := json.Marshal(resultMsg)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			responseWriter.Write(msg)
			return
		}

		// 2. Парсим и валидируем JWT.
		token, err := jwt.ParseWithClaims(
			cookie.Value,
			&Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(au.servConf.SecretKey), nil
			},
		)
		if err != nil || !token.Valid {
			au.logger.ZL.Debug("Invalid token", zap.Error(err))
			resultMsg := resultMsg{IsError: true, ResultMessage: "Invalid token"}
			msg, _ := json.Marshal(resultMsg)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			responseWriter.Write(msg)
			return
		}

		// 3. Извлекаем данные пользователя
		claims, ok := token.Claims.(*Claims)
		if !ok {
			au.logger.ZL.Debug("Failed to parse claims", zap.Error(err))
			resultMsg := resultMsg{IsError: true, ResultMessage: "Invalid token"}
			msg, _ := json.Marshal(resultMsg)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			responseWriter.Write(msg)
			return
		}
		// 4. Передаем userID в контекст.
		ctx := context.WithValue(context.Background(), KeyUserIDCtx, claims.UserID)
		next.ServeHTTP(responseWriter, gotRequest.WithContext(ctx))
	})
}

// MiddleCheckNoAuth мидлвар, который проверяет роуты, к которым должны обращаться не авторизованные пользователи.
func (au *Authorizer) MiddleCheckNoAuth(next http.Handler) http.Handler {
	fn := func(responseWriter http.ResponseWriter, gotRequest *http.Request) {

		_, err := gotRequest.Cookie("token")
		if err != nil {
			next.ServeHTTP(responseWriter, gotRequest.WithContext(gotRequest.Context()))
			return
		}
		resultMsg := resultMsg{IsError: true, ResultMessage: "Already authenticated"}
		msg, _ := json.Marshal(resultMsg)
		responseWriter.WriteHeader(http.StatusForbidden)
		responseWriter.Write(msg)
		return
	}
	return http.HandlerFunc(fn)
}

func (au *Authorizer) SetNewCookie(w http.ResponseWriter, userID int, userLogin string) (err error) {
	au.logger.ZL.Debug("setNewCookie got userID", zap.Int("userID", userID))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Когда создан токен.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(au.servConf.TokenExp)),
		},
		// Собственное утверждение.
		UserID:    userID,
		UserLogin: userLogin,
	})
	tokenString, err := token.SignedString([]byte(au.servConf.SecretKey))
	if err != nil {
		return fmt.Errorf("token.SignedString fail.. %w", err)
	}
	cookie := http.Cookie{
		Name:  "token",
		Value: tokenString,
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}

func (au *Authorizer) Logout(w http.ResponseWriter) {

	au.logger.ZL.Debug("logging out user by clearing cookie")

	// Создаем cookie с таким же именем, но с истекшим сроком действия

	cookie := http.Cookie{
		Name:    "token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Now().Add(-1 * time.Hour), // Устанавливаем время в прошлом
	}
	http.SetCookie(w, &cookie)
}

// Claims описывает утверждения, хранящиеся в токене + добавляет кастомное UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID    int
	UserLogin string
}

// GetUserID возвращает ID пользователя.
func (au *Authorizer) GetUserIDByCookie(tokenString string) (int, error) {
	// Создаем экземпляр структуры с утверждениями
	claims := &Claims{}
	// Парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(au.servConf.SecretKey), nil
	})
	if err != nil {
		au.logger.ZL.Info("Failed in case to get ownerId from token ", zap.Error(err))
		return 0, err
	}
	return claims.UserID, nil
}
