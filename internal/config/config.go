package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"APP_ENV" env-default:"production"`
	HTTPServer HTTPServer
	Database   Database
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:":8080"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"120s"`
}

type Database struct {
	Host           string `env:"DB_HOST"`
	Port           int    `env:"DB_PORT"`
	User           string `env:"DB_USER"`
	Pass           string `env:"DB_PASSWORD"`
	Name           string `env:"DB_NAME"`
	SSLMode        string `env:"DB_SSLMODE" env-default:"disable"`
	MigrationsPath string `env:"MIGRATIONS_PATH" env-default:"./migrations"`
	Dsn            string `env:"DATABASE_URL" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Failed to read config from env: %s", err)
	}

	return &cfg
}
