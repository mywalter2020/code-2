package server

import (
	"context"
	"fmt"
	"strings"

	"juyu-ai-platform/internal/types"
)

type runtimeService struct {
	rg       runtimeAgentRegistry
	bindings []types.BindingView
}

type runtimeAgentRegistry interface {
	Get(code string) (types.AbilityAgent, error)
}

func newRuntimeService(rg runtimeAgentRegistry, bindings []types.BindingView) *runtimeService {
	return &runtimeService{rg: rg, bindings: bindings}
}

func (s *runtimeService) enrichPRD(ctx context.Context, sess *types.RuntimeSession) map[string]any {
	if s == nil || s.rg == nil || sess == nil {
		return nil
	}
	req := runtimeRequestFromSession(sess)
	agent, err := s.rg.Get("content_gen")
	if err != nil {
		return nil
	}
	resp, err := agent.Run(ctx, req)
	if err != nil {
		return nil
	}
	content, _ := resp.Data["content"].(string)
	summary, _ := resp.Data["summary"].(string)
	productTitle := content
	if productTitle == "" {
		productTitle = summary
	}
	return map[string]any{
		"title":      nonEmptyString(productTitle, "产品学习与预览内容生成"),
		"background": nonEmptyString(summary, "基于现有能力生成更真实的 runtime PRD"),
		"markdown":   fmt.Sprintf("# Runtime PRD\n\n- input: %s\n- generated_by: content_gen\n\n%s", sess.Input.Message, content),
	}
}

func (s *runtimeService) planTodo(sess *types.RuntimeSession) []types.RuntimeTodoItem {
	if s == nil || sess == nil {
		return nil
	}
	binding := s.pickBinding(sess)
	if binding == nil {
		return nil
	}
	items := make([]types.RuntimeTodoItem, 0)
	if len(binding.Workflow) > 0 {
		for _, step := range binding.Workflow {
			itemType := strings.TrimSpace(step.Ability)
			invoke := strings.TrimSpace(step.InvokeBinding)
			if itemType == "" && invoke == "" {
				itemType = "workflow_step"
			}
			itemID := fmt.Sprintf("todo_%d", len(items)+1)
			items = append(items, types.RuntimeTodoItem{
				ID:            itemID,
				Title:         humanizeRuntimeTodoTitle(itemType, invoke),
				Type:          nonEmptyString(itemType, nonEmptyString(invoke, "workflow_step")),
				Status:        types.TodoStatusPending,
				ParallelGroup: "workflow",
			})
		}
	}
	if len(items) == 0 {
		for _, ability := range binding.Abilities {
			itemID := fmt.Sprintf("todo_%d", len(items)+1)
			items = append(items, types.RuntimeTodoItem{
				ID:            itemID,
				Title:         humanizeRuntimeTodoTitle(ability, ""),
				Type:          ability,
				Status:        types.TodoStatusPending,
				ParallelGroup: "binding",
			})
		}
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func (s *runtimeService) buildPreview(ctx context.Context, sess *types.RuntimeSession) map[string]any {
	if s == nil || s.rg == nil || sess == nil {
		return nil
	}
	req := runtimeRequestFromSession(sess)
	preview := map[string]any{}
	if agent, err := s.rg.Get("content_gen"); err == nil {
		if resp, err := agent.Run(ctx, req); err == nil {
			if content, ok := resp.Data["content"].(string); ok {
				preview["hero_subtitle"] = content
			}
		}
	}
	if agent, err := s.rg.Get("page_gen"); err == nil {
		if resp, err := agent.Run(ctx, req); err == nil {
			if page, ok := resp.Data["page"].(map[string]any); ok {
				if title, ok := page["title"].(string); ok {
					preview["hero_title"] = title
				}
				preview["page"] = page
			}
		}
	}
	if len(preview) == 0 {
		return nil
	}
	if _, ok := preview["hero_title"]; !ok {
		preview["hero_title"] = "Runtime Preview"
	}
	return preview
}

func (s *runtimeService) pickBinding(sess *types.RuntimeSession) *types.BindingView {
	scene := inferRuntimeScene(sess)
	if scene != "" {
		for i := range s.bindings {
			if strings.EqualFold(strings.TrimSpace(s.bindings[i].SceneType), scene) {
				return &s.bindings[i]
			}
		}
	}
	for i := range s.bindings {
		if s.bindings[i].MasterAgent == "product_ops" {
			return &s.bindings[i]
		}
	}
	if len(s.bindings) > 0 {
		return &s.bindings[0]
	}
	return nil
}

func runtimeRequestFromSession(sess *types.RuntimeSession) types.Request {
	payload := map[string]any{
		"url":      sess.Input.URL,
		"platform": nonEmptyString(sess.Input.Platform, "alibaba"),
		"type":     sess.Input.Type,
	}
	return types.Request{
		Scene:   inferRuntimeScene(sess),
		Input:   nonEmptyString(sess.Input.Message, sess.Input.URL),
		Payload: payload,
	}
}

func inferRuntimeScene(sess *types.RuntimeSession) string {
	if sess == nil {
		return "product"
	}
	inputType := strings.ToLower(strings.TrimSpace(sess.Input.Type))
	switch {
	case strings.Contains(inputType, "competition"), strings.Contains(inputType, "proposal"), strings.Contains(inputType, "ppt"):
		return "competition"
	case strings.Contains(inputType, "content"), strings.Contains(inputType, "publish"):
		return "content"
	case strings.Contains(inputType, "product"), strings.Contains(inputType, "preview"), strings.Contains(inputType, "title"), strings.Contains(inputType, "carousel"), strings.Contains(inputType, "video"), strings.Contains(inputType, "image"):
		return "product"
	}

	message := strings.ToLower(strings.TrimSpace(sess.Input.Message))
	switch {
	case strings.Contains(message, "比赛"), strings.Contains(message, "方案"), strings.Contains(message, "ppt"), strings.Contains(message, "competition"), strings.Contains(message, "proposal"):
		return "competition"
	case strings.Contains(message, "发布"), strings.Contains(message, "分发"), strings.Contains(message, "种草"), strings.Contains(message, "content"), strings.Contains(message, "publish"):
		return "content"
	}

	if strings.TrimSpace(sess.Input.URL) != "" {
		return "product"
	}
	return "product"
}

func humanizeRuntimeTodoTitle(ability, invoke string) string {
	key := nonEmptyString(ability, invoke)
	switch key {
	case "content_gen":
		return "生成商品内容"
	case "page_gen":
		return "生成预览页面"
	case "review_check":
		return "人工审核检查"
	case "publish_exec":
		return "执行发布"
	case "onshelf_exec":
		return "执行上架"
	default:
		if key == "" {
			return "执行任务"
		}
		return "执行 " + key
	}
}

func nonEmptyString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
