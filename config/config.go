package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

type BotConfig struct {
	WorkingDir     string
	StorageDir     string `default:"./storage"  envconfig:"STORAGE_DIR"`
	BotAPIToken    string `required:"true" envconfig:"BOT_API_TOKEN"`
	ConfigFilename string `default:"config.json"`
	ApiHost        string `default:"" envconfig:"API_HOST"`
	AppHost        string `default:"" envconfig:"APP_HOST"`
	ServerCertDir  string `default:"/tmp" envconfig:"CERT_DIR"`
	TokensDir      string `default:"/tmp" envconfig:"TOKENS_DIR"`
	TokensSalt     string `envconfig:"TOKENS_SALT"`
	ServerLogFile  string `default:"/tmp/server.log" envconfig:"LOG_FILE"`
}

var BotCfg BotConfig

func LoadBotConfig() error {
	if err := envconfig.Process("", &BotCfg); err != nil {
		return fmt.Errorf("can't load config: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("config can't get working directory: %w", err)
	}
	BotCfg.WorkingDir = wd

	if !filepath.IsAbs(BotCfg.StorageDir) {
		BotCfg.StorageDir = filepath.Join(wd, BotCfg.StorageDir)
	}

	return nil
}
