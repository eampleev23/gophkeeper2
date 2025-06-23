package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"golang.org/x/crypto/argon2"
	"strings"
	"time"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var (
	DefaultArgon2Params = Argon2Params{
		Memory:      64 * 1024, // 64 mb
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
)

func generateRandomBytes(length uint32) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// Заводим функцию как переменную для использования в тестах.

var (
	VerifyPassword = func(password, encodedHash string) (match bool, err error) {
		// Распаковываем параметры из хэша
		p, salt, hash, err := decodeHash(encodedHash)
		if err != nil {
			return false, fmt.Errorf("failed to decode hash: %w", err)
		}

		// Derive the key from the other password using the same parameters
		otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

		// Check that the contents of the hashed passwords are identical
		if len(hash) != len(otherHash) {
			return false, nil
		}
		for i := range hash {
			if hash[i] != otherHash[i] {
				return false, nil
			}
		}
		return true, nil
	}
)

func decodeHash(encodedHash string) (p *Argon2Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("incompatible argon2 version: %w", err)
	}

	p = &Argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("incompatible argon2 parameters: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}
	p.KeyLength = uint32(len(hash))
	return p, salt, hash, nil
}

func (d DBStore) CreateUser(ctx context.Context, req models.UserRegReq) (newUser *models.User, err error) {

	// Выделяем память под модель пользователя.
	newUser = &models.User{}

	// Генерируем соль.
	salt, err := generateRandomBytes(DefaultArgon2Params.SaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Хэшируем пароль с солью
	hash := argon2.IDKey(
		[]byte(req.Password),
		salt,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Parallelism,
		DefaultArgon2Params.KeyLength,
	)

	// Подгатавливаем данные для сохранения
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Parallelism,
		b64Salt, b64Hash)

	// Текущее время для created_at и updated_at
	now := time.Now()

	// Вставляем пользователя в БД.
	err = d.dbConn.QueryRow(
		`INSERT INTO users 
         (login, password_hash, salt, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5) 
         RETURNING id`,
		req.Login,
		encodedHash,
		b64Salt,
		now,
		now,
	).Scan(&newUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}
	newUser.Login = req.Login
	newUser.CreatedAt = now
	newUser.UpdatedAt = now
	newUser.PasswordHash = encodedHash

	return newUser, nil
}
