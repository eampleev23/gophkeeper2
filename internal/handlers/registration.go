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

func (handlers *Handlers) Registration(responseWriter http.ResponseWriter, request *http.Request) {

	// Создаем модель, для парсингда запроса.
	var userRegRequest models.UserRegReq

	// Пробуем спарсить запрос в модель.
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&userRegRequest); err != nil {
		sendResponse(true, "Not a valid user registration request", http.StatusBadRequest, responseWriter)
		return
	}

	// Спарсили, пробуем зарегистрировать нового пользователя.
	newUser, err := handlers.store.CreateUser(request.Context(), userRegRequest)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			sendResponse(true, "User with this login already exists", http.StatusConflict, responseWriter)
			return
		}
	}

	// Зарегистрировали, авторизуем сразу на лету.
	err = handlers.auth.SetNewCookie(responseWriter, newUser.ID, newUser.Login)
	if err != nil {
		sendResponse(false, "Fail authenticate user after successful registration", http.StatusOK, responseWriter)
		return
	}

	sendResponse(false, "Success authenticate user after successful registration", http.StatusOK, responseWriter)
}
