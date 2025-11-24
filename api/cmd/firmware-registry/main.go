package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"firmware-registry-api/internal/api"
	"firmware-registry-api/internal/api/handlers"
	"firmware-registry-api/internal/auth"
	"firmware-registry-api/internal/config"
	"firmware-registry-api/internal/db"
	"firmware-registry-api/internal/firmware"
	"firmware-registry-api/internal/logging"
	"firmware-registry-api/internal/webhook"

	"github.com/rs/zerolog/log"
)

func main() {
	cfgPath := os.Getenv("FW_CONFIG_FILE")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Config load failed")
	}

	// Initialize logger
	if err := logging.Setup(cfg); err != nil {
		log.Fatal().Err(err).Msg("Logger setup failed")
	}

	log.Info().
		Str("version", "1.0.0").
		Str("listen_addr", cfg.ListenAddr).
		Msg("Firmware Registry API starting")

	// Ensure directories exist
	if err := os.MkdirAll(cfg.StorageDir, 0o755); err != nil {
		log.Fatal().Err(err).Str("dir", cfg.StorageDir).Msg("Failed to create storage directory")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatal().Err(err).Str("dir", filepath.Dir(cfg.DBPath)).Msg("Failed to create database directory")
	}

	// DB + migrations
	log.Info().Str("db_path", cfg.DBPath).Msg("Opening database")
	database := db.OpenSQLite(cfg.DBPath)
	log.Info().Msg("Running database migrations")
	db.RunMigrations(cfg.DBPath, "./migrations")

	// Firmware layer
	fwRepo := &firmware.SQLiteRepo{DB: database}
	fwSvc := &firmware.Service{
		Repo:       fwRepo,
		Storage:    firmware.Storage{BaseDir: cfg.StorageDir},
		PublicBase: cfg.PublicBaseURL,
	}

	// Webhook layer
	whRepo := &webhook.SQLiteRepo{DB: database}
	whSvc := &webhook.Service{
		Repo:       whRepo,
		Secret:     cfg.Webhooks.Secret,
		TimeoutSec: cfg.Webhooks.TimeoutSec,
		Retries:    cfg.Webhooks.Retries,
	}

	// Initialize OIDC verifier if enabled
	var oidcVerifier *auth.OIDCVerifier
	if cfg.OIDC.Enabled {
		log.Info().Str("issuer", cfg.OIDC.IssuerURL).Msg("Initializing OIDC authentication")
		ctx := context.Background()
		var err error
		oidcVerifier, err = auth.NewOIDCVerifier(
			ctx,
			cfg.OIDC.IssuerURL,
			cfg.OIDC.ClientID,
			cfg.OIDC.Audience,
			cfg.OIDC.AdminRole,
			cfg.OIDC.DeviceRole,
		)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("OIDC enabled but failed to initialize, falling back to API key authentication only")
			cfg.OIDC.Enabled = false
		} else {
			log.Info().
				Str("issuer", cfg.OIDC.IssuerURL).
				Str("client_id", cfg.OIDC.ClientID).
				Str("admin_role", cfg.OIDC.AdminRole).
				Str("device_role", cfg.OIDC.DeviceRole).
				Msg("OIDC authentication enabled")
		}
	}

	authHandler := auth.Auth{
		AdminKey:     cfg.AdminKey,
		DeviceKey:    cfg.DeviceKey,
		OIDCEnabled:  cfg.OIDC.Enabled,
		OIDCVerifier: oidcVerifier,
	}

	fwHandler := &handlers.FirmwareHandler{
		Auth:     authHandler,
		Service:  fwSvc,
		Webhooks: whSvc,
		MaxBytes: cfg.MaxUploadMB * 1024 * 1024,
	}
	whHandler := &handlers.WebhookHandler{
		Auth: authHandler,
		Repo: whRepo,
	}

	router := api.NewRouter(fwHandler, whHandler)

	// Apply middlewares: logging first, then CORS
	handler := logging.HTTPLogger(router)
	handler = api.CORSMiddleware(handler)

	log.Info().
		Str("listen_addr", cfg.ListenAddr).
		Msg("Firmware Registry API listening")

	if err := http.ListenAndServe(cfg.ListenAddr, handler); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
