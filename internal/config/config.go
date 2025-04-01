package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env               string // dev || prod
	GRPC              GRPCConfig
	PostgresURL       string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	SchemaRegistryUrl string
	KafkaHost         string
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}

func MustLoad() *Config {
	loadEnvFile()

	env := getEnv("ENV", "dev")
	grpcPort := getEnvAsInt("GRPC_PORT", 50051)
	grpcTimeout := getEnvAsDuration("GRPC_TIMEOUT", 10*time.Second)
	postgresURL := buildPostgresURL()
	accessTokenTTL := getEnvAsDuration("ACCESS_TOKEN_TTL", 30*time.Minute)
	refreshTokenTTL := getEnvAsDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	schemaRegistryUrl := getEnv("SCHEMA_REGISTRY_URL", "http://localhost:6767")
	kafkaHost := getEnv("KAFKA_HOST", "http://localhost:9092")

	if postgresURL == "" {
		panic("postgresURL is required but not set")
	}

	return &Config{
		Env: env,
		GRPC: GRPCConfig{
			Port:    grpcPort,
			Timeout: grpcTimeout,
		},
		PostgresURL:       postgresURL,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		SchemaRegistryUrl: schemaRegistryUrl,
		KafkaHost:         kafkaHost,
	}
}

func loadEnvFile() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func buildPostgresURL() string {
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	db := getEnv("POSTGRES_DB", "main")
	port := getEnv("POSTGRES_PORT", "5431")
	host := getEnv("POSTGRES_HOST", "localhost")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db)
}
