package controller

import "github.com/pkg/errors"

var (
	ErrNotFoundChild   = errors.New("not found child")
	ErrTooManyChildren = errors.New("too many children")
)

var (
	ErrRetriable = errors.New("retry")
)
