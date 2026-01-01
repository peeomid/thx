package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultFormat = "default"
	defaultWhen   = "inbox"
)

// Config captures user configuration loaded from file/env/defaults.
type Config struct {
	Format   string         `mapstructure:"format"`
	Database string         `mapstructure:"database"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
}

// DefaultsConfig holds defaults for new todos.
type DefaultsConfig struct {
	When string   `mapstructure:"when"`
	Tags []string `mapstructure:"tags"`
}

// Load reads config from path (optional), environment, and defaults.
func Load(path string) (Config, error) {
	v := viper.New()
	v.SetDefault("format", defaultFormat)
	v.SetDefault("defaults.when", defaultWhen)
	v.SetDefault("defaults.tags", []string{})

	v.SetEnvPrefix("THX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, err
		}
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME/.config/thx")
		if err := v.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) {
				return Config{}, err
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
