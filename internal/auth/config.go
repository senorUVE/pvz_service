package auth

import "time"

const TokenTTL = 12 * time.Hour

type AuthConfig struct {
	SigningKey string `mapstructure:"jwt_signing_key"`
}
