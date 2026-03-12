package agents

import (
	"context"
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

type GenericAgent struct {
	code string
	name string
}

func NewGenericAgent(code, name string) *GenericAgent {
	return &GenericAgent{code: code, name: name}
}

func (a *GenericAgent) Code() string { return a.code }

func (a *GenericAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return types.Response{
		Agent:   a.code,
		Success: true,
		Data: map[string]any{
			"scene":   req.Scene,
			"input":   req.Input,
			"ability": a.code,
			"summary": fmt.Sprintf("%s completed", a.name),
			"result":  strings.ToUpper(a.code) + " finished task",
		},
	}, nil
}
