package config

import (
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host string `json:"host" mapstructure:"host"`
	Port int    `json:"port" mapstructure:"port"`
}

type RedisConfig struct {
	DSN      string `json:"dsn" mapstructure:"dsn"`
	Password string `json:"password" mapstructure:"password"`
}

type RabbitMQConfig struct {
	DSNList string `json:"dsn_list" mapstructure:"dsn_list"`
}

type Config struct {
	ServerConfig   ServerConfig   `json:"server_config" mapstructure:"server"`
	RedisConfig    RedisConfig    `json:"redis_config" mapstructure:"redis"`
	RabbitMQConfig RabbitMQConfig `json:"rabbit_mq_config" mapstructure:"rabbitmq"`
}

var config = &Config{}

func New() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
