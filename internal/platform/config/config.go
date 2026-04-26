package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv    string
	HTTPPort  string
	LogLevel  string
	JWTSecret string

	DatabaseURL string
	RedisURL    string

	WhatsAppToken         string
	WhatsAppPhoneNumberID string
	WhatsAppVerifyToken   string
	WhatsAppAPIVersion    string

	LLMProvider     string
	AnthropicAPIKey string
	AnthropicModel  string
	GroqAPIKey      string
	GroqModel       string

	MacDentAPIKey string
	CRMOrigin     string
}

func Load() *Config {
	return &Config{
		AppEnv:    env("APP_ENV", "dev"),
		HTTPPort:  env("HTTP_PORT", "8080"),
		LogLevel:  env("LOG_LEVEL", "info"),
		JWTSecret: env("JWT_SECRET", "dev-secret-change-me"),

		DatabaseURL: env("DATABASE_URL", "postgres://dentdesk:dentdesk@localhost:5432/dentdesk?sslmode=disable"),
		RedisURL:    env("REDIS_URL", "redis://localhost:6379/0"),

		WhatsAppToken:         env("WHATSAPP_TOKEN", ""),
		WhatsAppPhoneNumberID: env("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppVerifyToken:   env("WHATSAPP_VERIFY_TOKEN", "dentdesk-verify"),
		WhatsAppAPIVersion:    env("WHATSAPP_API_VERSION", "v20.0"),

		LLMProvider:     env("LLM_PROVIDER", "anthropic"),
		AnthropicAPIKey: env("ANTHROPIC_API_KEY", ""),
		AnthropicModel:  env("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
		GroqAPIKey:      env("GROQ_API_KEY", ""),
		GroqModel:       env("GROQ_MODEL", "llama-3.3-70b-versatile"),

		MacDentAPIKey: env("MACDENT_API_KEY", ""),
		CRMOrigin:     env("CRM_ORIGIN", "http://localhost:5173"),
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
