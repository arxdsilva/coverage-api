package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidCoverage = errors.New("coverage percent must be between 0 and 100")
)
