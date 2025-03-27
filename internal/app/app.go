package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	gatewayPort int,
	storageCfg config.StorageConfig,
	tokenTTL time.Duration,
) *App {
	
	storage, err := postgres.New(storageCfg)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, storage, tokenTTL)

	grpcApp, err := grpcapp.New(log, authService, grpcPort, gatewayPort)

	if err != nil {
		panic(err)
	}

	return &App{
		GRPCServer: grpcApp,
	}
}
