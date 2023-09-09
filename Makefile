all: lint test

prepare:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

build:
	@go build -v ./...

test:
	@go test -cover  -v ./...

cover:
	@go test -race -coverprofile=cover.out -coverpkg=./... ./...
	@go tool cover -html=cover.out -o cover.html

lint:
	golangci-lint --timeout=5m0s run -v ./...

bench:
	go test -benchmem -count 100 -bench .

mod:
	go mod tidy && go mod vendor
