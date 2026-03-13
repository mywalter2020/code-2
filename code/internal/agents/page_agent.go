package agents

import (
	"context"

	"juyu-ai-platform/internal/llm"
	"juyu-ai-platform/internal/types"
)

type PageAgent struct{ llm llm.PageGenerator }

func NewPageAgent() *PageAgent    { return &PageAgent{llm: llm.NewPageGeneratorFromEnv()} }
func (a *PageAgent) Code() string { return "page_gen" }
func (a *PageAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	biz := types.BuildBusinessContext(req, nil)
	platform := "alibaba"
	if req.Payload != nil {
		if p, ok := req.Payload["platform"].(string); ok && p != "" {
			platform = p
		}
	}
	page := map[string]any{"title": biz.Product.Title + " 页面预览", "sections": []string{"头图", "卖点", "详情", "确认区域"}}
	mode := "stub"
	configView := map[string]any{"provider": "stub", "enabled": false}
	if a.llm != nil {
		generated, err := a.llm.GeneratePage(ctx, biz.Product.Title, biz.Product.Description, platform)
		cfg := a.llm.Config()
		configView = map[string]any{
			"provider":    cfg.Provider,
			"enabled":     a.llm.Enabled(),
			"model":       cfg.Model,
			"temperature": cfg.Temperature,
			"max_tokens":  cfg.MaxTokens,
		}
		if err == nil {
			page = generated
			mode = a.llm.ProviderName()
		} else {
			fallback := llm.NewStubPageGenerator(llm.LoadPageGenConfig())
			page, _ = fallback.GeneratePage(ctx, biz.Product.Title, biz.Product.Description, platform)
			mode = "stub"
			configView["fallback"] = "stub"
			configView["provider_error"] = err.Error()
		}
	}
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":       req.Scene,
			"input":       req.Input,
			"ability":     a.Code(),
			"summary":     "预览页生成完成",
			"product":     biz.Product,
			"page":        page,
			"page_mode":   mode,
			"platform":    platform,
			"page_config": configView,
		},
	}, nil
}
