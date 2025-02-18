package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig     `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // development or production
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiryHours int    `mapstructure:"expiry_hours"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "development")
	viper.SetDefault("database.dsn", "root@tcp(127.0.0.1:3306)/url_shortener?parseTime=true")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiry_hours", 24)

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("URL_SHORTENER")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
} 