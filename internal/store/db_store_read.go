package store

import (
	"context"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
)

func (d DBStore) GetUserByLogin(ctx context.Context, userLoginReq models.UserLoginReq) (userModelResponse *models.User, err error) {

	userModelResponse = &models.User{}

	// Получаем данные по логину.
	row := d.dbConn.QueryRowContext(ctx,
		`SELECT id, login, password_hash, salt FROM users WHERE login = $1 LIMIT 1`,
		userLoginReq.Login,
	)

	// Разбираем результат.
	err = row.Scan(
		&userModelResponse.ID,
		&userModelResponse.Login,
		&userModelResponse.PasswordHash,
		&userModelResponse.Salt,
	)

	if err != nil {
		return userModelResponse, fmt.Errorf("faild to get user by login and password like this %w", err)
	}

	return userModelResponse, err
}
