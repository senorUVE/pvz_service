package repository

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")

	ErrPVZNotFound = errors.New("pvz not found")

	ErrReceptionNotFound = errors.New("reception not found")

	ErrProductNotFound = errors.New("product not found")

	ErrNoActiveReception = errors.New("no active reception found")
)
