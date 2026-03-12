package main

import (
	"log"
	"net/http"
	"time"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/agents"
	"juyu-ai-platform/internal/config"
	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/server"
	"juyu-ai-platform/internal/store"
	"juyu-ai-platform/internal/types"
)

func main() {
	cfg, err := config.Load(config.ResolveConfigPath())
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	adapterRegistry := adapters.NewRegistry()
	adapterRegistry.Register(adapters.NewAlibabaAdapter())
	adapterRegistry.Register(adapters.NewTaobaoAdapter())
	adapterRegistry.Register(adapters.NewDouyinAdapter())

	credentialStore := adapters.NewCredentialStore()
	credentialStore.Set(types.AdapterCredentials{Platform: "alibaba", Fields: map[string]string{"app_key": "demo", "secret": "configured"}})
	credentialStore.Set(types.AdapterCredentials{Platform: "taobao", Fields: map[string]string{"app_key": "demo", "secret": "configured"}})
	credentialStore.Set(types.AdapterCredentials{Platform: "douyin", Fields: map[string]string{"client_id": "demo", "client_secret": "configured"}})

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
			rg.Register(agents.NewPublishAgent(adapterRegistry))
		case "onshelf_exec":
			rg.Register(agents.NewOnShelfAgent(adapterRegistry))
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
		pgDSN := config.GetEnv("JUYU_PG_DSN", store.DefaultPostgresDSN())
		if err := config.WaitForPostgres(pgDSN, 20, 2*time.Second); err != nil {
			log.Fatalf("postgres not ready: %v", err)
		}
		pgStore, err := store.NewPostgresStore(pgDSN)
		if err != nil {
			log.Fatalf("init postgres store failed: %v", err)
		}
		st = pgStore
	default:
		st = store.NewMemoryStore()
	}

	orc := orchestrator.New(rt, rg, st, cfg.Bindings)
	api := server.New(
		orc,
		rg,
		server.BuildAbilityMetadata(cfg.AbilityAgents),
		server.BuildMasterMetadata(cfg.MasterAgents),
		server.BuildBindingViews(cfg.Bindings),
	)
	server.AttachAdapterRuntime(api, adapterRegistry, credentialStore)

	mux := http.NewServeMux()
	api.Register(mux)

	addr := ":8080"
	log.Printf("JuYu AI platform listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
