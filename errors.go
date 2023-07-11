package deepcopy

import (
	"errors"
)

var (
	ErrTypeInvalid         = errors.New("ErrTypeInvalid")
	ErrTypeNonCopyable     = errors.New("ErrTypeNonCopyable")
	ErrValueInvalid        = errors.New("ErrValueInvalid")
	ErrValueUnaddressable  = errors.New("ErrValueUnaddressable")
	ErrFieldRequireCopying = errors.New("ErrFieldRequireCopying")
)
