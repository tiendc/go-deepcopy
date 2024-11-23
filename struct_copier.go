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

// structCopier data structure of copier that copies from a `struct`
type structCopier struct {
	ctx          *Context
	fieldCopiers []copier
}

// Copy implementation of Copy function for struct copier
func (c *structCopier) Copy(dst, src reflect.Value) error {
	for _, cp := range c.fieldCopiers {
		if err := cp.Copy(dst, src); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocognit,gocyclo
func (c *structCopier) init(dstType, srcType reflect.Type) (err error) {
	var dstCopyingMethods map[string]*reflect.Method
	if c.ctx.CopyBetweenStructFieldAndMethod {
		dstCopyingMethods = c.parseCopyingMethods(dstType)
	}

	dstDirectFields, mapDstDirectFields, dstInheritedFields, mapDstInheritedFields := c.parseAllFields(dstType)
	srcDirectFields, mapSrcDirectFields, srcInheritedFields, mapSrcInheritedFields := c.parseAllFields(srcType)
	c.fieldCopiers = make([]copier, 0, len(dstDirectFields)+len(dstInheritedFields))

	for _, key := range append(srcDirectFields, srcInheritedFields...) {
		// Find field details from `src` having the key
		sfDetail := mapSrcDirectFields[key]
		if sfDetail == nil {
			sfDetail = mapSrcInheritedFields[key]
		}
		if sfDetail == nil || sfDetail.ignored || sfDetail.done {
			continue
		}

		// Copying methods have higher priority, so if a method defined in the dst struct, use it
		if dstCopyingMethods != nil {
			methodName := "Copy" + strings.ToUpper(key[:1]) + key[1:]
			dstCpMethod, exists := dstCopyingMethods[methodName]
			if exists && !dstCpMethod.Type.In(1).AssignableTo(sfDetail.field.Type) {
				return fmt.Errorf("%w: struct method '%v.%s' does not accept argument type '%v' from '%v[%s]'",
					ErrMethodInvalid, dstType, dstCpMethod.Name, sfDetail.field.Type, srcType, sfDetail.field.Name)
			}
			if exists {
				c.fieldCopiers = append(c.fieldCopiers, c.createField2MethodCopier(dstCpMethod, sfDetail))
				sfDetail.markDone()
				continue
			}
		}

		// Find field details from `dst` having the key
		dfDetail := mapDstDirectFields[key]
		if dfDetail == nil {
			dfDetail = mapDstInheritedFields[key]
		}
		if dfDetail == nil || dfDetail.ignored || dfDetail.done {
			// Found no corresponding dest field to copy to, raise an error in case this is required
			if sfDetail.required {
				return fmt.Errorf("%w: struct field '%v[%s]' requires copying",
					ErrFieldRequireCopying, srcType, sfDetail.field.Name)
			}
			continue
		}

		copier, err := c.buildCopier(dstType, srcType, dfDetail, sfDetail)
		if err != nil {
			return err
		}
		c.fieldCopiers = append(c.fieldCopiers, copier)
		dfDetail.markDone()
		sfDetail.markDone()
	}

	// Remaining dst fields can't be copied
	for _, dfDetail := range mapDstDirectFields {
		if !dfDetail.done && dfDetail.required {
			return fmt.Errorf("%w: struct field '%v[%s]' requires copying",
				ErrFieldRequireCopying, dstType, dfDetail.field.Name)
		}
	}
	for _, dfDetail := range mapDstInheritedFields {
		if !dfDetail.done && dfDetail.required {
			return fmt.Errorf("%w: struct field '%v[%s]' requires copying",
				ErrFieldRequireCopying, dstType, dfDetail.field.Name)
		}
	}

	return nil
}

// parseCopyingMethods collects all copying methods from the given struct type
func (c *structCopier) parseCopyingMethods(structType reflect.Type) map[string]*reflect.Method {
	ptrType := reflect.PointerTo(structType)
	numMethods := ptrType.NumMethod()
	result := make(map[string]*reflect.Method, numMethods)
	for i := 0; i < numMethods; i++ {
		method := ptrType.Method(i)
		// Method name must be something like `Copy<something>`
		if !strings.HasPrefix(method.Name, "Copy") {
			continue
		}
		// Method must accept an arg and return error type (1st arg is the struct itself)
		if method.Type.NumIn() != 2 || method.Type.NumOut() != 1 {
			continue
		}
		if method.Type.Out(0) != errType {
			continue
		}
		result[method.Name] = &method
	}
	return result
}

// parseAllFields parses all fields of a struct including direct fields and fields inherited from embedded structs
func (c *structCopier) parseAllFields(typ reflect.Type) (
	directFieldKeys []string,
	mapDirectFields map[string]*fieldDetail,
	inheritedFieldKeys []string,
	mapInheritedFields map[string]*fieldDetail,
) {
	numFields := typ.NumField()
	directFieldKeys = make([]string, 0, numFields)
	mapDirectFields = make(map[string]*fieldDetail, numFields)
	inheritedFieldKeys = make([]string, 0, numFields)
	mapInheritedFields = make(map[string]*fieldDetail, numFields)

	for i := 0; i < numFields; i++ {
		sf := typ.Field(i)
		fDetail := &fieldDetail{field: &sf, index: []int{i}}
		parseTag(fDetail)
		if fDetail.ignored {
			continue
		}
		directFieldKeys = append(directFieldKeys, fDetail.key)
		mapDirectFields[fDetail.key] = fDetail

		// Parse embedded struct to get its fields
		if sf.Anonymous {
			for key, detail := range c.parseAllNestedFields(sf.Type, fDetail.index) {
				inheritedFieldKeys = append(inheritedFieldKeys, key)
				mapInheritedFields[key] = detail
				fDetail.nestedFields = append(fDetail.nestedFields, detail)
			}
		}
	}
	return directFieldKeys, mapDirectFields, inheritedFieldKeys, mapInheritedFields
}

// parseAllNestedFields parses all fields with initial index of starting field
func (c *structCopier) parseAllNestedFields(typ reflect.Type, index []int) map[string]*fieldDetail {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	numFields := typ.NumField()
	result := make(map[string]*fieldDetail, numFields)

	for i := 0; i < numFields; i++ {
		sf := typ.Field(i)
		fDetail := &fieldDetail{field: &sf, index: append(index, i)}
		parseTag(fDetail)
		if fDetail.ignored {
			continue
		}
		result[fDetail.key] = fDetail
		// Parse embedded struct recursively to get its fields
		if sf.Anonymous {
			for key, detail := range c.parseAllNestedFields(sf.Type, fDetail.index) {
				result[key] = detail
				fDetail.nestedFields = append(fDetail.nestedFields, detail)
			}
		}
	}
	return result
}

func (c *structCopier) buildCopier(dstType, srcType reflect.Type, dstDetail, srcDetail *fieldDetail) (copier, error) {
	df, sf := dstDetail.field, srcDetail.field

	// OPTIMIZATION: buildCopier() can handle this nicely
	if simpleKindMask&(1<<sf.Type.Kind()) > 0 {
		if sf.Type == df.Type {
			// NOTE: pass nil to unset custom copier and trigger direct copying.
			// We can pass `&directCopier{}` for the same result (but it's a bit slower).
			return c.createField2FieldCopier(dstDetail, srcDetail, nil), nil
		}
		if sf.Type.ConvertibleTo(df.Type) {
			return c.createField2FieldCopier(dstDetail, srcDetail, defaultConvCopier), nil
		}
	}

	cp, err := buildCopier(c.ctx, df.Type, sf.Type)
	if err != nil {
		return nil, err
	}
	if c.ctx.IgnoreNonCopyableTypes && (srcDetail.required || dstDetail.required) {
		_, isNopCopier := cp.(*nopCopier)
		if isNopCopier && dstDetail.required {
			return nil, fmt.Errorf("%w: struct field '%v[%s]' requires copying",
				ErrFieldRequireCopying, dstType, dstDetail.field.Name)
		}
		if isNopCopier && srcDetail.required {
			return nil, fmt.Errorf("%w: struct field '%v[%s]' requires copying",
				ErrFieldRequireCopying, srcType, srcDetail.field.Name)
		}
	}
	return c.createField2FieldCopier(dstDetail, srcDetail, cp), nil
}

func (c *structCopier) createField2MethodCopier(dM *reflect.Method, sfDetail *fieldDetail) copier {
	return &structField2MethodCopier{
		dstMethod:           dM.Index,
		dstMethodUnexported: !dM.IsExported(),
		srcFieldIndex:       sfDetail.index,
		srcFieldUnexported:  !sfDetail.field.IsExported(),
	}
}

func (c *structCopier) createField2FieldCopier(df, sf *fieldDetail, cp copier) copier {
	return &structField2FieldCopier{
		copier:             cp,
		dstFieldIndex:      df.index,
		dstFieldUnexported: !df.field.IsExported(),
		srcFieldIndex:      sf.index,
		srcFieldUnexported: !sf.field.IsExported(),
	}
}

// structFieldDirectCopier data structure of copier that copies from
// a src field to a dst field directly
type structField2FieldCopier struct {
	copier             copier
	dstFieldIndex      []int
	dstFieldUnexported bool
	srcFieldIndex      []int
	srcFieldUnexported bool
}

// Copy implementation of Copy function for struct field copier direct
func (c *structField2FieldCopier) Copy(dst, src reflect.Value) (err error) {
	if len(c.srcFieldIndex) == 1 {
		src = src.Field(c.srcFieldIndex[0])
	} else {
		// NOTE: When a struct pointer is embedded (e.g. type StructX struct { *BaseStruct }),
		// this retrieval can fail if the embedded struct pointer is nil. Just skip copying when fails.
		src, err = src.FieldByIndexErr(c.srcFieldIndex)
		if err != nil {
			// There's no src field to copy from, reset the dst field to zero
			c.setFieldZero(dst, c.dstFieldIndex)
			return nil //nolint:nilerr
		}
	}
	if c.srcFieldUnexported {
		if !src.CanAddr() {
			return fmt.Errorf("%w: accessing unexported field requires it to be addressable",
				ErrValueUnaddressable)
		}
		src = reflect.NewAt(src.Type(), unsafe.Pointer(src.UnsafeAddr())).Elem() //nolint:gosec
	}

	if len(c.dstFieldIndex) == 1 {
		dst = dst.Field(c.dstFieldIndex[0])
	} else {
		// Get dst field with making sure it's settable
		dst = c.getFieldWithInit(dst, c.dstFieldIndex)
	}
	if c.dstFieldUnexported {
		if !dst.CanAddr() {
			return fmt.Errorf("%w: accessing unexported field requires it to be addressable",
				ErrValueUnaddressable)
		}
		dst = reflect.NewAt(dst.Type(), unsafe.Pointer(dst.UnsafeAddr())).Elem() //nolint:gosec
	}

	// Use custom copier if set
	if c.copier != nil {
		return c.copier.Copy(dst, src)
	}
	// Otherwise, just perform simple direct copying
	dst.Set(src)
	return nil
}

// getFieldWithInit gets deep nested field with init value for pointer ones
func (c *structField2FieldCopier) getFieldWithInit(field reflect.Value, index []int) reflect.Value {
	for _, idx := range index {
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		field = field.Field(idx)
	}
	return field
}

// setFieldZero sets zero to a deep nested field
func (c *structField2FieldCopier) setFieldZero(field reflect.Value, index []int) {
	field, err := field.FieldByIndexErr(index)
	if err == nil && field.IsValid() {
		field.Set(reflect.Zero(field.Type())) // NOTE: Go1.18 has no SetZero
	}
}

// structField2MethodCopier data structure of copier that copies between `fields` and `methods`
type structField2MethodCopier struct {
	dstMethod           int
	dstMethodUnexported bool
	srcFieldIndex       []int
	srcFieldUnexported  bool
}

// Copy implementation of Copy function for struct field copier between `fields` and `methods`
func (c *structField2MethodCopier) Copy(dst, src reflect.Value) (err error) {
	if len(c.srcFieldIndex) == 1 {
		src = src.Field(c.srcFieldIndex[0])
	} else {
		// NOTE: When a struct pointer is embedded (e.g. type StructX struct { *BaseStruct }),
		// this retrieval can fail if the embedded struct pointer is nil. Just skip copying when fails.
		src, err = src.FieldByIndexErr(c.srcFieldIndex)
		if err != nil {
			return nil //nolint:nilerr
		}
	}
	if c.srcFieldUnexported {
		if !src.CanAddr() {
			return fmt.Errorf("%w: accessing unexported field requires it to be addressable",
				ErrValueUnaddressable)
		}
		src = reflect.NewAt(src.Type(), unsafe.Pointer(src.UnsafeAddr())).Elem() //nolint:gosec
	}

	dst = dst.Addr().Method(c.dstMethod)
	if c.dstMethodUnexported {
		dst = reflect.NewAt(dst.Type(), unsafe.Pointer(dst.UnsafeAddr())).Elem() //nolint:gosec
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
