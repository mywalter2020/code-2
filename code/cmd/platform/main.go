package main

import (
	"context"
	"fmt"
	"log"

	"juyu-ai-platform/internal/agents"
	"juyu-ai-platform/internal/config"
	"juyu-ai-platform/internal/orchestrator"
	"juyu-ai-platform/internal/registry"
	"juyu-ai-platform/internal/router"
	"juyu-ai-platform/internal/types"
)

func main() {
	cfg, err := config.Load("configs/agents.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	rg := registry.New()
	for _, ability := range cfg.AbilityAgents {
		if ability.Enabled {
			rg.Register(agents.NewEchoAgent(ability.Code))
		}
	}

	rt := router.New(cfg.MasterAgents)
	orc := orchestrator.New(rt, rg, cfg.Bindings)

	req := types.Request{
		Scene: "product",
		Input: "在阿里平台生成商品页面并准备上架",
		Payload: map[string]any{
			"platform": "alibaba",
		},
	}

	results, err := orc.Execute(context.Background(), req)
	if err != nil {
		log.Fatalf("execute failed: %v", err)
	}

	for _, item := range results {
		fmt.Printf("agent=%s success=%v data=%v\n", item.Agent, item.Success, item.Data)
	}
}
