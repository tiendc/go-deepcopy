[![Go Version][gover-img]][gover] [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][rpt-img]][rpt]

# Fast deep-copy library for Go

## Functionalities

- True deep copy
- Very fast (see [benchmarks](#benchmarks) section)
- Ability to copy data between convertible types (for example: copy from `int` to `float`)
- Ability to copy between **pointers** and **values** (for example: copy from `*int` to `int`)
- Ability to copy between struct fields and methods
- Ability to copy between unexported struct fields

## Installation

```shell
go get github.com/tiendc/go-deepcopy
```

## Usage

TBD

## Benchmarks

TBD

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
