GO ?= go
TEST_DATABASE_URL ?= postgres://projectname:local-development-only@127.0.0.1:5432/projectname?sslmode=disable
.PHONY: build dev test integration check cross smoke smoke-tui
build:
	mkdir -p bin
	$(GO) build -trimpath -o bin/SINet ./cmd/SINet
	$(GO) build -trimpath -o bin/server ./cmd/server
dev:
	docker compose up -d --build
check:
	test -z "$$($(GO) fmt ./...)"
	$(GO) vet ./...
	$(GO) mod verify
test:
	$(GO) test ./...
integration:
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' $(GO) test -race ./... -count=1
cross:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -trimpath -o bin/SINet-linux-amd64 ./cmd/SINet
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -trimpath -o bin/SINet-darwin-arm64 ./cmd/SINet
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -trimpath -o bin/SINet-windows-amd64.exe ./cmd/SINet
smoke: build
	python3 tests/smoke.py

smoke-tui: build
	python3 tests/tui_smoke.py
