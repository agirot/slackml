package config

import (
	"fmt"
	"github.com/cristalhq/aconfig"
)

type Conf struct {
	CacheFile      string `default:"cache.json"`
	Mailing        []List `json:"mailing"`
	SlackChannelID string `required:"true"`
	SlackToken     string `required:"true"`

	CheckInterval string `default:"5s"`
	BackgroundMod bool   `default:"false"`
}

type List struct {
	ID           string `json:"id"`
	TitleFilters []string
	InitCount    float64
}

func InitConfig() (Conf, error) {
	var cfg Conf
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		AllowUnknownEnvs: true,
		EnvPrefix:        "ML",
		FlagPrefix:       "ML",
		Files:            []string{"slackml.json"},
	})

	loader.Flags()
	if err := loader.Load(); err != nil {
		return Conf{}, fmt.Errorf("can't load config file: %w", err)
	}

	return cfg, nil
}
