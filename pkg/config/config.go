package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Tistory TistoryConfig
	Google  GoogleConfig
	State   StateConfig
	HTTP    HTTPConfig
}

type TistoryConfig struct {
	SitemapURL string `mapstructure:"sitemap_url"`
}

type GoogleConfig struct {
	SiteURL       string `mapstructure:"site_url"`
	SitemapURL    string `mapstructure:"sitemap_url"`
	ClientID      string `mapstructure:"client_id"`
	ClientSecret  string `mapstructure:"client_secret"`
	RefreshToken  string `mapstructure:"refresh_token"`
}

type StateConfig struct {
	FilePath string `mapstructure:"file_path"`
}

type HTTPConfig struct {
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
	UserAgent      string `mapstructure:"user_agent"`
}

func (c *HTTPConfig) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

func Load(envFile, configFile string) (*Config, error) {
	// load .env (ignore error if file doesn't exist)
	_ = godotenv.Load(envFile)

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.SetDefault("http.timeout_seconds", 10)
	v.SetDefault("http.user_agent", "tistory-indexer/1.0")
	v.SetDefault("state.file_path", "./data/last_state.json")

	v.AutomaticEnv()
	v.BindEnv("google.client_id", "GOOGLE_CLIENT_ID")         //nolint:errcheck
	v.BindEnv("google.client_secret", "GOOGLE_CLIENT_SECRET") //nolint:errcheck
	v.BindEnv("google.refresh_token", "GOOGLE_REFRESH_TOKEN") //nolint:errcheck

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", configFile, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Tistory.SitemapURL == "" {
		return fmt.Errorf("tistory.sitemap_url is required")
	}
	if c.Google.SiteURL == "" {
		return fmt.Errorf("google.site_url is required")
	}
	if c.Google.SitemapURL == "" {
		return fmt.Errorf("google.sitemap_url is required")
	}
	if c.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID env is required")
	}
	if c.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET env is required")
	}
	if c.Google.RefreshToken == "" {
		return fmt.Errorf("GOOGLE_REFRESH_TOKEN env is required")
	}
	return nil
}
