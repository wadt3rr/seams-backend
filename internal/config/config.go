package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"APP_ENV" env-default:"local"`
	HTTPServer `yaml:"http-server"`
	Database   Database `yaml:"database"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:":8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"iddle-timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"120s"`
}

type Database struct {
	Host           string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port           int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User           string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Pass           string `yaml:"pass" env:"DB_PASS" env-default:"password"`
	Name           string `yaml:"name" env:"DB_NAME" env-default:"seams_db"`
	SSLMode        string `yaml:"sslmode" env:"DB_SSLMODE" env-default:"require"`
	MigrationsPath string `yaml:"migrations-path" env:"MIGRATIONS_PATH" env-default:"./migrations"`
	Dsn            string `env:"DATABASE_URL"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	return &cfg
}
