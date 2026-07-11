package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	Database struct {
		DSN string
	}
}

func LoadConfig(configDir string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	v.SetEnvPrefix("EXCHANGEAPP")
	v.AutomaticEnv()
	v.BindEnv("app.port", "EXCHANGEAPP_APP_PORT")
	v.BindEnv("database.dsn", "EXCHANGEAPP_DATABASE_DSN")
	v.BindEnv("redis.addr", "EXCHANGEAPP_REDIS_ADDR")
	v.BindEnv("redis.password", "EXCHANGEAPP_REDIS_PASSWORD")
	v.BindEnv("redis.db", "EXCHANGEAPP_REDIS_DB")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
