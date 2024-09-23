[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# Fast deep-copy library for Go

## Functionalities

- True deep copy
- Very fast (see [benchmarks](#benchmarks) section)
- Ability to copy almost all Go types (number, string, bool, function, slice, map, struct)
- Ability to copy data between convertible types (for example: copy from `int` to `float`)
- Ability to copy between `pointers` and `values` (for example: copy from `*int` to `int`)
- Ability to copy between struct fields through struct methods
- Ability to copy inherited fields from embedded structs
- Ability to copy unexported struct fields
- Ability to configure copying behavior

## Installation

```shell
go get github.com/tiendc/go-deepcopy
```

## Usage

- [First example](#first-example)
- [Copy between struct fields with different names](#copy-between-struct-fields-with-different-names)
- [Skip copying struct fields](#skip-copying-struct-fields)
- [Copy struct fields via struct methods](#copy-struct-fields-via-struct-methods)
- [Copy inherited fields from embedded structs](#copy-inherited-fields-from-embedded-structs)
- [Copy unexported struct fields](#copy-unexported-struct-fields)
- [Configure copying behavior](#configure-copying-behavior)

### First example

  [Playground](https://go.dev/play/p/CrP_rZlkNzm)

```go
    type SS struct {
        B bool
    }
    type S struct {
        I  int
        U  uint
        St string
        V  SS
    }
    type DD struct {
        B bool
    }
    type D struct {
        I int
        U uint
        X string
        V DD
    }
    src := []S{{I: 1, U: 2, St: "3", V: SS{B: true}}, {I: 11, U: 22, St: "33", V: SS{B: false}}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:1 U:2 X: V:{B:true}}
    // {I:11 U:22 X: V:{B:false}}
```

### Copy between struct fields with different names

  [Playground](https://go.dev/play/p/WchsGRns0O-)

```go
    type S struct {
        X  int    `copy:"Key"` // 'Key' is used to match the fields
        U  uint
        St string
    }
    type D struct {
        Y int     `copy:"Key"`
        U uint
    }
    src := []S{{X: 1, U: 2, St: "3"}, {X: 11, U: 22, St: "33"}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {Y:1 U:2}
    // {Y:11 U:22}
```

### Skip copying struct fields

- By default, matching fields will be copied. If you don't want to copy a field, use tag value `-`.

  [Playground](https://go.dev/play/p/8KPe1Susjp1)

```go
    // S and D both have `I` field, but we don't want to copy it
    // Tag `-` can be used in both struct definitions or just in one
    type S struct {
        I  int
        U  uint
        St string
    }
    type D struct {
        I int `copy:"-"`
        U uint
    }
    src := []S{{I: 1, U: 2, St: "3"}, {I: 11, U: 22, St: "33"}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:0 U:2}
    // {I:0 U:22}
```

### Copy struct fields via struct methods

- **Note**: If a copying method is defined within a struct, it will have higher priority than matching fields.

  [Playground 1](https://go.dev/play/p/rCawGa5AZh3) /
  [Playground 2](https://go.dev/play/p/vDOhHXyUoyD)

```go
type S struct {
    X  int
    U  uint
    St string
}

type D struct {
    x string
    U uint
}

// Copy method should be in form of `Copy<source-field>` (or key) and return `error` type
func (d *D) CopyX(i int) error {
    d.x = fmt.Sprintf("%d", i)
    return nil
}
```
```go
    src := []S{{X: 1, U: 2, St: "3"}, {X: 11, U: 22, St: "33"}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {x:1 U:2}
    // {x:11 U:22}
```

### Copy inherited fields from embedded structs

- This is default behaviour from version 1.0, for lower versions, you can use custom copying function
to achieve the same result.

  [Playground 1](https://go.dev/play/p/Zjj12AMRYXt) /
  [Playground 2](https://go.dev/play/p/cJGLqpPVHXI)

```go
    type SBase struct {
        St string
    }
    // Source struct has an embedded one
    type S struct {
        SBase
        I int
    }
    // but destination struct doesn't
    type D struct {
        I  int
        St string
    }

    src := []S{{I: 1, SBase: SBase{"abc"}}, {I: 11, SBase: SBase{"xyz"}}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:1 St:abc}
    // {I:11 St:xyz}
```

### Copy unexported struct fields

- By default, unexported struct fields will be ignored when copy. If you want to copy them, use tag attribute `required`.

  [Playground](https://go.dev/play/p/HYWFbnafdfr)

```go
    type S struct {
        i  int
        U  uint
        St string
    }
    type D struct {
        i int `copy:",required"`
        U uint
    }
    src := []S{{i: 1, U: 2, St: "3"}, {i: 11, U: 22, St: "33"}}
    var dst []D
    _ = deepcopy.Copy(&dst, &src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {i:1 U:2}
    // {i:11 U:22}
```

### Configure copying behavior

- Not allow to copy between `ptr` type and `value` (default is `allow`)

  [Playground](https://go.dev/play/p/ZYzGaCNwp2i)

```go
    type S struct {
        I  int
        U  uint
    }
    type D struct {
        I *int
        U uint
    }
    src := []S{{I: 1, U: 2}, {I: 11, U: 22}}
    var dst []D
    err := deepcopy.Copy(&dst, &src, deepcopy.CopyBetweenPtrAndValue(false))
    fmt.Println("error:", err)

    // Output:
    // error: ErrTypeNonCopyable: int -> *int
```

- Ignore ErrTypeNonCopyable, the process will not return that kind of error, but some copyings won't be performed.
  
  [Playground 1](https://go.dev/play/p/YPz49D_oiTY) /
  [Playground 2](https://go.dev/play/p/DNrBJUP-rrM)

```go
    type S struct {
        I []int
        U uint
    }
    type D struct {
        I int
        U uint
    }
    src := []S{{I: []int{1, 2, 3}, U: 2}, {I: []int{1, 2, 3}, U: 22}}
    var dst []D
    // The copy will succeed with ignoring copy of field `I`
    _ = deepcopy.Copy(&dst, &src, deepcopy.IgnoreNonCopyableTypes(true))

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:0 U:2}
    // {I:0 U:22}
```

## Benchmarks

### Go-DeepCopy vs ManualCopy vs JinzhuCopier vs Deepcopier

This is the benchmark result of the latest version of the lib.

  [Benchmark code](https://gist.github.com/tiendc/0a739fd880b9aac5373de95458d54808)

```
BenchmarkCopy/Go-DeepCopy
BenchmarkCopy/Go-DeepCopy-10         	 1753189	       686.7 ns/op

BenchmarkCopy/ManualCopy
BenchmarkCopy/ManualCopy-10          	29309067	        40.64 ns/op

BenchmarkCopy/JinzhuCopier
BenchmarkCopy/JinzhuCopier-10        	  135361	      8873 ns/op

BenchmarkCopy/Deepcopier
BenchmarkCopy/Deepcopier-10          	   40412	     31290 ns/op
```

## Contributing

- You are welcome to make pull requests for new functions and bug fixes.

## License

- [MIT License](LICENSE)

[doc-img]: https://pkg.go.dev/badge/github.com/tiendc/go-deepcopy
[doc]: https://pkg.go.dev/github.com/tiendc/go-deepcopy
[gover-img]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[gover]: https://img.shields.io/badge/Go-%3E%3D%201.18-blue
[ci-img]: https://github.com/tiendc/go-deepcopy/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/tiendc/go-deepcopy/actions/workflows/go.yml
[cov-img]: https://codecov.io/gh/tiendc/go-deepcopy/branch/main/graph/badge.svg
[cov]: https://codecov.io/gh/tiendc/go-deepcopy
[rpt-img]: https://goreportcard.com/badge/github.com/tiendc/go-deepcopy
[rpt]: https://goreportcard.com/report/github.com/tiendc/go-deepcopy
