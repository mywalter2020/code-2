package agents

import (
	"context"
	"fmt"

	"juyu-ai-platform/internal/types"
)

type PageAgent struct{}

func NewPageAgent() *PageAgent    { return &PageAgent{} }
func (a *PageAgent) Code() string { return "page_gen" }
func (a *PageAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	biz := types.BuildBusinessContext(req, nil)
	return types.Response{
		Agent:   a.Code(),
		Success: true,
		Data: map[string]any{
			"scene":   req.Scene,
			"input":   req.Input,
			"ability": a.Code(),
			"summary": "预览页生成完成",
			"product": biz.Product,
			"page": map[string]any{
				"title":    fmt.Sprintf("%s 页面预览", biz.Product.Title),
				"sections": []string{"头图", "卖点", "详情", "确认区域"},
			},
		},
	}, nil
}
