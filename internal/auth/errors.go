package auth

import "errors"

var (
	ErrInvalidSignMethod = errors.New("invalid signing method")
	ErrClaimIdFails      = errors.New("claim parsing id fails")
	ErrClaimMissing      = errors.New("claim missing")
	ErrTokenExpired      = errors.New("token expired")
)
