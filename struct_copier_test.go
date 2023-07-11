package deepcopy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Copy_struct(t *testing.T) {
	t.Run("#1: copy fields directly", func(t *testing.T) {
		type SS struct {
			I int
			U uint
		}
		type DD struct {
			I int
			U uint
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, U: 2}, d)
	})

	t.Run("#2: copy fields with conversion", func(t *testing.T) {
		type SS struct {
			I int
			F float32
		}
		type DD struct {
			I IntT
			F uint
		}

		var s SS = SS{I: 1, F: 2.2}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, F: 2}, d)
	})

	t.Run("#3: copy fields with lossy conversion (int -> int8)", func(t *testing.T) {
		type SS struct {
			I int
			F float32
			X string
		}
		type DD struct {
			I int8
			F uint
			Y bool
		}

		var s SS = SS{I: 128, F: 2.2}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: -128, F: 2}, d)
	})

	t.Run("#4: copy fields with conversion (ptr -> value)", func(t *testing.T) {
		type SS struct {
			I *int
			F float32
		}
		type DD struct {
			I int
			F uint
		}

		var s SS = SS{I: ptrOf(1), F: 2.2}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, F: 2}, d)
	})

	t.Run("#5: copy fields with conversion (value -> ptr)", func(t *testing.T) {
		type SS struct {
			I int
			F float32
		}
		type DD struct {
			I *int8
			F uint
		}

		var s SS = SS{I: 128, F: 2.2}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: ptrOf(int8(-128)), F: 2}, d)
	})

	t.Run("#6: copy fields with conversion (slice -> array)", func(t *testing.T) {
		type SS struct {
			I []int
			F float32
		}
		type DD struct {
			I [5]int8
			F uint
		}

		var s SS = SS{I: []int{126, 127, 128}, F: 2.2}
		var d DD = DD{I: [5]int8{1, 2, 3, 4, 5}}
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: [5]int8{126, 127, -128, 0, 0}, F: 2}, d)
	})

	t.Run("#7: struct-in-struct", func(t *testing.T) {
		type SS2 struct {
			I int
			X bool
		}
		type SS struct {
			I  int
			SS SS2
		}
		type DD2 struct {
			I int
			Y string
		}
		type DD struct {
			I  int
			DD *DD2 `copy:"SS"`
		}

		var s SS = SS{I: 1, SS: SS2{I: 11, X: true}}
		var d DD
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, DD: &DD2{I: 11}}, d)
	})

	t.Run("#8: with src field is ignored", func(t *testing.T) {
		type SS struct {
			I int `copy:"-"`
			F float32
		}
		type DD struct {
			I int
			F uint
		}

		var s SS = SS{I: 1, F: 2.2}
		var d DD = DD{I: 100}
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 100, F: 2}, d)
	})

	t.Run("#9: with dst field is ignored", func(t *testing.T) {
		type SS struct {
			I int
			F float32
		}
		type DD struct {
			I int `copy:"-"`
			F uint
		}

		var s SS = SS{I: 1, F: 2.2}
		var d DD = DD{I: 100}
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 100, F: 2}, d)
	})
}

func Test_Copy_struct_error(t *testing.T) {
	t.Run("#1: struct -> slice (error)", func(t *testing.T) {
		type SS struct {
			I int
		}
		var s SS = SS{111}
		var d []int
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrTypeNonCopyable)
	})

	t.Run("#2: src required but no equivalent dst field", func(t *testing.T) {
		type SS struct {
			I int `copy:",required"`
			U uint
		}
		type DD struct {
			U uint
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#3: dst required but no equivalent src field", func(t *testing.T) {
		type SS struct {
			I int
		}
		type DD struct {
			U uint `copy:",required"`
		}

		var s SS = SS{I: 1}
		var d DD
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#4: struct -> slice (ignore error)", func(t *testing.T) {
		type SS struct {
			I int
		}
		var s SS = SS{111}
		var d []int
		err := Copy(&d, s, IgnoreNonCopyableTypes(true))
		assert.Nil(t, err)
	})

	t.Run("#5: src ignore non-copyable but set 'required'", func(t *testing.T) {
		type SS struct {
			I int `copy:",required"`
			U uint
		}
		type DD struct {
			I []string
			U uint
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, s, IgnoreNonCopyableTypes(true))
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#6: dst ignore non-copyable but set 'required'", func(t *testing.T) {
		type SS struct {
			I int
			U uint
		}
		type DD struct {
			I []string `copy:",required"`
			U uint
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, s, IgnoreNonCopyableTypes(true))
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#7: non-copyable between src and dst fields", func(t *testing.T) {
		type SS struct {
			I []float32
			U uint
		}
		type DD struct {
			I []string
			U uint
		}

		var s SS = SS{I: []float32{1}, U: 2}
		var d DD
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrTypeNonCopyable)
	})
}

