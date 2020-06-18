cwd = $(shell pwd)
SOURCES = $(shell find . -type f -iname "*.go") go.mod go.sum

.PHONY: test coverage showcoverage

test: coverage

coverage: $(SOURCES)
	go test -coverpkg=./... -coverprofile=coverage ./...

showcoverage: coverage
	go tool cover -html=coverage
