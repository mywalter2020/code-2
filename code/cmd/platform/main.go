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
)

func main() {
	cfg, err := config.Load(config.ResolveConfigPath())
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	rg := registry.New()
	for _, ability := range cfg.AbilityAgents {
		if ability.Enabled {
			rg.Register(agents.NewBaseAgent(ability.Code, ability.Name))
		}
	}

	rt := router.New(cfg.MasterAgents)
	orc := orchestrator.New(rt, rg, cfg.Bindings)
	api := server.New(orc, rg)

	mux := http.NewServeMux()
	api.Register(mux)

	addr := ":8080"
	log.Printf("JuYu AI platform listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
