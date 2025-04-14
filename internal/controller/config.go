package controller

type ServiceConfig struct {
	Salt string `mapstructure:"hash_salt"`
	Cost int    `mapstructure:"hash_cost"`
}
