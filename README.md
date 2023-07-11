[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# Fast deep-copy library for Go

## Functionalities

- True deep copy
- Very fast (see [benchmarks](#benchmarks) section)
- Ability to copy data between convertible types (for example: copy from `int` to `float`)
- Ability to copy between **pointers** and **values** (for example: copy from `*int` to `int`)
- Ability to copy between struct fields and methods
- Ability to copy between unexported struct fields
- Ability to configure copying behavior

## Installation

```shell
go get github.com/tiendc/go-deepcopy
```

## Usage

TBD

## Benchmarks

### Go-DeepCopy vs ManualCopy vs JinzhuCopier vs Deepcopier

[Benchmark code](https://gist.github.com/tiendc/0a739fd880b9aac5373de95458d54808)

```
BenchmarkCopy/Go-DeepCopy
BenchmarkCopy/Go-DeepCopy-10         	  664150	      1796 ns/op

BenchmarkCopy/ManualCopy
BenchmarkCopy/ManualCopy-10          	 3047484	       391.4 ns/op

BenchmarkCopy/JinzhuCopier
BenchmarkCopy/JinzhuCopier-10        	   64623	     18541 ns/op

BenchmarkCopy/Deepcopier
BenchmarkCopy/Deepcopier-10          	   38239	     31253 ns/op
```

## Contributing

- You are welcome to make pull requests for new functions and bug fixes.

## Authors

- Dao Cong Tien ([tiendc](https://github.com/tiendc))

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
