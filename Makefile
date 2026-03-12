GO := /root/.openclaw/workspace/.local/go/bin/go
APP := platform
CONFIG := /root/.openclaw/workspace/configs/agents.yaml

.PHONY: tidy build run run-sqlite test

tidy:
	$(GO) mod tidy -C code

build:
	$(GO) build -C code -o $(APP) ./cmd/platform

run:
	cd code && JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

run-sqlite:
	cd code && JUYU_STORE=sqlite JUYU_SQLITE_PATH=../juyu.db JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

test:
	cd code && $(GO) test ./...
