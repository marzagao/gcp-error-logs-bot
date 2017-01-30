package config

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
)

type Config struct {
	Server   *server.Config
	Slack    *Slack
	GCP      *GCP
	LogLevel string `envconfig:"LOG_LEVEL"`
	Timezone string `envconfig:"BOT_TIMEZONE"`
}

type Slack struct {
	WebhookURL string `envconfig:"BOT_SLACK_WEBHOOK_URL"`
}

type GCP struct {
	ProjectID string `envconfig:"BOT_GCP_PROJECT_ID"`
	LogName   string `envconfig:"BOT_GCP_LOG_NAME"`
}

func LoadConfig() *Config {
	var cfg Config
	config.LoadEnvConfig(&cfg)
	return &cfg
}
