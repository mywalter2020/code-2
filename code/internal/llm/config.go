package llm

import (
	"strconv"
	"strings"

	"juyu-ai-platform/internal/config"
)

type ContentGenConfig struct {
	Provider       string  `json:"provider"`
	Model          string  `json:"model,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"max_tokens,omitempty"`
	SystemPrompt   string  `json:"system_prompt,omitempty"`
	PromptTemplate string  `json:"prompt_template,omitempty"`
	Enabled        bool    `json:"enabled"`
}

func LoadContentGenConfig() ContentGenConfig {
	temperature := 0.4
	if v := strings.TrimSpace(config.GetEnv("CONTENT_GEN_TEMPERATURE", "")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			temperature = f
		}
	}
	maxTokens := 220
	if v := strings.TrimSpace(config.GetEnv("CONTENT_GEN_MAX_TOKENS", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxTokens = n
		}
	}
	provider := strings.TrimSpace(config.GetEnv("CONTENT_GEN_PROVIDER", "nvidia"))
	model := strings.TrimSpace(config.GetEnv("NVIDIA_MODEL", config.GetEnv("NVIDIA_DEFAULT_MODEL", "meta/llama-3.1-405b-instruct")))
	systemPrompt := config.GetEnv("CONTENT_GEN_SYSTEM_PROMPT", "你擅长生成电商商品发布文案，输出准确、简洁、可直接使用。")
	promptTemplate := config.GetEnv("CONTENT_GEN_PROMPT_TEMPLATE", "你是电商运营文案助手。请为以下商品生成一段简洁但可直接用于发布页的中文商品文案，控制在120字内。输出纯文本，不要加标题。商品标题：{{title}}。商品描述：{{description}}。目标平台：{{platform}}。")
	return ContentGenConfig{
		Provider:       provider,
		Model:          model,
		Temperature:    temperature,
		MaxTokens:      maxTokens,
		SystemPrompt:   systemPrompt,
		PromptTemplate: promptTemplate,
		Enabled:        strings.TrimSpace(config.GetEnv("NVIDIA_URL", "")) != "" && strings.TrimSpace(config.GetEnv("NVIDIA_KEY", "")) != "",
	}
}
