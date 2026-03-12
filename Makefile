GO := /root/.openclaw/workspace/.local/go/bin/go
APP := platform
CONFIG := /root/.openclaw/workspace/configs/agents.yaml

.PHONY: tidy build run test

tidy:
	$(GO) mod tidy -C code

build:
	$(GO) build -C code -o $(APP) ./cmd/platform

run:
	cd code && JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

test:
	$(GO) test ./code/...
