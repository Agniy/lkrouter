package config

import (
	"fmt"
	"log"
)
import "github.com/ilyakaznacheev/cleanenv"

var cfg *Config
var cfgError error

func init() {
	cfg, cfgError = NewConfig()
	if cfgError != nil {
		log.Fatalf("Config error: %s", cfgError)
	}
}

func GetConfig() *Config {
	return cfg
}

type (
	// Config -.
	Config struct {
		App
		Mongo
	}

	// App -.
	App struct {
		GoogleAppCredPath string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
		Debug             bool   `env:"DEBUG" env-default:"false"`
		Domain            string `env:"DOMAIN" env-default:"localhost"`
		Port              string `env:"PORT" env-default:"8080"`
		CertPath          string `env:"CERT_PATH" env-default:""`
		KeyPath           string `env:"KEY_PATH" env-default:""`
	}

	Mongo struct {
		MongoDbConnection string `env:"MONGO_DB_CONNECTION"`
		MongoDbName       string `env:"MONGO_DB_NAME"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./.env", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
