package agents

import (
	"context"
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

type BaseAgent struct {
	code string
	name string
}

func NewBaseAgent(code, name string) *BaseAgent {
	return &BaseAgent{code: code, name: name}
}

func (a *BaseAgent) Code() string {
	return a.code
}

func (a *BaseAgent) Run(ctx context.Context, req types.Request) (types.Response, error) {
	result := map[string]any{
		"scene":   req.Scene,
		"input":   req.Input,
		"ability": a.code,
		"summary": fmt.Sprintf("%s completed", a.name),
	}

	switch a.code {
	case "content_gen":
		result["content"] = fmt.Sprintf("基于输入生成内容：%s", req.Input)
	case "page_gen":
		result["page"] = map[string]any{
			"title":    fmt.Sprintf("%s 页面预览", req.Input),
			"sections": []string{"头图", "卖点", "详情", "确认区域"},
		}
	case "review_check":
		result["review"] = "待人工确认"
	case "publish_exec":
		result["publish_status"] = "published"
	case "onshelf_exec":
		result["shelf_status"] = "on_shelf"
	default:
		result["result"] = strings.ToUpper(a.code) + " finished task"
	}

	return types.Response{
		Agent:   a.code,
		Success: true,
		Data:    result,
	}, nil
}
