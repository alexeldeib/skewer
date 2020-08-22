default: tidy fmt lint test

cover:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out -o coverage.txt

fmt:
	gofmt -l -w -s .

lint:
	golangci-lint run --fix

test:
	go test -v -race ./...

tidy:
	go mod tidy
