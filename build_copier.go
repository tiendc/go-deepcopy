package deepcopy

import (
	"fmt"
	"reflect"
	"sync"
)

// cacheKey key data structure of cached copiers
type cacheKey struct {
	dstType reflect.Type
	srcType reflect.Type
	flags   uint8
}

var (
	// copierCacheMap global cache for any parsed type
	copierCacheMap = make(map[cacheKey]copier, 10) //nolint:mnd

	// mu read/write cache lock
	mu sync.RWMutex

	// simpleKindMask mask for checking basic kinds such as int, string, ...
	simpleKindMask = func() uint32 {
		n := uint32(0)
		n |= 1 << reflect.Bool
		n |= 1 << reflect.String
		n |= 1 << reflect.Int
		n |= 1 << reflect.Int8
		n |= 1 << reflect.Int16
		n |= 1 << reflect.Int32
		n |= 1 << reflect.Int64
		n |= 1 << reflect.Uint
		n |= 1 << reflect.Uint8
		n |= 1 << reflect.Uint16
		n |= 1 << reflect.Uint32
		n |= 1 << reflect.Uint64
		n |= 1 << reflect.Float32
		n |= 1 << reflect.Float64
		n |= 1 << reflect.Complex64
		n |= 1 << reflect.Complex128
		n |= 1 << reflect.Uintptr
		n |= 1 << reflect.Func
		return n
	}()
)

const (
	// flagCopyBetweenPtrAndValue indicates copying will be performed between `pointers` and `values`
	flagCopyBetweenPtrAndValue = 1
	// flagCopyBetweenStructFieldAndMethod indicates copying will be performed between `struct fields` and `functions`
	flagCopyBetweenStructFieldAndMethod = 2
	// flagIgnoreNonCopyableTypes indicates copying will skip copying non-copyable types without raising errors
	flagIgnoreNonCopyableTypes = 3
)

// prepare prepares context for copiers
func (ctx *Context) prepare() {
	if ctx.UseGlobalCache {
		ctx.copierCacheMap = copierCacheMap
		ctx.mu = &mu
	} else {
		ctx.copierCacheMap = make(map[cacheKey]copier, 5) //nolint:mnd
		ctx.mu = &sync.RWMutex{}
	}

	// Recalculate the flags
	ctx.flags = 0
	if ctx.CopyBetweenPtrAndValue {
		ctx.flags |= 1 << flagCopyBetweenPtrAndValue
	}
	if ctx.CopyBetweenStructFieldAndMethod {
		ctx.flags |= 1 << flagCopyBetweenStructFieldAndMethod
	}
	if ctx.IgnoreNonCopyableTypes {
		ctx.flags |= 1 << flagIgnoreNonCopyableTypes
	}
}

// createCacheKey creates and returns  key for caching a copier
func (ctx *Context) createCacheKey(dstType, srcType reflect.Type) *cacheKey {
	return &cacheKey{
		dstType: dstType,
		srcType: srcType,
		flags:   ctx.flags,
	}
}

// defaultContext creates a default context
func defaultContext() *Context {
	return &Context{
		CopyBetweenPtrAndValue:          true,
		CopyBetweenStructFieldAndMethod: true,
		UseGlobalCache:                  true,
	}
}

// buildCopier build copier for handling copy from `srcType` to `dstType`
// nolint: gocognit, gocyclo
func buildCopier(ctx *Context, dstType, srcType reflect.Type) (copier, error) {
	dstKind, srcKind := dstType.Kind(), srcType.Kind()
	if dstKind == reflect.Interface {
		return &toIfaceCopier{ctx: ctx}, nil
	}
	if srcKind == reflect.Interface {
		return &fromIfaceCopier{ctx: ctx}, nil
	}

	// nolint: nestif
	if srcKind == reflect.Pointer {
		if dstKind == reflect.Pointer { // ptr -> ptr
			copier := &ptr2PtrCopier{ctx: ctx}
			if err := copier.init(dstType, srcType); err != nil {
				return nil, err
			}
			return copier, nil
		} else { // ptr -> value
			if !ctx.CopyBetweenPtrAndValue {
				return onNonCopyable(ctx, dstType, srcType)
			}
			copier := &ptr2ValueCopier{ctx: ctx}
			if err := copier.init(dstType, srcType); err != nil {
				return nil, err
			}
			return copier, nil
		}
	} else {
		if dstKind == reflect.Pointer { // value -> ptr
			if !ctx.CopyBetweenPtrAndValue {
				return onNonCopyable(ctx, dstType, srcType)
			}
			copier := &value2PtrCopier{ctx: ctx}
			if err := copier.init(dstType, srcType); err != nil {
				return nil, err
			}
			return copier, nil
		}
	}

	// Both are not Pointers
	if srcKind == reflect.Slice || srcKind == reflect.Array {
		if dstKind != reflect.Slice && dstKind != reflect.Array {
			return onNonCopyable(ctx, dstType, srcType)
		}
		copier := &sliceCopier{ctx: ctx}
		if err := copier.init(dstType, srcType); err != nil {
			return nil, err
		}
		return copier, nil
	}

	if srcKind == reflect.Struct {
		if dstKind != reflect.Struct {
			return onNonCopyable(ctx, dstType, srcType)
		}
		copier := &structCopier{ctx: ctx}
		if err := copier.init(dstType, srcType); err != nil {
			return nil, err
		}
		return copier, nil
	}

	if srcKind == reflect.Map {
		if dstKind != reflect.Map {
			return onNonCopyable(ctx, dstType, srcType)
		}
		copier := &mapCopier{ctx: ctx}
		if err := copier.init(dstType, srcType); err != nil {
			return nil, err
		}
		return copier, nil
	}

	// Trivial case
	if simpleKindMask&(1<<srcKind) > 0 {
		if dstType == srcType {
			return &directCopier{}, nil
		}
		if srcType.ConvertibleTo(dstType) {
			return &convCopier{}, nil
		}
	}

	return onNonCopyable(ctx, dstType, srcType)
}

func onNonCopyable(ctx *Context, dstType, srcType reflect.Type) (copier, error) {
	if ctx.IgnoreNonCopyableTypes {
		return &nopCopier{}, nil
	}
	return nil, fmt.Errorf("%w: %v -> %v", ErrTypeNonCopyable, srcType, dstType)
}
