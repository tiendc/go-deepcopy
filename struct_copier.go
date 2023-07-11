package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var (
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

type structCopier struct {
	ctx          *Context
	fieldCopiers []copier
}

func (c *structCopier) Copy(dst, src reflect.Value) error {
	for _, cp := range c.fieldCopiers {
		if err := cp.Copy(dst, src); err != nil {
			return err
		}
	}
	return nil
}

// nolint: gocognit, gocyclo
func (c *structCopier) init(dstType, srcType reflect.Type) (err error) {
	cacheKey := c.ctx.createCacheKey(dstType, srcType)
	c.ctx.mu.RLock()
	cp, ok := c.ctx.copierCacheMap[*cacheKey]
	c.ctx.mu.RUnlock()
	if ok {
		c.fieldCopiers = cp.(*structCopier).fieldCopiers // nolint: forcetypeassert
		return nil
	}

	srcFields := srcType.NumField()
	dstFields := dstType.NumField()
	mapDstField := make(map[string]*fieldDetail, dstFields)
	for i := 0; i < dstFields; i++ {
		df := dstType.Field(i)
		dfDetail := &fieldDetail{field: &df}
		parseTag(dfDetail)
		if dfDetail.ignored {
			continue
		}
		mapDstField[dfDetail.key] = dfDetail
	}

	c.fieldCopiers = make([]copier, 0, dstFields)
	for i := 0; i < srcFields; i++ {
		sf := srcType.Field(i)
		sfDetail := &fieldDetail{field: &sf}
		parseTag(sfDetail)
		if sfDetail.ignored {
			continue
		}

		dfDetail := mapDstField[sfDetail.key]
		if dfDetail == nil {
			// Find method with same name as source field
			dMethod := c.findDstMethod(dstType, sfDetail)
			if dMethod == nil {
				if sfDetail.required {
					return fmt.Errorf("%w: struct field %s[%s] requires copying",
						ErrFieldRequireCopying, srcType.Name(), sfDetail.field.Name)
				}
				continue
			}
			c.fieldCopiers = append(c.fieldCopiers, c.createMethodCopier(dMethod, &sf))
			continue
		}

		// Destination field found and matched
		df := dfDetail.field
		delete(mapDstField, sfDetail.key)

		// OPTIMIZATION: buildCopier() can handle this nicely, but it will add another wrapping layer
		if simpleKindMask&(1<<sf.Type.Kind()) > 0 {
			if sf.Type == df.Type {
				c.fieldCopiers = append(c.fieldCopiers, c.createDirectCopier(df, &sf))
				continue
			}
			if sf.Type.ConvertibleTo(df.Type) {
				c.fieldCopiers = append(c.fieldCopiers, c.createConvCopier(df, &sf))
				continue
			}
		}

		cp, err := buildCopier(c.ctx, df.Type, sf.Type)
		if err != nil {
			return err
		}
		if c.ctx.IgnoreNonCopyableTypes && (sfDetail.required || dfDetail.required) {
			_, isNopCopier := cp.(*nopCopier)
			if isNopCopier && dfDetail.required {
				return fmt.Errorf("%w: struct field %s[%s] requires copying",
					ErrFieldRequireCopying, dstType.Name(), dfDetail.field.Name)
			}
			if isNopCopier && sfDetail.required {
				return fmt.Errorf("%w: struct field %s[%s] requires copying",
					ErrFieldRequireCopying, srcType.Name(), sfDetail.field.Name)
			}
		}
		c.fieldCopiers = append(c.fieldCopiers, c.createCustomCopier(df, &sf, cp))
	}

	for _, dfDetail := range mapDstField {
		if dfDetail.required {
			return fmt.Errorf("%w: struct field %s[%s] requires copying",
				ErrFieldRequireCopying, dstType.Name(), dfDetail.field.Name)
		}
	}

	c.ctx.mu.Lock()
	c.ctx.copierCacheMap[*cacheKey] = c
	c.ctx.mu.Unlock()
	return nil
}

func (c *structCopier) findDstMethod(dstType reflect.Type, sfDetail *fieldDetail) *reflect.Method {
	if !c.ctx.CopyBetweenStructFieldAndMethod {
		return nil
	}
	// Find method with name is 'Copy' + source field
	// (e.g. src field is 'Amount', dst method should be CopyAmount)
	methodName := "Copy" + strings.ToUpper(sfDetail.key[:1]) + sfDetail.key[1:]
	dMethod, found := reflect.PointerTo(dstType).MethodByName(methodName)
	if !found {
		return nil
	}
	if dMethod.Type.NumIn() != 2 || dMethod.Type.NumOut() != 1 {
		return nil
	}
	if !dMethod.Type.In(1).AssignableTo(sfDetail.field.Type) {
		return nil
	}
	if dMethod.Type.Out(0) != errType {
		return nil
	}
	return &dMethod
}

func (c *structCopier) createDirectCopier(df, sf *reflect.StructField) copier {
	if df.IsExported() && sf.IsExported() {
		return &structFieldDirectCopier{
			dstField: df.Index[0],
			srcField: sf.Index[0],
		}
	}
	return &structUnexportedFieldCopier{
		copier:             &directCopier{},
		dstField:           df.Index[0],
		dstFieldUnexported: !df.IsExported(),
		srcField:           sf.Index[0],
		srcFieldUnexported: !sf.IsExported(),
	}
}

func (c *structCopier) createConvCopier(df, sf *reflect.StructField) copier {
	if df.IsExported() && sf.IsExported() {
		return &structFieldConvCopier{
			dstField: df.Index[0],
			srcField: sf.Index[0],
		}
	}
	return &structUnexportedFieldCopier{
		copier:             &convCopier{},
		dstField:           df.Index[0],
		dstFieldUnexported: !df.IsExported(),
		srcField:           sf.Index[0],
		srcFieldUnexported: !sf.IsExported(),
	}
}

func (c *structCopier) createMethodCopier(dM *reflect.Method, sf *reflect.StructField) copier {
	return &structFieldMethodCopier{
		dstMethod:           dM.Index,
		dstMethodUnexported: !dM.IsExported(),
		srcField:            sf.Index[0],
		srcFieldUnexported:  !sf.IsExported(),
	}
}

func (c *structCopier) createCustomCopier(df, sf *reflect.StructField, cp copier) copier {
	if df.IsExported() && sf.IsExported() {
		return &structFieldCopier{
			copier:   cp,
			dstField: df.Index[0],
			srcField: sf.Index[0],
		}
	}
	return &structUnexportedFieldCopier{
		copier:             cp,
		dstField:           df.Index[0],
		dstFieldUnexported: !df.IsExported(),
		srcField:           sf.Index[0],
		srcFieldUnexported: !sf.IsExported(),
	}
}

type structFieldDirectCopier struct {
	dstField int
	srcField int
}

func (c *structFieldDirectCopier) Copy(dst, src reflect.Value) error {
	dst.Field(c.dstField).Set(src.Field(c.srcField))
	return nil
}

type structFieldConvCopier struct {
	dstField int
	srcField int
}

func (c *structFieldConvCopier) Copy(dst, src reflect.Value) error {
	dstVal := dst.Field(c.dstField)
	dstVal.Set(src.Field(c.srcField).Convert(dstVal.Type()))
	return nil
}

type structFieldCopier struct {
	copier   copier
	dstField int
	srcField int
}

func (c *structFieldCopier) Copy(dst, src reflect.Value) error {
	return c.copier.Copy(dst.Field(c.dstField), src.Field(c.srcField))
}

type structFieldMethodCopier struct {
	dstMethod           int
	dstMethodUnexported bool
	srcField            int
	srcFieldUnexported  bool
}

func (c *structFieldMethodCopier) Copy(dst, src reflect.Value) error {
	src = src.Field(c.srcField)
	if c.srcFieldUnexported {
		if !src.CanAddr() {
			return fmt.Errorf("%w: accessing unexported field requires it to be addressable",
				ErrValueUnaddressable)
		}
		src = reflect.NewAt(src.Type(), unsafe.Pointer(src.UnsafeAddr())).Elem()
	}
	dst = dst.Addr().Method(c.dstMethod)
	if c.dstMethodUnexported {
		dst = reflect.NewAt(dst.Type(), unsafe.Pointer(dst.UnsafeAddr())).Elem()
	}
	errVal := dst.Call([]reflect.Value{src})[0]
	if errVal.IsNil() {
		return nil
	}
	err, ok := errVal.Interface().(error)
	if !ok {
		return fmt.Errorf("%w: struct method returned non-error value", ErrTypeInvalid)
	}
	return err
}

type structUnexportedFieldCopier struct {
	copier             copier
	dstField           int
	dstFieldUnexported bool
	srcField           int
	srcFieldUnexported bool
}

func (c *structUnexportedFieldCopier) Copy(dst, src reflect.Value) error {
	src = src.Field(c.srcField)
	if c.srcFieldUnexported {
		if !src.CanAddr() {
			return fmt.Errorf("%w: accessing unexported field requires it to be addressable",
				ErrValueUnaddressable)
		}
		src = reflect.NewAt(src.Type(), unsafe.Pointer(src.UnsafeAddr())).Elem()
	}
	dst = dst.Field(c.dstField)
	if c.dstFieldUnexported {
		dst = reflect.NewAt(dst.Type(), unsafe.Pointer(dst.UnsafeAddr())).Elem()
	}
	return c.copier.Copy(dst, src)
}
