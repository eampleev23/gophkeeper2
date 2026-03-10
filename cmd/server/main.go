package main

import (
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/handlers"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/middlewares"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	servConfig := server_config.NewServerConfig()
	logger, err := logger.NewZapLogger(servConfig.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create zap logger: %w", err)
	}
	logger.ZL.Debug("logger created")
	auth, err := auth.Initialize(servConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize a new authorizer: %w", err)
	}
	store, err := store.NewStorage(servConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	if len(servConfig.DBDSN) != 0 {
		// Отложенно закрываем соединение с бд.
		defer func() {
			if err := store.DBConnClose(); err != nil {
				logger.ZL.Info("store failed to properly close the DB connection")
			}
		}()
	}
	handlers, err := handlers.NewHandlers(store, servConfig, logger, auth)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	logger.ZL.Info("Running server", zap.String("address", servConfig.RunAddr))

	routers := chi.NewRouter()

	routers.Use(logger.RequestLogger)
	routers.Use(middlewares.CheckAndSetContenType)

	routers.Get("/", handlers.Root)
	routers.Get("/health", handlers.Health)

	routers.Group(func(router chi.Router) {
		router.Use(auth.MiddleCheckNoAuth)
		router.Post("/api/user/login/", handlers.Login)
		router.Post("/api/user/registration/", handlers.Registration)
	})

	routers.Group(func(router chi.Router) {
		router.Use(auth.MiddleCheckAuth)
		router.Post("/api/user/logout/", handlers.Logout)

		// Ключевой блоб пользователя (зашифрованный DEK и т.д.)
		router.Get("/api/vault/keys", handlers.GetUserKeys)
		router.Put("/api/vault/keys", handlers.SetUserKeys)

		// Записи хранилища (vault items)
		router.Get("/api/vault/items", handlers.ListVaultItems)
		router.Post("/api/vault/items", handlers.CreateVaultItem)
		router.Get("/api/vault/items/{id}", handlers.GetVaultItem)
		router.Put("/api/vault/items/{id}", handlers.UpdateVaultItem)
		router.Delete("/api/vault/items/{id}", handlers.DeleteVaultItem)
	})

	if servConfig.TLSEnabled {
		err = http.ListenAndServeTLS(servConfig.RunAddr, "server.crt", "server.key", routers)
	} else {
		err = http.ListenAndServe(servConfig.RunAddr, routers)
	}
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
