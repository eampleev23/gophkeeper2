package handlers

import (
	"encoding/json"
	"errors"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
)

/*
На вход хэндлер ожидает json такого формата:
{
    "login": "<login>",
    "password": "<password>"
}
*/

type errorMsg struct {
	ErrorMessage string `json:"error_message"`
}

func (handlers *Handlers) Registration(responseWriter http.ResponseWriter, request *http.Request) {

	handlers.logger.ZL.Info("Handling Registration")

	// Получаем куки чтобы проверить не авторизованный ли пользователь хочет зарегистрироваться.
	_, isAuth, err := handlers.GetUserID(request)
	if err != nil {
		errorMsg := errorMsg{ErrorMessage: "Can't read cookies"}
		msg, _ := json.Marshal(errorMsg)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Write(msg)
		return
	}

	// Если авторизованный, то отдаем соответствующую ошибку.
	if isAuth {
		errorMsg := errorMsg{ErrorMessage: "Already authenticated"}
		msg, _ := json.Marshal(errorMsg)
		responseWriter.WriteHeader(http.StatusForbidden)
		responseWriter.Write(msg)
		return
	}

	// Создаем модель, для парсингда запроса.
	var userRegRequest models.UserRegReq

	// Пробуем спарсить запрос в модель.
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&userRegRequest); err != nil {
		errorMsg := errorMsg{ErrorMessage: "Not a valid user registration request"}
		msg, _ := json.Marshal(errorMsg)
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write(msg)
		return
	}

	// Спарсили, пробуем зарегистрировать нового пользователя.
	newUser, err := handlers.store.CreateUser(request.Context(), userRegRequest)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			errorMsg := errorMsg{ErrorMessage: "User with this login already exists"}
			msg, _ := json.Marshal(errorMsg)
			responseWriter.WriteHeader(http.StatusConflict)
			responseWriter.Write(msg)
			return
		}
	}

	err = handlers.auth.SetNewCookie(responseWriter, newUser.ID, newUser.Login)
	if err != nil {
		errorMsg := errorMsg{ErrorMessage: "Fail authenticate user after successful registration"}
		msg, _ := json.Marshal(errorMsg)
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write(msg)
		return
	}
	errorMsg := errorMsg{ErrorMessage: "Success authenticate user after successful registration"}
	msg, _ := json.Marshal(errorMsg)
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write(msg)
}
