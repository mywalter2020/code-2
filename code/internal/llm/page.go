package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"juyu-ai-platform/internal/config"
)

type PageGenConfig struct {
	Provider       string  `json:"provider"`
	Model          string  `json:"model,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"max_tokens,omitempty"`
	SystemPrompt   string  `json:"system_prompt,omitempty"`
	PromptTemplate string  `json:"prompt_template,omitempty"`
	Enabled        bool    `json:"enabled"`
}

type PageGenerator interface {
	ProviderName() string
	Enabled() bool
	Config() PageGenConfig
	GeneratePage(ctx context.Context, title, description, platform string) (map[string]any, error)
}

func LoadPageGenConfig() PageGenConfig {
	temperature := 0.3
	if v := strings.TrimSpace(config.GetEnv("PAGE_GEN_TEMPERATURE", "")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			temperature = f
		}
	}
	maxTokens := 260
	if v := strings.TrimSpace(config.GetEnv("PAGE_GEN_MAX_TOKENS", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxTokens = n
		}
	}
	provider := strings.TrimSpace(config.GetEnv("PAGE_GEN_PROVIDER", config.GetEnv("CONTENT_GEN_PROVIDER", "stub")))
	model := strings.TrimSpace(config.GetEnv("PAGE_GEN_MODEL", config.GetEnv("CONTENT_GEN_MODEL", config.GetEnv("NVIDIA_MODEL", "meta/llama-3.1-405b-instruct"))))
	systemPrompt := config.GetEnv("PAGE_GEN_SYSTEM_PROMPT", "你擅长生成电商商品页面结构，输出 JSON，字段必须稳定。")
	promptTemplate := config.GetEnv("PAGE_GEN_PROMPT_TEMPLATE", "请基于以下商品信息生成一个 JSON 页面草图。字段必须包含 title 和 sections，sections 为字符串数组。不要输出 markdown，不要输出解释。商品标题：{{title}}。商品描述：{{description}}。目标平台：{{platform}}。")
	enabled := false
	switch provider {
	case "nvidia":
		enabled = strings.TrimSpace(config.GetEnv("NVIDIA_URL", "")) != "" && strings.TrimSpace(config.GetEnv("NVIDIA_KEY", "")) != ""
	case "openai_compat", "openai-compatible":
		enabled = strings.TrimSpace(config.GetEnv("OPENAI_COMPAT_URL", "")) != "" && strings.TrimSpace(config.GetEnv("OPENAI_COMPAT_KEY", "")) != "" && strings.TrimSpace(model) != ""
	case "stub", "mock", "":
		enabled = false
	}
	return PageGenConfig{Provider: provider, Model: model, Temperature: temperature, MaxTokens: maxTokens, SystemPrompt: systemPrompt, PromptTemplate: promptTemplate, Enabled: enabled}
}

func NewPageGeneratorFromEnv() PageGenerator {
	cfg := LoadPageGenConfig()
	switch cfg.Provider {
	case "openai_compat", "openai-compatible":
		if client := NewOpenAICompatPageGeneratorFromEnv(); client != nil {
			return client
		}
		return NewStubPageGenerator(cfg)
	case "nvidia":
		if client := NewNVIDIAPageGeneratorFromEnv(); client != nil {
			return client
		}
		return NewStubPageGenerator(cfg)
	default:
		return NewStubPageGenerator(cfg)
	}
}

type StubPageGenerator struct{ cfg PageGenConfig }

func NewStubPageGenerator(cfg PageGenConfig) *StubPageGenerator {
	cfg.Enabled = false
	if cfg.Provider == "" {
		cfg.Provider = "stub"
	}
	return &StubPageGenerator{cfg: cfg}
}

func (s *StubPageGenerator) ProviderName() string  { return "stub" }
func (s *StubPageGenerator) Enabled() bool         { return false }
func (s *StubPageGenerator) Config() PageGenConfig { return s.cfg }
func (s *StubPageGenerator) GeneratePage(ctx context.Context, title, description, platform string) (map[string]any, error) {
	_ = ctx
	return map[string]any{
		"title":    title + " 页面预览",
		"sections": []string{"头图", "卖点", "详情", "确认区域"},
		"hero":     map[string]any{"headline": title, "subheadline": description, "platform": platform},
	}, nil
}

func normalizePageJSON(raw string, fallbackTitle string) map[string]any {
	var page map[string]any
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		return map[string]any{"title": fallbackTitle + " 页面预览", "sections": []string{"头图", "卖点", "详情", "确认区域"}, "raw": raw}
	}
	if _, ok := page["title"].(string); !ok {
		page["title"] = fallbackTitle + " 页面预览"
	}
	if _, ok := page["sections"]; !ok {
		page["sections"] = []string{"头图", "卖点", "详情", "确认区域"}
	}
	return page
}

func pagePrompt(cfg PageGenConfig, title, description, platform string) string {
	prompt := cfg.PromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{title}}", title)
	prompt = strings.ReplaceAll(prompt, "{{description}}", description)
	prompt = strings.ReplaceAll(prompt, "{{platform}}", platform)
	return prompt
}

func pageSummary(page map[string]any) string {
	b, _ := json.Marshal(page)
	return fmt.Sprintf("page_json:%s", string(b))
}
