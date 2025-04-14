package handler

import "errors"

var ErrEmptyToken = errors.New("empty token")

var ErrInvalidAuthHeader = errors.New("invalid auth header")

var ErrInvalidToken = errors.New("invalid token")

var ErrInternalServer = errors.New("internal error")
