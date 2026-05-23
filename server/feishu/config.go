package feishu

import (
	"strings"

	"github.com/zakirullin/files.md/server/config"
)

type Config struct {
	AppID             string
	AppSecret         string
	AllowedOpenIDs    []string
	DefaultUserID     int64
	EnableCardActions bool
}

func ConfigFromServer(cfg config.Config) Config {
	return Config{
		AppID:             cfg.FeishuAppID,
		AppSecret:         cfg.FeishuAppSecret,
		AllowedOpenIDs:    splitCSV(cfg.FeishuAllowedOpenIDs),
		DefaultUserID:     cfg.FeishuDefaultUserID,
		EnableCardActions: cfg.FeishuEnableCardActions,
	}
}

func (c Config) Enabled() bool {
	return c.AppID != "" && c.AppSecret != ""
}

func splitCSV(s string) []string {
	var values []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}
