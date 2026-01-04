APP_NAME=tagtastic

.PHONY: help build test lint fmt clean release codename sync-themes release-prep release-bump quality

help:
	@echo "Targets:"
	@echo "  make build   - Build local binary"
	@echo "  make test    - Run tests"
	@echo "  make lint    - Run golangci-lint"
	@echo "  make fmt     - Format code"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make release - Build multi-platform binaries (GoReleaser)"
	@echo "  make codename - Print the next available release codename"
	@echo "  make sync-themes - Sync data/themes.yaml into internal/data/themes.yaml"
	@echo "  make release-prep VERSION=x.y.z - Prepare CHANGELOG, VERSION, and tag"
	@echo "  make release-bump BUMP=patch [PRE=beta] - Prepare next SemVer bump and tag"
	@echo "  make quality - Run gofmt, go vet, and golangci-lint"

build:
	go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w ./
	goimports -w ./

clean:
	rm -rf bin dist coverage.out

release:
	goreleaser release --clean

codename:
	go run ./cmd/tools/next-codename

sync-themes:
	go run ./cmd/tools/sync-themes

release-prep:
	go run ./cmd/tools/release $(VERSION)

release-bump:
	go run ./cmd/tools/release --bump $(BUMP) $(if $(PRE),--pre $(PRE),) $(if $(PRENUM),--pre-num $(PRENUM),)

quality:
	gofmt -w ./
	go vet ./...
	golangci-lint run ./...
