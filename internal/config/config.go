package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Sender   SenderConfig   `mapstructure:"sender"`
	Resume   ResumeConfig   `mapstructure:"resume"`
	Contacts ContactsConfig `mapstructure:"contacts"`
	Scraper  ScraperConfig  `mapstructure:"scraper"`
	LLM      LLMConfig      `mapstructure:"llm"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Output   OutputConfig   `mapstructure:"output"`
}

type SenderConfig struct {
	Name  string            `mapstructure:"name"`
	Email string            `mapstructure:"email"`
	Links map[string]string `mapstructure:"links"`
}

type ResumeConfig struct {
	TextPath    string   `mapstructure:"text_path"`
	Attachments []string `mapstructure:"attachments"`
}

type ContactsConfig struct {
	Path string `mapstructure:"path"`
}

type ScraperConfig struct {
	Provider         string `mapstructure:"provider"`
	RateLimitMs      int    `mapstructure:"rate_limit_ms"`
	TimeoutMs        int    `mapstructure:"timeout_ms"`
	MaxContentLength int    `mapstructure:"max_content_length"`
	RodFallback      bool   `mapstructure:"rod_fallback"`
	FirecrawlAPIKey  string `mapstructure:"-"`
}

type LLMConfig struct {
	Provider   string  `mapstructure:"provider"`
	APIKeyEnv  string  `mapstructure:"api_key_env"`
	Model      string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens  int     `mapstructure:"max_tokens"`
	RateLimitMs int    `mapstructure:"rate_limit_ms"`
	APIKey     string  `mapstructure:"-"`
}

type SMTPConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	UsernameEnv string `mapstructure:"username_env"`
	PasswordEnv string `mapstructure:"password_env"`
	RateLimitMs int    `mapstructure:"rate_limit_ms"`
	Username    string `mapstructure:"-"`
	Password    string `mapstructure:"-"`
}

type OutputConfig struct {
	Path string `mapstructure:"path"`
}

func Load(cfgFile string) (*Config, error) {
	_ = godotenv.Load()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	cfg.LLM.APIKey = os.Getenv(cfg.LLM.APIKeyEnv)
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("environment variable %s is not set", cfg.LLM.APIKeyEnv)
	}

	cfg.SMTP.Username = os.Getenv(cfg.SMTP.UsernameEnv)
	cfg.SMTP.Password = os.Getenv(cfg.SMTP.PasswordEnv)

	if cfg.Scraper.Provider == "firecrawl" {
		cfg.Scraper.FirecrawlAPIKey = os.Getenv("FIRECRAWL_API_KEY")
		if cfg.Scraper.FirecrawlAPIKey == "" {
			return nil, fmt.Errorf("FIRECRAWL_API_KEY is required when scraper provider is firecrawl")
		}
	}

	return &cfg, nil
}
