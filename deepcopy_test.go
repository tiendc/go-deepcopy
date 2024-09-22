package deepcopy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type srcStruct2 struct {
	S   string
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F32 float32
	F64 float64
	B   bool

	Method int
}
type srcStruct1 struct {
	S   string
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F32 float32
	F64 float64
	B   bool

	II []int
	UU []uint
	SS []string
	V  srcStruct2
	VV []srcStruct2

	M1 map[int]string
	M2 *map[int8]int8
	M3 map[int]int
	M4 map[[3]int]*srcStruct2

	P2V *int
	V2P int
	S2A []string

	MatchedX   string `copy:"Matched"`
	UnmatchedX string `copy:"UnmatchedX"`
}

type dstStruct2 struct {
	S   string
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F32 float32
	F64 float64
	B   bool

	MethodVal int
}

func (d *dstStruct2) CopyMethod(v int) error {
	d.MethodVal = v * 2
	return nil
}

type dstStruct1 struct {
	S   StrT
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F32 float32
	F64 float64
	B   *bool

	II []int
	UU []uint
	SS []*StrT
	V  dstStruct2
	VV []dstStruct2

	M1 map[int]string
	M2 map[int8]int8
	M3 *map[int]IntT
	M4 map[[3]int]*dstStruct2

	P2V int
	V2P *int
	S2A [3]string

	MatchedY   string `copy:"Matched"`
	UnmatchedY string `copy:"UnmatchedY"`
}

var (
	srcStructA = srcStruct2{
		S:   "string",
		I:   10,
		I8:  -8,
		I16: -16,
		I32: -32,
		I64: -64,
		U:   10,
		U8:  8,
		U16: 16,
		U32: 32,
		U64: 64,
		F32: 32.32,
		F64: 64.64,
		B:   true,

		Method: 1234,
	}

	srcStructB = srcStruct2{
		S:   "string",
		I:   10,
		I8:  -8,
		I16: -16,
		I32: -32,
		I64: -64,
		U:   10,
		U8:  8,
		U16: 16,
		U32: 32,
		U64: 64,
		F32: 32.32,
		F64: 64.64,
		B:   true,

		Method: 4321,
	}

	srcStruct = srcStruct1{
		S:   "string",
		I:   10,
		I8:  -8,
		I16: -16,
		I32: -32,
		I64: -64,
		U:   10,
		U8:  8,
		U16: 16,
		U32: 32,
		U64: 64,
		F32: 32.32,
		F64: 64.64,
		B:   true,

		II: []int{},
		UU: nil,
		SS: []string{"string1", "string2", "", "string4", "string5"},
		V:  srcStructA,
		VV: []srcStruct2{srcStructA, srcStructB, srcStructA, srcStructB, srcStructA},

		M1: map[int]string{1: "11", 2: "22", 3: "33"},
		M2: nil,
		M3: map[int]int{7: 77, 8: 88, 9: 99},
		//nolint:gofmt
		M4: map[[3]int]*srcStruct2{
			[3]int{1, 1, 1}: &srcStructA,
			[3]int{2, 2, 2}: &srcStructB,
		},

		P2V: nil,
		V2P: 0,
		S2A: nil,

		MatchedX:   "hahaha",
		UnmatchedX: "hihihi",
	}
)

func Test_Copy(t *testing.T) {
	var dst dstStruct1
	err := Copy(&dst, srcStruct)
	assert.Nil(t, err)
	// TODO: need verification here
}

func Test_ClearCache(t *testing.T) {
	ClearCache()
	assert.Equal(t, 0, len(copierCacheMap))
}

func Test_ConfigOption(t *testing.T) {
	ctx := defaultContext()

	CopyBetweenPtrAndValue(false)(ctx)
	assert.Equal(t, false, ctx.CopyBetweenPtrAndValue)
	CopyBetweenPtrAndValue(true)(ctx)
	assert.Equal(t, true, ctx.CopyBetweenPtrAndValue)

	CopyBetweenStructFieldAndMethod(false)(ctx)
	assert.Equal(t, false, ctx.CopyBetweenStructFieldAndMethod)
	CopyBetweenStructFieldAndMethod(true)(ctx)
	assert.Equal(t, true, ctx.CopyBetweenStructFieldAndMethod)

	IgnoreNonCopyableTypes(false)(ctx)
	assert.Equal(t, false, ctx.IgnoreNonCopyableTypes)
	IgnoreNonCopyableTypes(true)(ctx)
	assert.Equal(t, true, ctx.IgnoreNonCopyableTypes)

	UseGlobalCache(false)(ctx)
	assert.Equal(t, false, ctx.UseGlobalCache)
	UseGlobalCache(true)(ctx)
	assert.Equal(t, true, ctx.UseGlobalCache)
}
