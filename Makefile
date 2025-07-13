all: lint test

lint:
	golangci-lint run ./...

test:
	go test -race ./...

test-coverage:
	go test -race -covermode=atomic -coverprofile=coverage.txt ./...
