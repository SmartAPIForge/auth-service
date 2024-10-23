package app

import (
	grpcapp "auth-service/internal/app/grpc"
	authservice "auth-service/internal/services/auth"
	"auth-service/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GrpcApp *grpcapp.GrpcApp
}

func NewApp(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	authService := authservice.NewAuthService(log, storage, tokenTTL)

	grpcApp := grpcapp.NewGrpcApp(log, authService, grpcPort)

	return &App{
		GrpcApp: grpcApp,
	}
}
