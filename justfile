default: tidy fmt lint test

cover: tidy fmt
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

fmt:
	gofmt -l -w -s .

lint: tidy fmt
	golangci-lint run --fix

test: tidy fmt
	go test -v -race ./...

tidy:
	go mod tidy
