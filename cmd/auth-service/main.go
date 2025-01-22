package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"
	"auth-service/internal/lib/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.MustSetupLogger(cfg.Env)

	application := app.NewApp(
		log,
		cfg.GRPC.Port,
		cfg.PostgresURL,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	application.GrpcApp.MustRun()

	stopWait(application)
}

func stopWait(application *app.App) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	application.GrpcApp.Stop()
}
