package handlers

import (
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"net/http"
)

type Handlers struct {
	s    store.Store
	c    *server_config.ServerConfig
	l    *logger.ZapLog
	auth *auth.Authorizer
}

func NewHandlers(
	s store.Store,
	c *server_config.ServerConfig,
	l *logger.ZapLog,
	auth *auth.Authorizer,
) (
	*Handlers,
	error) {
	return &Handlers{
		s:    s,
		c:    c,
		l:    l,
		auth: auth,
	}, nil
}

func (handlers *Handlers) GetUserID(r *http.Request) (userID int, isAuth bool, err error) {
	handlers.l.ZL.Debug("GetUserID started.. ")
	cookie, err := r.Cookie("token")
	if err != nil {
		return 0, false, nil //nolint:nilerr // нужно будет исправить логику
	}
	userID, err = handlers.auth.GetUserID(cookie.Value)
	if err != nil {
		return 0, false, nil //nolint:nilerr // нужно будет исправить логику
	}
	return userID, true, nil
}
