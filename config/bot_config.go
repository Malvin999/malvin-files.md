package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type BotConfig struct {
	StorageDir     string `default:"./storage"  envconfig:"STORAGE_DIR"`
	BotAPIToken    string `required:"true" envconfig:"BOT_API_TOKEN"`
	ConfigFilename string `default:"config.json"`
	HabitsHost     string `default:"" envconfig:"HABITS_HOST"`
	ServerCertDir  string `default:"/tmp" envconfig:"SERVER_CERT_DIR"`
	ServerLogFile  string `default:"/tmp/server.log" envconfig:"SERVER_LOG_FILE"`
}

var BotCfg BotConfig

func LoadBotConfig() error {
	if err := envconfig.Process("", &BotCfg); err != nil {
		return fmt.Errorf("can't load config: %w", err)
	}

	return nil
}
