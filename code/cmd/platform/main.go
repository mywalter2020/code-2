package main

import (
	"log"
	"net/http"

	"juyu-ai-platform/internal/agents"
	"juyu-ai-platform/internal/config"
	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/server"
	"juyu-ai-platform/internal/store"
)

func main() {
	cfg, err := config.Load(config.ResolveConfigPath())
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	rg := registry.New()
	for _, ability := range cfg.AbilityAgents {
		if !ability.Enabled {
			continue
		}
		switch ability.Code {
		case "content_gen":
			rg.Register(agents.NewContentAgent())
		case "page_gen":
			rg.Register(agents.NewPageAgent())
		case "review_check":
			rg.Register(agents.NewReviewAgent())
		case "publish_exec":
			rg.Register(agents.NewPublishAgent())
		case "onshelf_exec":
			rg.Register(agents.NewOnShelfAgent())
		default:
			rg.Register(agents.NewGenericAgent(ability.Code, ability.Name))
		}
	}

	rt := router.New(cfg.MasterAgents)

	var st store.TaskStore
	driver := config.GetEnv("JUYU_STORE", "memory")
	switch driver {
	case "sqlite":
		sqliteStore, err := store.NewSQLiteStore(store.ParseDSN(config.GetEnv("JUYU_SQLITE_PATH", "juyu.db")))
		if err != nil {
			log.Fatalf("init sqlite store failed: %v", err)
		}
		st = sqliteStore
	case "postgres", "pg":
		pgStore, err := store.NewPostgresStore(config.GetEnv("JUYU_PG_DSN", store.DefaultPostgresDSN()))
		if err != nil {
			log.Fatalf("init postgres store failed: %v", err)
		}
		st = pgStore
	default:
		st = store.NewMemoryStore()
	}

	orc := orchestrator.New(rt, rg, st, cfg.Bindings)
	api := server.New(orc, rg)

	mux := http.NewServeMux()
	api.Register(mux)

	addr := ":8080"
	log.Printf("JuYu AI platform listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
