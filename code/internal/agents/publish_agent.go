package agents

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type PublishAgent struct{}

func NewPublishAgent() *PublishAgent { return &PublishAgent{} }
func (a *PublishAgent) Code() string { return "publish_exec" }
func (a *PublishAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	if req.Payload != nil {
		if v, ok := req.Payload["simulate_error"].(bool); ok && v {
			return types.Response{}, ErrSimulatedPublishFailure
		}
	}
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":          req.Scene,
			"input":          req.Input,
			"ability":        a.Code(),
			"summary":        "发布执行完成",
			"publish_status": "published",
		},
	}, nil
}
