package constant

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrUniqueConstraint = errors.New("unique constraint violation")
	ErrConflict         = errors.New("resource conflict")
)
