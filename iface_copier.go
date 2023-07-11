package deepcopy

import (
	"reflect"
)

type fromIfaceCopier struct {
	ctx *Context
}

func (c *fromIfaceCopier) Copy(dst, src reflect.Value) error {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Set(reflect.Zero(dst.Type())) // TODO: Go1.18 has no SetZero
			return nil
		}
	}
	cp, err := buildCopier(c.ctx, dst.Type(), src.Type())
	if err != nil {
		return err
	}
	return cp.Copy(dst, src)
}

type toIfaceCopier struct {
	ctx *Context
}

func (c *toIfaceCopier) Copy(dst, src reflect.Value) error {
	for src.Kind() == reflect.Interface {
		src = src.Elem()
		if !src.IsValid() {
			dst.Set(reflect.Zero(dst.Type())) // TODO: Go1.18 has no SetZero
			return nil
		}
	}

	// As `dst` is interface, we clone the `src` and assign back to the `dst`
	srcType := src.Type()
	cloneSrc := reflect.New(srcType).Elem()
	cp, err := buildCopier(c.ctx, srcType, srcType)
	if err != nil {
		return err
	}
	if err = cp.Copy(cloneSrc, src); err != nil {
		return err
	}
	dst.Set(cloneSrc)
	return nil
}
