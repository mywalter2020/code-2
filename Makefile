GO := /root/.openclaw/workspace/.local/go/bin/go
APP := platform
CONFIG := /root/.openclaw/workspace/configs/agents.yaml

.PHONY: tidy build run run-sqlite run-pg test

tidy:
	$(GO) mod tidy -C code

build:
	$(GO) build -C code -o $(APP) ./cmd/platform

run:
	cd code && JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

run-sqlite:
	cd code && JUYU_STORE=sqlite JUYU_SQLITE_PATH=../juyu.db JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

run-pg:
	cd code && JUYU_STORE=postgres JUYU_PG_DSN='host=127.0.0.1 port=5432 user=postgres password=postgres dbname=juyu sslmode=disable timezone=Asia/Shanghai' JUYU_CONFIG=$(CONFIG) $(GO) run ./cmd/platform

test:
	cd code && $(GO) test ./...
