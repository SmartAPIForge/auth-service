package app

import (
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/kafka"
	authservice "auth-service/internal/services/auth"
	userservice "auth-service/internal/services/user"
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
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	schemaRegistryUrl string,
	kafkaHost string,
) *App {
	storage, err := postgres.NewStorage(postgresURL)
	if err != nil {
		panic(err)
	}

	schemaManager := kafka.NewSchemaManager(schemaRegistryUrl)
	kafkaProducer := kafka.NewKafkaProducer(kafkaHost, log, schemaManager)

	authService := authservice.NewAuthService(log, storage, accessTokenTTL, refreshTokenTTL, kafkaProducer)
	userService := userservice.NewUserService(log, storage)

	grpcApp := grpcapp.NewGrpcApp(
		log,
		authService,
		userService,
		grpcPort,
	)

	return &App{
		GrpcApp: grpcApp,
	}
}
