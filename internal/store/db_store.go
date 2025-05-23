package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStore struct {
	dbConn *sql.DB
	c      *server_config.ServerConfig
	l      *logger.ZapLog
}

func (d DBStore) DBConnClose() (err error) {
	if err := d.dbConn.Close(); err != nil {
		return fmt.Errorf("failed to properly close the DB connection %w", err)
	}
	return nil
}

func NewDBStore(c *server_config.ServerConfig, l *logger.ZapLog) (*DBStore, error) {
	db, err := sql.Open("pgx", c.DBDSN)
	if err != nil {
		return &DBStore{}, fmt.Errorf("%w", errors.New("sql.open failed in case to create store"))
	}
	if err := runMigrations(c.DBDSN); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	return &DBStore{
		dbConn: db,
		c:      c,
		l:      l,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
