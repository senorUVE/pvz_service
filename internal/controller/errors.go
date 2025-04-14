package controller

import "errors"

var (
	ErrShortPassword      = errors.New("password is too short")
	ErrInvalidCity        = errors.New("invalid city")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidPasswd      = errors.New("invalid password")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidUUID        = errors.New("invalid UUID")
	ErrFutureDate         = errors.New("date cannot be in the future")
	ErrInvalidProductType = errors.New("invalid product type")
	ErrInvalidPage        = errors.New("page must be ≥ 1")
	ErrInvalidLimit       = errors.New("limit must be between 1 and 30")
	ErrInvalidDateRange   = errors.New("endDate must be ≥ startDate")
	ErrWeakPassword       = errors.New("password must be ≥ 8 characters with special chars")
	ErrInvalidRole        = errors.New("invalid role, allowed: moderator, employee")
)
