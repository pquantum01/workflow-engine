package repository

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrForeignKey = errors.New("foreign key violation")
)
