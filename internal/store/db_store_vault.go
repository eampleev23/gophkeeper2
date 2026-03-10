package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
)

// GetUserEncryptedKeys возвращает зашифрованный ключевой блоб пользователя.
func (d *DBStore) GetUserEncryptedKeys(ctx context.Context, userID int) ([]byte, error) {
	var blob []byte
	err := d.dbConn.QueryRowContext(ctx,
		`SELECT encrypted_blob FROM user_encrypted_keys WHERE user_id = $1`,
		userID,
	).Scan(&blob)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user encrypted keys: %w", err)
	}
	return blob, nil
}

// SetUserEncryptedKeys сохраняет или обновляет зашифрованный ключевой блоб пользователя.
func (d *DBStore) SetUserEncryptedKeys(ctx context.Context, userID int, encryptedBlob []byte) error {
	_, err := d.dbConn.ExecContext(ctx,
		`INSERT INTO user_encrypted_keys (user_id, encrypted_blob, updated_at)
		 VALUES ($1, $2, CURRENT_TIMESTAMP)
		 ON CONFLICT (user_id) DO UPDATE SET encrypted_blob = $2, updated_at = CURRENT_TIMESTAMP`,
		userID, encryptedBlob,
	)
	if err != nil {
		return fmt.Errorf("set user encrypted keys: %w", err)
	}
	return nil
}

// CreateVaultItem создаёт запись хранилища; возвращает id.
func (d *DBStore) CreateVaultItem(ctx context.Context, userID int, itemType, metaName string, payload []byte) (int64, error) {
	var id int64
	err := d.dbConn.QueryRowContext(ctx,
		`INSERT INTO vault_items (user_id, type, meta_name, payload) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, itemType, metaName, payload,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create vault item: %w", err)
	}
	return id, nil
}

// ListVaultItems возвращает метаданные записей пользователя без payload.
func (d *DBStore) ListVaultItems(ctx context.Context, userID int) ([]models.VaultItemMeta, error) {
	rows, err := d.dbConn.QueryContext(ctx,
		`SELECT id, type, meta_name, updated_at FROM vault_items WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list vault items: %w", err)
	}
	defer rows.Close()

	var list []models.VaultItemMeta
	for rows.Next() {
		var m models.VaultItemMeta
		if err := rows.Scan(&m.ID, &m.Type, &m.MetaName, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vault item meta: %w", err)
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// GetVaultItem возвращает запись по id, только если она принадлежит userID.
func (d *DBStore) GetVaultItem(ctx context.Context, userID int, itemID int64) (*models.VaultItem, error) {
	item := &models.VaultItem{}
	err := d.dbConn.QueryRowContext(ctx,
		`SELECT id, user_id, type, meta_name, payload, created_at, updated_at
		 FROM vault_items WHERE id = $1 AND user_id = $2`,
		itemID, userID,
	).Scan(&item.ID, &item.UserID, &item.Type, &item.MetaName, &item.Payload, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get vault item: %w", err)
	}
	return item, nil
}

// UpdateVaultItem обновляет meta_name и payload записи; запись должна принадлежать userID.
func (d *DBStore) UpdateVaultItem(ctx context.Context, userID int, itemID int64, metaName string, payload []byte) error {
	res, err := d.dbConn.ExecContext(ctx,
		`UPDATE vault_items SET meta_name = $1, payload = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND user_id = $4`,
		metaName, payload, itemID, userID,
	)
	if err != nil {
		return fmt.Errorf("update vault item: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteVaultItem удаляет запись; она должна принадлежать userID.
func (d *DBStore) DeleteVaultItem(ctx context.Context, userID int, itemID int64) error {
	res, err := d.dbConn.ExecContext(ctx,
		`DELETE FROM vault_items WHERE id = $1 AND user_id = $2`,
		itemID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete vault item: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
