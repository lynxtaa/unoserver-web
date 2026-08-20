.PHONY: fmt fmt-check lint lint-fix build test testp run docs

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

test:
	go test -v ./...

testp:
	go run gotest.tools/gotestsum --format pkgname

run:
	go run ./cmd/server

docs:
	go run github.com/swaggo/swag/cmd/swag init --generalInfo cmd/server/main.go --output docs --parseDependency --parseInternal
