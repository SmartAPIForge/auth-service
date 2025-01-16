package app

import (
	grpcapp "auth-service/internal/app/grpc"
	authservice "auth-service/internal/services/auth"
	"auth-service/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GrpcApp *grpcapp.GrpcApp
}

func NewApp(
	log *slog.Logger,
	grpcPort int,
	postgresURL string,
	tokenTTL time.Duration,
) *App {
	storage, err := postgres.NewStorage(postgresURL)
	if err != nil {
		panic(err)
	}

	authService := authservice.NewAuthService(log, storage, tokenTTL)
	// tokenService
	// userService

	grpcApp := grpcapp.NewGrpcApp(
		log,
		authService,
		grpcPort,
	)

	return &App{
		GrpcApp: grpcApp,
	}
}
