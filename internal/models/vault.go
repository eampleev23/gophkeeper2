package models

import "time"

// VaultItem — полная запись хранилища (для get).
type VaultItem struct {
	ID        int64     `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	MetaName  string    `json:"meta_name,omitempty"`
	Payload   []byte    `json:"-"` // зашифрованный payload, не в JSON по умолчанию
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VaultItemMeta — метаданные для списка (без payload).
type VaultItemMeta struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	MetaName  string    `json:"meta_name,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
