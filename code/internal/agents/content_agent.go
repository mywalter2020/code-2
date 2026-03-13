package agents

import (
	"context"
	"fmt"

	"juyu-ai-platform/internal/types"
)

type ContentAgent struct{}

func NewContentAgent() *ContentAgent { return &ContentAgent{} }
func (a *ContentAgent) Code() string { return "content_gen" }
func (a *ContentAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	biz := types.BuildBusinessContext(req, nil)
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":   req.Scene,
			"input":   req.Input,
			"ability": a.Code(),
			"summary": "内容生成完成",
			"content": fmt.Sprintf("基于输入生成内容：%s", biz.Product.Title),
			"product": biz.Product,
		},
	}, nil
}
