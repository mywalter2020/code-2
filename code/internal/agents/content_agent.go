package agents

import (
	"context"
	"fmt"

	"juyu-ai-platform/internal/llm"
	"juyu-ai-platform/internal/types"
)

type ContentAgent struct {
	llm llm.ContentGenerator
}

func NewContentAgent() *ContentAgent { return &ContentAgent{llm: llm.NewContentGeneratorFromEnv()} }
func (a *ContentAgent) Code() string { return "content_gen" }
func (a *ContentAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	biz := types.BuildBusinessContext(req, nil)
	platform := "alibaba"
	if req.Payload != nil {
		if p, ok := req.Payload["platform"].(string); ok && p != "" {
			platform = p
		}
	}
	content := fmt.Sprintf("基于输入生成内容：%s", biz.Product.Title)
	mode := "stub"
	configView := map[string]any{"provider": "stub", "enabled": false}
	if a.llm != nil {
		generated, err := a.llm.GenerateProductContent(ctx, biz.Product.Title, biz.Product.Description, platform)
		if err != nil {
			return types.Response{}, err
		}
		content = generated
		mode = a.llm.ProviderName()
		cfg := a.llm.Config()
		configView = map[string]any{
			"provider":    cfg.Provider,
			"enabled":     a.llm.Enabled(),
			"model":       cfg.Model,
			"temperature": cfg.Temperature,
			"max_tokens":  cfg.MaxTokens,
		}
	}
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":        req.Scene,
			"input":        req.Input,
			"ability":      a.Code(),
			"summary":      "内容生成完成",
			"content":      content,
			"product":      biz.Product,
			"content_mode": mode,
			"platform":     platform,
			"llm_config":   configView,
		},
	}, nil
}
