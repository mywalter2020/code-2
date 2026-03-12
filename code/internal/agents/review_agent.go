package agents

import (
	"context"

	"juyu-ai-platform/internal/types"
)

type ReviewAgent struct{}

func NewReviewAgent() *ReviewAgent { return &ReviewAgent{} }
func (a *ReviewAgent) Code() string { return "review_check" }
func (a *ReviewAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":   req.Scene,
			"input":   req.Input,
			"ability": a.Code(),
			"summary": "待人工确认",
			"review":  "待人工确认",
		},
	}, nil
}
