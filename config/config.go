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
		RedisConfig
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
		MongoDbConnection string `env:"MONGO_DB_CONNECTION" env-default:"mongodb://127.0.0.1:27017"`
		MongoDbName       string `env:"MONGO_DB_NAME" env-default:"teleporta"`
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
		TranscribeAddr string `env:"TRANSCRIBE_ADDR" env-default:"http://10.0.0.127:8095/transcriber/"`
	}

	Webhook struct {
		WebhookURL      string `env:"WEBHOOK_URL" env-default:""`
		WebhookUsername string `env:"WEBHOOK_USERNAME" env-default:""`
		WebhookPassword string `env:"WEBHOOK_PASSWORD" env-default:""`
	}

	RedisConfig struct {
		RedisHost string `env:"REDIS_HOST" env-default:"127.0.0.1"`
		RedisPort string `env:"REDIS_PORT" env-default:"6379"`
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
