package agents

import (
	"context"
	"fmt"

	"juyu-ai-platform/internal/types"
)

type EchoAgent struct {
	code string
}

func NewEchoAgent(code string) *EchoAgent {
	return &EchoAgent{code: code}
}

func (a *EchoAgent) Code() string {
	return a.code
}

func (a *EchoAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return types.Response{
		Agent:   a.code,
		Success: true,
		Data: map[string]any{
			"scene":  req.Scene,
			"input":  req.Input,
			"result": fmt.Sprintf("%s finished task", a.code),
		},
	}, nil
}
