package config

import (
	"api/config/env"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v81"
)

type Config struct {
	AppEnv env.Env `env:"APP_ENV" env-default:"local"`
	Port   string  `env:"PORT" env-default:"3001"`

	PgDBURL  string `env:"PG_DB_URL" env-default:"postgres://postgres:postgres@0.0.0.0:5432/postgres"`
	RedisURL string `env:"REDIS_URL" env-default:"redis://default:redis@0.0.0.0:6379"`

	SecretAuthCodeValue     string `env:"SECRET_AUTH_CODE_VALUE" env-default:"123456"`
	SecretAuthCodeValueHash string `env:"SECRET_AUTH_CODE_VALUE_HASH" env-default:"123456"`
	SecretSession           string `env:"SECRET_SESSION" env-default:"123456"`

	SessionTTL          time.Duration `env:"SESSION_TTL" env-default:"1209600s"`
	SessionCookieDomain string        `env:"SESSION_COOKIE_DOMAIN" env-default:"localhost"`

	AppHost     string `env:"APP_HOST" env-default:"localhost:5173"`
	AppProtocol string `env:"APP_PROTOCOL" env-default:"http"`

	StripeSecretKey            string `env:"STRIPE_SECRET_KEY"`
	StripeSigningSecretWebhook string `env:"STRIPE_SIGNING_SECRET_WEBHOOK"`

	ResendAPIKey string `env:"RESEND_API_KEY"`
	SMTPAddr     string `env:"SMTP_ADDR" env-default:"0.0.0.0:1025"`

	SlackClientID       string `env:"SLACK_CLIENT_ID"`
	SlackClientSecret   string `env:"SLACK_CLIENT_SECRET"`
	SlackRedirectURI    string `env:"SLACK_REDIRECT_URI" env-default:"https://supposedly-simple-mallard.ngrok-free.app/app/slack/oauth"`
	SlackStatusBotToken string `env:"SLACK_STATUS_BOT_TOKEN"`

	GitHubAppID        int64  `env:"GITHUB_APP_ID"`
	GitHubClientID     string `env:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string `env:"GITHUB_CLIENT_SECRET"`
	GitHubPrivateKey   string `env:"GITHUB_PRIVATE_KEY"`
	GitHubRedirectURL  string `env:"GITHUB_REDIRECT_URL" env-default:"http://localhost:3001/oauth/github/callback"`

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL" env-default:"http://localhost:3001/oauth/google/callback"`

	ReplicateAPIToken string `env:"REPLICATE_API_TOKEN"`

	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
}

func MustNewConfig() *Config {
	useSystemEnv := strings.ToLower(os.Getenv("USE_SYSTEM_ENV"))
	appEnv := env.FromString(os.Getenv("APP_ENV"))
	if useSystemEnv == "" || useSystemEnv == "false" {
		err := godotenv.Load(".env." + string(appEnv))
		if err != nil {
			panic(err)
		}
	}

	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}

	cfg.AppEnv = appEnv
	stripe.Key = cfg.StripeSecretKey

	return &cfg
}
