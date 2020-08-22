default: tidy fmt lint test

fmt:
	gofmt -l -w -s .

lint:
	golangci-lint run --fix

test:
	go test -v ./... -count=1

tidy:
	go mod tidy
