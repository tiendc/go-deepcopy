package deepcopy

import (
	"reflect"
)

type sliceCopier struct {
	ctx        *Context
	itemCopier copier
}

func (c *sliceCopier) Copy(dst, src reflect.Value) error {
	srcLen := src.Len()
	if dst.Kind() == reflect.Slice { // Slice/Array -> Slice
		// `src` is nil slice, set `dst` nil
		if src.Kind() == reflect.Slice && src.IsNil() {
			dst.Set(reflect.Zero(dst.Type())) // TODO: Go1.18 has no SetZero
			return nil
		}
		newSlice := reflect.MakeSlice(dst.Type(), srcLen, srcLen)
		for i := 0; i < srcLen; i++ {
			if err := c.itemCopier.Copy(newSlice.Index(i), src.Index(i)); err != nil {
				return err
			}
		}
		dst.Set(newSlice)
		return nil
	}

	// Slice/Array -> Array
	dstLen := dst.Len()
	if dstLen < srcLen {
		srcLen = dstLen
	}
	i := 0
	for ; i < srcLen; i++ {
		if err := c.itemCopier.Copy(dst.Index(i), src.Index(i)); err != nil {
			return err
		}
	}
	for ; i < dstLen; i++ {
		item := dst.Index(i)
		item.Set(reflect.Zero(item.Type())) // TODO: Go1.18 has no SetZero
	}
	return nil
}

func (c *sliceCopier) init(dstType, srcType reflect.Type) (err error) {
	dstType, srcType = dstType.Elem(), srcType.Elem()
	srcKind := srcType.Kind()

	// OPTIMIZATION: buildCopier() can handle this nicely, but it will add another wrapping layer
	if simpleKindMask&(1<<srcKind) > 0 {
		if srcType == dstType {
			c.itemCopier = &sliceItemDirectCopier{}
			return nil
		}
		if srcType.ConvertibleTo(dstType) {
			c.itemCopier = &sliceItemConvCopier{}
			return nil
		}
	}

	// OPTIMIZATION: buildCopier() can handle this nicely, but it will add another wrapping layer
	if srcKind == reflect.Struct {
		c.ctx.mu.RLock()
		cp, ok := c.ctx.copierCacheMap[*c.ctx.createCacheKey(dstType, srcType)]
		c.ctx.mu.RUnlock()
		if ok {
			c.itemCopier = cp
			return nil
		}
	}

	c.itemCopier, err = buildCopier(c.ctx, dstType, srcType)
	return
}

type sliceItemDirectCopier struct {
}

func (c *sliceItemDirectCopier) Copy(dst, src reflect.Value) error {
	dst.Set(src)
	return nil
}

type sliceItemConvCopier struct {
}

func (c *sliceItemConvCopier) Copy(dst, src reflect.Value) error {
	dst.Set(src.Convert(dst.Type()))
	return nil
}
