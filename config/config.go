package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

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
		AWSConfig
		LivekitConfig
		Webhook
		TranscribeConfig
	}

	// App -.
	App struct {
		GoogleAppCredPath string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
		GinMode           string `env:"GIN_MODE" env-default:"release"`
		Debug             string `env:"DEBUG" env-default:"0"`
		Domain            string `env:"DOMAIN" env-default:""`
		Port              string `env:"PORT" env-default:"8090"`
		CertPath          string `env:"CERT_PATH" env-default:""`
		KeyPath           string `env:"KEY_PATH" env-default:""`
	}

	Mongo struct {
		MongoDbConnection string `env:"MONGO_DB_CONNECTION"`
		MongoDbName       string `env:"MONGO_DB_NAME"`
	}

	AWSConfig struct {
		AWSAccessKey string `env:"AWS_ACCESS_KEY" env-default:""`
		AWSSecret    string `env:"AWS_SECRET" env-default:""`
		AWSRegion    string `env:"AWS_REGION" env-default:""`
		AWSBucket    string `env:"AWS_BUCKET" env-default:""`
	}

	LivekitConfig struct {
		LVHost      string `env:"LIVEKIT_HOST" env-default:""`
		LVApiKey    string `env:"LIVEKIT_API_KEY" env-default:""`
		LVApiSecret string `env:"LIVEKIT_API_SECRET" env-default:""`
	}

	TranscribeConfig struct {
		TranscribeAddr string `env:"TRANSCRIBE_ADDR" env-default:"http://localhost:8099/transcriber/start"`
	}

	Webhook struct {
		WebhookURL      string `env:"WEBHOOK_URL" env-default:""`
		WebhookUsername string `env:"WEBHOOK_USERNAME" env-default:""`
		WebhookPassword string `env:"WEBHOOK_PASSWORD" env-default:""`
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
