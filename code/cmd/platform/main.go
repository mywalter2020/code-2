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
)

func main() {
	cfgPath := config.ResolveConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("config validation failed (%s): %v", cfgPath, err)
	}

	credentialStore := adapters.LoadCredentialsFromEnv()
	dryRun := adapters.DryRunFromEnv()
	storeDriver := config.GetEnv("JUYU_STORE", "memory")
	sqlitePath := config.GetEnv("JUYU_SQLITE_PATH", "juyu.db")
	pgDSN := config.GetEnv("JUYU_PG_DSN", store.DefaultPostgresDSN())
	apiKey := config.GetEnv("JUYU_API_KEY", "")
	operatorTokens := config.GetEnv("JUYU_OPERATOR_TOKENS", "")
	if err := config.ValidateRuntime(storeDriver, sqlitePath, pgDSN, apiKey, dryRun); err != nil {
		log.Fatalf("runtime validation failed: %v", err)
	}

	adapterRegistry := adapters.NewDefaultRegistry(credentialStore, dryRun)

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
	switch storeDriver {
	case "sqlite":
		sqliteStore, err := store.NewSQLiteStore(store.ParseDSN(sqlitePath))
		if err != nil {
			log.Fatalf("init sqlite store failed: %v", err)
		}
		st = sqliteStore
	case "postgres", "pg":
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
		apiKey,
		operatorTokens,
	)
	server.AttachAdapterRuntime(api, adapterRegistry, credentialStore)

	runtimeDriver := config.GetEnv("JUYU_RUNTIME_STORE", "memory")
	switch runtimeDriver {
	case "postgres", "pg":
		runtimePG, err := store.NewRuntimePostgresStore(pgDSN)
		if err != nil {
			log.Fatalf("init runtime postgres store failed: %v", err)
		}
		server.AttachRuntimeStore(api, runtimePG, runtimeDriver)
	default:
		server.AttachRuntimeStore(api, store.NewRuntimeMemoryStore(), runtimeDriver)
	}

	mux := http.NewServeMux()
	api.Register(mux)

	addr := ":8080"
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("JuYu AI platform listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
