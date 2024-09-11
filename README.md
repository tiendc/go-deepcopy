[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# Fast deep-copy library for Go

## Functionalities

- True deep copy
- Very fast (see [benchmarks](#benchmarks) section)
- Ability to copy data between convertible types (for example: copy from `int` to `float`)
- Ability to copy between `pointers` and `values` (for example: copy from `*int` to `int`)
- Ability to copy between struct fields and methods
- Ability to copy between unexported struct fields
- Ability to configure copying behavior

## Installation

```shell
go get github.com/tiendc/go-deepcopy
```

## Usage

- [First example](#first-example)
- [Copy between struct fields with different names](#copy-between-struct-fields-with-different-names)
- [Ignore copying struct fields](#ignore-copying-struct-fields)
- [Copy between struct fields and methods](#copy-between-struct-fields-and-methods)
- [Copy between unexported struct fields](#copy-between-unexported-struct-fields)
- [Configure copying behavior](#configure-copying-behavior)

### First example

[Playground](https://go.dev/play/p/GsgjDl1vxVd)

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
    _ = deepcopy.Copy(&dst, src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:1 U:2 X: V:{B:true}}
    // {I:11 U:22 X: V:{B:false}}
```

### Copy between struct fields with different names

[Playground](https://go.dev/play/p/MY8ReuT2K3Y)

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
    _ = deepcopy.Copy(&dst, src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {Y:1 U:2}
    // {Y:11 U:22}
```

### Ignore copying struct fields

- By default, matching fields will be copied. If you don't want to copy a field, use tag `-`.

  [Playground](https://go.dev/play/p/RtlmWN1AEsy)

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
    _ = deepcopy.Copy(&dst, src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:0 U:2}
    // {I:0 U:22}
```

### Copy between struct fields and methods

  [Playground 1](https://go.dev/play/p/zb2NU32G2mG) /
  [Playground 2](https://go.dev/play/p/C3FpFwzoPFm)

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
    _ = deepcopy.Copy(&dst, src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {x:1 U:2}
    // {x:11 U:22}
```

### Copy between unexported struct fields

- By default, unexported struct fields will be ignored when copy. If you want to copy them, use tag `required`.

  [Playground](https://go.dev/play/p/9fNq7kwM1y8)

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
    _ = deepcopy.Copy(&dst, src)

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {i:1 U:2}
    // {i:11 U:22}
```

### Configure copying behavior

- Not allow to copy between `ptr` type and `value` (default is `allow`)

  [Playground](https://go.dev/play/p/_SGEYYE4N_m)

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
    err := deepcopy.Copy(&dst, src, deepcopy.CopyBetweenPtrAndValue(false))
    fmt.Println("error:", err)

    // Output:
    // error: ErrTypeNonCopyable: int -> *int
```

- Ignore ErrTypeNonCopyable, the process will not return that error, but some copying won't be performed.
  
  [Playground 1](https://go.dev/play/p/u63SwMKNxU5) /
  [Playground 2](https://go.dev/play/p/ZomOQW2PsPP)

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
    _ = deepcopy.Copy(&dst, src, deepcopy.IgnoreNonCopyableTypes(true))

    for _, d := range dst {
        fmt.Printf("%+v\n", d)
    }

    // Output:
    // {I:0 U:2}
    // {I:0 U:22}
```

## Benchmarks

### Go-DeepCopy vs ManualCopy vs JinzhuCopier vs Deepcopier

[Benchmark code](https://gist.github.com/tiendc/0a739fd880b9aac5373de95458d54808)

```
BenchmarkCopy/Go-DeepCopy
BenchmarkCopy/Go-DeepCopy-10         	 1712484	       685.5 ns/op

BenchmarkCopy/ManualCopy
BenchmarkCopy/ManualCopy-10          	27953836	        41.14 ns/op

BenchmarkCopy/JinzhuCopier
BenchmarkCopy/JinzhuCopier-10        	  129792	      9177 ns/op

BenchmarkCopy/Deepcopier
BenchmarkCopy/Deepcopier-10          	   42990	     27988 ns/op
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
