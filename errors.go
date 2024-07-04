package xplugeth

import (
	"errors"
)

var (
	ErrTypeMismatch = errors.New("types do not match")
	ErrSingletonAlreadySet = errors.New("singleton already set")
)