package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env            string     `yaml:"env" env-default:"dev"` // dev || prod
	GRPC           GRPCConfig `yaml:"grpc"`
	StoragePath    string     `yaml:"storage_path" env-required:"true"`
	MigrationsPath string
	AccessTokenTTL time.Duration `yaml:"access_token_ttl" env-default:"10m"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	res := os.Getenv("CONFIG_PATH")
	return res
}
