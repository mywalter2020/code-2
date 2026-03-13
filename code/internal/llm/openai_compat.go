package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"juyu-ai-platform/internal/config"
)

type OpenAICompatClient struct {
	url     string
	key     string
	model   string
	cfg     ContentGenConfig
	pageCfg PageGenConfig
	client  *http.Client
}

type oaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaRequest struct {
	Model       string      `json:"model"`
	Messages    []oaMessage `json:"messages"`
	Temperature float64     `json:"temperature,omitempty"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
}

type oaResponse struct {
	Choices []struct {
		Message oaMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewOpenAICompatClientFromEnv() *OpenAICompatClient {
	url := strings.TrimSpace(config.GetEnv("OPENAI_COMPAT_URL", ""))
	key := strings.TrimSpace(config.GetEnv("OPENAI_COMPAT_KEY", ""))
	model := strings.TrimSpace(config.GetEnv("CONTENT_GEN_MODEL", config.GetEnv("OPENAI_COMPAT_MODEL", "")))
	if url == "" || key == "" || model == "" {
		return nil
	}
	return &OpenAICompatClient{
		url:     url,
		key:     key,
		model:   model,
		cfg:     LoadContentGenConfig(),
		pageCfg: LoadPageGenConfig(),
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *OpenAICompatClient) ProviderName() string { return "openai_compat" }
func (c *OpenAICompatClient) Enabled() bool {
	return c != nil && c.url != "" && c.key != "" && c.model != ""
}
func (c *OpenAICompatClient) Config() ContentGenConfig { return c.cfg }
func (c *OpenAICompatClient) GenerateProductContent(ctx context.Context, title, description, platform string) (string, error) {
	prompt := c.cfg.PromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{title}}", title)
	prompt = strings.ReplaceAll(prompt, "{{description}}", description)
	prompt = strings.ReplaceAll(prompt, "{{platform}}", platform)
	return c.complete(ctx, c.cfg.SystemPrompt, prompt, c.cfg.Temperature, c.cfg.MaxTokens)
}
func (c *OpenAICompatClient) GeneratePage(ctx context.Context, title, description, platform string) (map[string]any, error) {
	raw, err := c.complete(ctx, c.pageCfg.SystemPrompt, pagePrompt(c.pageCfg, title, description, platform), c.pageCfg.Temperature, c.pageCfg.MaxTokens)
	if err != nil {
		return nil, err
	}
	return normalizePageJSON(raw, title), nil
}
func (c *OpenAICompatClient) complete(ctx context.Context, systemPrompt, prompt string, temperature float64, maxTokens int) (string, error) {
	body := oaRequest{Model: c.model, Messages: []oaMessage{{Role: "system", Content: systemPrompt}, {Role: "user", Content: prompt}}, Temperature: temperature, MaxTokens: maxTokens}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai-compatible api status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var out oaResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", fmt.Errorf("openai-compatible api error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai-compatible api returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

type OpenAICompatPageGenerator struct {
	client *OpenAICompatClient
	cfg    PageGenConfig
}

func NewOpenAICompatPageGeneratorFromEnv() *OpenAICompatPageGenerator {
	client := NewOpenAICompatClientFromEnv()
	if client == nil {
		return nil
	}
	cfg := LoadPageGenConfig()
	client.pageCfg = cfg
	return &OpenAICompatPageGenerator{client: client, cfg: cfg}
}
func (g *OpenAICompatPageGenerator) ProviderName() string { return "openai_compat" }
func (g *OpenAICompatPageGenerator) Enabled() bool {
	return g != nil && g.client != nil && g.client.Enabled()
}
func (g *OpenAICompatPageGenerator) Config() PageGenConfig { return g.cfg }
func (g *OpenAICompatPageGenerator) GeneratePage(ctx context.Context, title, description, platform string) (map[string]any, error) {
	return g.client.GeneratePage(ctx, title, description, platform)
}
