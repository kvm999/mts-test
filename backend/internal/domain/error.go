package domain

import "errors"

var (
	ErrUserValidation = errors.New("user validation error")
	ErrUserNotFound   = errors.New("user not found")

	ErrProductValidation = errors.New("product validation error")
	ErrProductNotFound   = errors.New("product not found")

	ErrOrderValidation = errors.New("order validation error")
	ErrOrderNotFound   = errors.New("order not found")

	ErrInsufficientStock = errors.New("insufficient product stock")
	ErrInvalidQuantity   = errors.New("invalid quantity")
)
