package deepcopy

import (
	"reflect"
)

type copier interface {
	Copy(dst, src reflect.Value) error
}

type nopCopier struct {
}

func (c *nopCopier) Copy(dst, src reflect.Value) error {
	return nil
}

type value2PtrCopier struct {
	ctx    *Context
	copier copier
}

func (c *value2PtrCopier) Copy(dst, src reflect.Value) error {
	if dst.IsNil() {
		dst.Set(reflect.New(dst.Type().Elem()))
	}
	dst = dst.Elem()
	return c.copier.Copy(dst, src)
}

func (c *value2PtrCopier) init(dstType, srcType reflect.Type) (err error) {
	c.copier, err = buildCopier(c.ctx, dstType.Elem(), srcType)
	return
}

type ptr2ValueCopier struct {
	ctx    *Context
	copier copier
}

func (c *ptr2ValueCopier) Copy(dst, src reflect.Value) error {
	src = src.Elem()
	if !src.IsValid() {
		dst.Set(reflect.Zero(dst.Type())) // TODO: Go1.18 has no SetZero
		return nil
	}
	return c.copier.Copy(dst, src)
}

func (c *ptr2ValueCopier) init(dstType, srcType reflect.Type) (err error) {
	c.copier, err = buildCopier(c.ctx, dstType, srcType.Elem())
	return
}

type ptr2PtrCopier struct {
	ctx    *Context
	copier copier
}

func (c *ptr2PtrCopier) Copy(dst, src reflect.Value) error {
	src = src.Elem()
	if !src.IsValid() {
		dst.Set(reflect.Zero(dst.Type())) // TODO: Go1.18 has no SetZero
		return nil
	}
	if dst.IsNil() {
		dst.Set(reflect.New(dst.Type().Elem()))
	}
	dst = dst.Elem()
	return c.copier.Copy(dst, src)
}

func (c *ptr2PtrCopier) init(dstType, srcType reflect.Type) (err error) {
	c.copier, err = buildCopier(c.ctx, dstType.Elem(), srcType.Elem())
	return
}

type directCopier struct {
}

func (c *directCopier) Copy(dst, src reflect.Value) error {
	dst.Set(src)
	return nil
}

type convCopier struct {
}

func (c *convCopier) Copy(dst, src reflect.Value) error {
	dst.Set(src.Convert(dst.Type()))
	return nil
}
