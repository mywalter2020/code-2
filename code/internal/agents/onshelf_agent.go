package agents

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type OnShelfAgent struct{}

func NewOnShelfAgent() *OnShelfAgent { return &OnShelfAgent{} }
func (a *OnShelfAgent) Code() string { return "onshelf_exec" }
func (a *OnShelfAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":        req.Scene,
			"input":        req.Input,
			"ability":      a.Code(),
			"summary":      "上下架执行完成",
			"shelf_status": "on_shelf",
		},
	}, nil
}
