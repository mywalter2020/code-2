package agents

import (
	"context"

	"juyu-ai-platform/internal/adapters"
	"juyu-ai-platform/internal/types"
)

type PublishAgent struct {
	adapters *adapters.Registry
}

func NewPublishAgent(registry *adapters.Registry) *PublishAgent {
	return &PublishAgent{adapters: registry}
}
func (a *PublishAgent) Code() string { return "publish_exec" }
func (a *PublishAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	if req.Payload != nil {
		if v, ok := req.Payload["simulate_error"].(bool); ok && v {
			return types.Response{}, ErrSimulatedPublishFailure
		}
	}
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
	result, err := adapter.Publish(ctx, types.AdapterRequest{Platform: platform, Action: "publish", Operator: req.Operator, Payload: payload})
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
			"summary":        "发布执行完成",
			"publish_status": result.Status,
			"platform":       platform,
			"product":        biz.Product,
			"publish":        biz.Publish,
			"adapter_result": result,
		},
	}, nil
}
