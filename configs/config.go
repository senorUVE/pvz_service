package config

import (
	"fmt"

	"github.com/senorUVE/pvz_service/internal/auth"
	"github.com/senorUVE/pvz_service/internal/controller"
	"github.com/senorUVE/pvz_service/internal/repository"
	"github.com/spf13/viper"
)

type Config struct {
	AppName       string                   `mapstructure:"app_name"`
	AppPort       string                   `mapstructure:"app_port"`
	AuthConfig    auth.AuthConfig          `mapstructure:"auth_config"`
	DBConfig      repository.DBConfig      `mapstructure:"db_config"`
	ServiceConfig controller.ServiceConfig `mapstructure:"service_config"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("test_config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found")
		} else {
			return config, fmt.Errorf("failed to read config: %w", err)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
