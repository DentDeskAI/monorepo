// Package config loads and validates application configuration from environment
// variables and optional .env files using Viper.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	WhatsApp  WhatsAppConfig
	LLM       LLMConfig
	TenantMode string
}

type AppConfig struct {
	Env  string
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type WhatsAppConfig struct {
	APIURL          string
	PhoneNumberID   string
	AccessToken     string
	VerifyToken     string
}

type LLMConfig struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// Load reads configuration from environment variables.
// Optionally loads a .env file if present (ignored in production).
func Load() (*Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRES_IN_HOURS", 24)
	viper.SetDefault("TENANT_MODE", "jwt")

	hours := viper.GetInt("JWT_EXPIRES_IN_HOURS")

	cfg := &Config{
		App: AppConfig{
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetString("APP_PORT"),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:    viper.GetString("JWT_SECRET"),
			ExpiresIn: time.Duration(hours) * time.Hour,
		},
		WhatsApp: WhatsAppConfig{
			APIURL:        viper.GetString("WHATSAPP_API_URL"),
			PhoneNumberID: viper.GetString("WHATSAPP_PHONE_NUMBER_ID"),
			AccessToken:   viper.GetString("WHATSAPP_ACCESS_TOKEN"),
			VerifyToken:   viper.GetString("WHATSAPP_VERIFY_TOKEN"),
		},
		LLM: LLMConfig{
			Provider: viper.GetString("LLM_PROVIDER"),
			APIKey:   viper.GetString("LLM_API_KEY"),
			Model:    viper.GetString("LLM_MODEL"),
			BaseURL:  viper.GetString("LLM_BASE_URL"),
		},
		TenantMode: viper.GetString("TENANT_MODE"),
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
