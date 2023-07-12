package deepcopy

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	DefaultTagName = "copy"
)

type Context struct {
	// CopyBetweenPtrAndValue allow or not copying between pointers and values (default is `true`)
	CopyBetweenPtrAndValue bool

	// CopyBetweenStructFieldAndMethod allow or not copying between struct fields and methods (default is `true`)
	CopyBetweenStructFieldAndMethod bool

	// IgnoreNonCopyableTypes ignore error ErrTypeNonCopyable (default is `false`)
	IgnoreNonCopyableTypes bool

	// UseGlobalCache if false (default is `true`)
	UseGlobalCache bool

	// copierCacheMap cache to speed up parsing types
	copierCacheMap map[cacheKey]copier
	mu             *sync.RWMutex
	flags          uint8
}

type Option func(ctx *Context)

func CopyBetweenPtrAndValue(flag bool) Option {
	return func(ctx *Context) {
		ctx.CopyBetweenPtrAndValue = flag
	}
}
func CopyBetweenStructFieldAndMethod(flag bool) Option {
	return func(ctx *Context) {
		ctx.CopyBetweenStructFieldAndMethod = flag
	}
}
func IgnoreNonCopyableTypes(flag bool) Option {
	return func(ctx *Context) {
		ctx.IgnoreNonCopyableTypes = flag
	}
}
func UseGlobalCache(flag bool) Option {
	return func(ctx *Context) {
		ctx.UseGlobalCache = flag
	}
}

// Copy performs deep copy from `src` to `dst`.
// `dst` must be a pointer to the output var, `src` can be either value or pointer.
// In case you want to copy unexported struct fields within `src`, `src` must be a pointer.
func Copy(dst, src any, options ...Option) (err error) {
	if src == nil || dst == nil {
		return fmt.Errorf("%w: source and destination must be non-nil", ErrValueInvalid)
	}
	dstVal, srcVal := reflect.ValueOf(dst), reflect.ValueOf(src)
	dstType, srcType := dstVal.Type(), srcVal.Type()
	if dstType.Kind() != reflect.Pointer {
		return fmt.Errorf("%w: destination must be pointer", ErrTypeInvalid)
	}
	dstVal, dstType = dstVal.Elem(), dstType.Elem()
	if !dstVal.IsValid() {
		return fmt.Errorf("%w: destination must be non-nil", ErrValueInvalid)
	}

	ctx := defaultContext()
	for _, opt := range options {
		opt(ctx)
	}
	ctx.prepare()

	cacheKey := ctx.createCacheKey(dstType, srcType)
	if ctx.UseGlobalCache {
		ctx.mu.RLock()
		cp, ok := ctx.copierCacheMap[*cacheKey]
		ctx.mu.RUnlock()
		if ok {
			return cp.Copy(dstVal, srcVal)
		}
	}

	cp, err := buildCopier(ctx, dstType, srcType)
	if err != nil {
		return err
	}
	if ctx.UseGlobalCache {
		ctx.mu.Lock()
		ctx.copierCacheMap[*cacheKey] = cp
		ctx.mu.Unlock()
	}

	return cp.Copy(dstVal, srcVal)
}

// ClearCache clears global cache
func ClearCache() {
	mu.Lock()
	copierCacheMap = map[cacheKey]copier{}
	mu.Unlock()
}
