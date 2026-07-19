package bootstrap

import (
	"time"

	"repin/internal/pkg/db"
	"repin/internal/pkg/httpx"
	"repin/internal/pkg/logger"
	"repin/internal/pkg/migration"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"develop"`

	HTTP struct {
		Host string `env:"HTTP_SERVER_HOST" envDefault:"0.0.0.0"`
		Port string `env:"HTTP_SERVER_PORT" envDefault:"8080"`
	}

	Database struct {
		URL    string `env:"DATABASE_URL"`
		Schema string `env:"DATABASE_SCHEMA" envDefault:"public"`
	}

	Migration struct {
		Dir   string `env:"MIGRATIONS_DIR" envDefault:"migrations"`
		Table string `env:"MIGRATIONS_TABLE" envDefault:"schema_migrations"`
	}

	Telegram struct {
		AppID          int           `env:"TELEGRAM_API_ID"`
		AppHash        string        `env:"TELEGRAM_API_HASH"`
		Phone          string        `env:"TELEGRAM_PHONE"`
		Channel        string        `env:"TELEGRAM_CHANNEL"`
		BotToken       *string       `env:"TELEGRAM_BOT_TOKEN"`
		ChannelRefresh time.Duration `env:"TELEGRAM_CHANNEL_REFRESH_INTERVAL" envDefault:"6h"`
	}

	Proxy struct {
		URL *string `env:"PROXY_URL"`
	}

	OpenRouter struct {
		APIKey        *string       `env:"OPENROUTER_API_KEY"`
		Model         string        `env:"OPENROUTER_MODEL" envDefault:"google/gemini-2.5-flash-lite"`
		FallbackModel string        `env:"OPENROUTER_FALLBACK_MODEL" envDefault:"openai/gpt-4o-mini"`
		Timeout       time.Duration `env:"OPENROUTER_TIMEOUT" envDefault:"90s"`
		MaxRetries    int           `env:"OPENROUTER_MAX_RETRIES" envDefault:"2"`
		Referer       string        `env:"OPENROUTER_REFERER"`
	}

	Worker struct {
		Concurrency  int           `env:"WORKER_CONCURRENCY" envDefault:"2"`
		PollInterval time.Duration `env:"WORKER_POLL_INTERVAL" envDefault:"2s"`
		JobLease     time.Duration `env:"WORKER_JOB_LEASE" envDefault:"30m"`
	}

	Favicon struct {
		Dir *string `env:"FAVICON_DIR"`
	}

	Storage struct {
		Endpoint  string `env:"MINIO_ENDPOINT"`
		AccessKey string `env:"MINIO_ACCESS_KEY"`
		SecretKey string `env:"MINIO_SECRET_KEY"`
		Bucket    string `env:"MINIO_BUCKET"`
	}

	Logger struct {
		Debug bool `env:"LOGGER_DEBUG" envDefault:"true"`
	}
}

func (c Config) TelegramBotToken() string {
	return stringOrEmpty(c.Telegram.BotToken)
}

func (c Config) ProxyURL() string {
	return stringOrEmpty(c.Proxy.URL)
}

func (c Config) OpenRouterKey() string {
	return stringOrEmpty(c.OpenRouter.APIKey)
}

func (c Config) FaviconDir() string {
	return stringOrEmpty(c.Favicon.Dir)
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func (c Config) PGConfig() db.Config {
	return db.Config{URL: c.Database.URL, Schema: c.Database.Schema}
}

func (c Config) MigrationConfig() migration.Config {
	return migration.Config{Dir: c.Migration.Dir, Table: c.Migration.Table}
}

func (c Config) LoggerConfig() logger.Config {
	return logger.Config{Debug: c.Logger.Debug}
}

func (c Config) HTTPConfig() httpx.Config {
	return httpx.Config{Host: c.HTTP.Host, Port: c.HTTP.Port}
}
