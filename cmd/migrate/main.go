// Программа только применяет миграции к БД и выходит. Используется в CI и при ручном запуске.
// Переменная окружения: DATABASE_URL (или DATABASE_URI) — строка подключения к PostgreSQL.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eampleev23/gophkeeper2.git/internal/store"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URI")
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL or DATABASE_URI is required")
	}
	if err := store.RunMigrations(dsn); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	fmt.Println("migrations applied OK")
}
