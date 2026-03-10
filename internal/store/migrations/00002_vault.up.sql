-- Ключевой блоб пользователя (DEK_encrypted, позже PK_user, SK_user_encrypted).
-- Одна запись на пользователя; сервер не расшифровывает.
CREATE TABLE IF NOT EXISTS user_encrypted_keys
(
    user_id       INT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    encrypted_blob BYTEA NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Записи хранилища: зашифрованный payload + метаданные для списка.
CREATE TABLE IF NOT EXISTS vault_items
(
    id         BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id    INT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       VARCHAR(32) NOT NULL,
    meta_name  VARCHAR(512),
    payload    BYTEA       NOT NULL,
    created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS vault_items_user_id_idx ON vault_items (user_id);
CREATE INDEX IF NOT EXISTS vault_items_user_updated_idx ON vault_items (user_id, updated_at DESC);
