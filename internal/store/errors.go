package store

import "errors"

var (
	ErrConflict     = errors.New("resource conflict")
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)
