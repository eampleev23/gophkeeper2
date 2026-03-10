package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

// Ошибки хранилища

var (
	ErrUserNotFound = errors.New("user not found")
)

type Store interface {
	DBConnClose() (err error)
	CreateUser(ctx context.Context, userRegReq models.UserRegReq) (newUser *models.User, err error)
	GetUserByLogin(ctx context.Context, userLoginReq models.UserLoginReq) (userModelResponse *models.User, err error)

	// Ключевой блоб пользователя (DEK_encrypted и т.д.); сервер не расшифровывает.
	GetUserEncryptedKeys(ctx context.Context, userID int) (encryptedBlob []byte, err error)
	SetUserEncryptedKeys(ctx context.Context, userID int, encryptedBlob []byte) error

	// Записи хранилища (vault).
	CreateVaultItem(ctx context.Context, userID int, itemType, metaName string, payload []byte) (id int64, err error)
	ListVaultItems(ctx context.Context, userID int) ([]models.VaultItemMeta, error)
	GetVaultItem(ctx context.Context, userID int, itemID int64) (*models.VaultItem, error)
	UpdateVaultItem(ctx context.Context, userID int, itemID int64, metaName string, payload []byte) error
	DeleteVaultItem(ctx context.Context, userID int, itemID int64) error
}

func NewStorage(serv_conf *server_config.ServerConfig, logger *logger.ZapLog) (Store, error) {
	s, err := NewDBStore(serv_conf, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating new db store: %w", err)
	}
	logger.ZL.Debug("DB store created success..")
	return s, nil
}
