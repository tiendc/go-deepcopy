package deepcopy

import (
	"errors"
	"reflect"
)

type errorCopier struct {
}

func (c *errorCopier) Copy(dst, src reflect.Value) error {
	return errTest
}

type StrT string
type IntT int

func ptrOf[T any](v T) *T {
	return &v
}

var (
	errTest = errors.New("err test")
)
