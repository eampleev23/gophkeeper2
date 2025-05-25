package auth

import (
	"context"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

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

// Auth мидлвар, который проверяет авторизацию.
func (au *Authorizer) Auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			// Получаем логгер из контекста запроса
			logger, ok := r.Context().Value(keyLogger).(*logger.ZapLog)
			if !ok {
				log.Printf("Error getting logger")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			logger.ZL.Debug("No cookie", zap.String("err", err.Error()))
			next.ServeHTTP(w, r)
			return
		}
		// если кука уже установлена, то через контекст передаем 0
		ctx := context.WithValue(r.Context(), KeyUserIDCtx, 0)
		next.ServeHTTP(w, r.WithContext(ctx))
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
	}
	http.SetCookie(w, &cookie)
	return nil
}

// Claims описывает утверждения, хранящиеся в токене + добавляет кастомное UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID    int
	UserLogin string
}

// GetUserID возвращает ID пользователя.
func (au *Authorizer) GetUserID(tokenString string) (int, error) {
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