func Test_Copy_struct_unexported(t *testing.T) {
	t.Run("#1: unexported -> unexported", func(t *testing.T) {
		type SS struct {
			I int
			u uint
		}
		type DD struct {
			I int
			u uint `copy:",required"`
		}

		var s SS = SS{I: 1, u: 2}
		var d DD
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, u: 2}, d)
	})

	t.Run("#2: unexported -> exported", func(t *testing.T) {
		type SS struct {
			I int
			u uint
		}
		type DD struct {
			I int
			U uint `copy:"u,required"`
		}

		var s SS = SS{I: 1, u: 2}
		var d DD
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, U: 2}, d)
	})

	t.Run("#3: exported -> unexported", func(t *testing.T) {
		type SS struct {
			I int
			U uint
		}
		type DD struct {
			I int
			u uint `copy:"U,required"`
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, DD{I: 1, u: 2}, d)
	})

	t.Run("#4: exported -> unexported with conversion", func(t *testing.T) {
		type SS struct {
			I int
			U uint
		}
		type DD struct {
			i IntT `copy:"I,required"`
			U uint
		}

		var s SS = SS{I: 1, U: 2}
		var d DD
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, DD{i: 1, U: 2}, d)
	})

	t.Run("#5: unexported -> unexported with conversion", func(t *testing.T) {
		type SS struct {
			i int `copy:"I,required"`
			U uint
		}
		type DD struct {
			i *int `copy:"I,required"`
			U uint
		}

		var s SS = SS{i: 1, U: 2}
		var d DD
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, DD{i: ptrOf(1), U: 2}, d)
	})
}

func Test_Copy_struct_unexported_error(t *testing.T) {
	t.Run("#1: src is unaddressable", func(t *testing.T) {
		type SS struct {
			I int
			u uint
		}
		type DD struct {
			I int
			u uint `copy:",required"`
		}

		var s SS = SS{I: 1, u: 2}
		var d DD
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrValueUnaddressable)
	})
}

type testD1 struct {
	x1 int
	x2 int
	x3 int
	x4 uint
	U  uint
}

func (d *testD1) CopyI1(i1 int) error {
	d.x1 = i1 * 2
	return nil
}
func (d *testD1) CopyI2(i2 int) { // incorrect method prototype (no return error)
	d.x2 = i2 * 2
}
func (d *testD1) CopyI3(i3 int, v string) error { // incorrect method prototype (2 input args)
	d.x3 = i3 * 2
	return nil
}
func (d *testD1) CopyI4(i4 uint) error { // incorrect method prototype (unmatched input type)
	d.x4 = i4 * 2
	return nil
}
func (d *testD1) CopyI5(i5 int) string { // incorrect method prototype (not return error type)
	return ""
}
func (d *testD1) CopyI6(i6 int) error { // incorrect method prototype (unmatched input type)
	return errTest
}

func Test_Copy_struct_method(t *testing.T) {
	t.Run("#1: field -> dst method", func(t *testing.T) {
		type SS struct {
			I1 int
			U  uint
		}

		var s SS = SS{I1: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, testD1{x1: 2, U: 2}, d)
	})

	t.Run("#2: unexported field -> dst method", func(t *testing.T) {
		type SS struct {
			i1 int `copy:"I1,required"`
			U  uint
		}

		var s SS = SS{i1: 1, U: 2}
		var d testD1
		err := Copy(&d, &s)
		assert.Nil(t, err)
		assert.Equal(t, testD1{x1: 2, U: 2}, d)
	})

	t.Run("#3: incorrect method prototype (CopyI2())", func(t *testing.T) {
		type SS struct {
			I2 int
			U  uint
		}

		var s SS = SS{I2: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.Nil(t, err)
		assert.Equal(t, testD1{U: 2}, d)
	})

	t.Run("#4: not allow copying from field -> method", func(t *testing.T) {
		type SS struct {
			I1 int
			U  uint
		}

		var s SS = SS{I1: 1, U: 2}
		var d testD1
		err := Copy(&d, s, CopyBetweenStructFieldAndMethod(false))
		assert.Nil(t, err)
		assert.Equal(t, testD1{U: 2}, d)
	})
}

func Test_Copy_struct_method_error(t *testing.T) {
	t.Run("#1: unexported field -> dst method", func(t *testing.T) {
		type SS struct {
			i1 int `copy:"I1,required"`
			U  uint
		}

		var s SS = SS{i1: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrValueUnaddressable)
	})

	t.Run("#2: incorrect method prototype (CopyI2())", func(t *testing.T) {
		type SS struct {
			I2 int `copy:",required"`
			U  uint
		}

		var s SS = SS{I2: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#3: incorrect method prototype (CopyI3())", func(t *testing.T) {
		type SS struct {
			I3 int `copy:",required"`
			U  uint
		}

		var s SS = SS{I3: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#4: incorrect method prototype (CopyI4())", func(t *testing.T) {
		type SS struct {
			I4 int `copy:",required"`
			U  uint
		}

		var s SS = SS{I4: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#5: incorrect method prototype (CopyI5())", func(t *testing.T) {
		type SS struct {
			I5 int `copy:",required"`
			U  uint
		}

		var s SS = SS{I5: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, ErrFieldRequireCopying)
	})

	t.Run("#6: copy method return error (CopyI6())", func(t *testing.T) {
		type SS struct {
			I6 int `copy:",required"`
			U  uint
		}

		var s SS = SS{I6: 1, U: 2}
		var d testD1
		err := Copy(&d, s)
		assert.ErrorIs(t, err, errTest)
	})
}
