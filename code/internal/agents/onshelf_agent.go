package agents

import (
	"context"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/types"
)

type OnShelfAgent struct {
	adapters *adapters.Registry
}

func NewOnShelfAgent(registry *adapters.Registry) *OnShelfAgent {
	return &OnShelfAgent{adapters: registry}
}
func (a *OnShelfAgent) Code() string { return "onshelf_exec" }
func (a *OnShelfAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	platform := "alibaba"
	if req.Payload != nil {
		if p, ok := req.Payload["platform"].(string); ok && p != "" {
			platform = p
		}
	}
	adapter, err := a.adapters.Get(platform)
	if err != nil {
		return types.Response{}, err
	}
	biz := types.BuildBusinessContext(req, nil)
	payload := map[string]any{}
	for k, v := range req.Payload {
		payload[k] = v
	}
	payload["product"] = biz.Product
	payload["publish"] = biz.Publish
	result, err := adapter.OnShelf(ctx, types.AdapterRequest{Platform: platform, Action: "on_shelf", Operator: req.Operator, Payload: payload})
	if err != nil {
		return types.Response{}, err
	}
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":          req.Scene,
			"input":          req.Input,
			"ability":        a.Code(),
			"summary":        "上下架执行完成",
			"shelf_status":   result.Status,
			"platform":       platform,
			"product":        biz.Product,
			"publish":        biz.Publish,
			"adapter_result": result,
		},
	}, nil
}
