.PHONY: fmt fmt-check lint lint-fix build build-debug test run docs

fmt:
	gofmt -s -w .

fmt-check:
	test -z "$$(gofmt -s -l .)"

lint:
	golangci-lint run


lint-fix:
	golangci-lint run --fix

build:
	go build -trimpath -ldflags="-s -w" -o build/server ./cmd/server

build-debug:
	go build -o build/server ./cmd/server

test:
	go test -v ./...

run:
	go run ./cmd/server

docs:
	swag init --generalInfo cmd/server/main.go --output docs --parseDependency --parseInternal
