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

type NVIDIAClient struct {
	url    string
	key    string
	model  string
	client *http.Client
}

type nvidiaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type nvidiaRequest struct {
	Model       string          `json:"model"`
	Messages    []nvidiaMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type nvidiaResponse struct {
	Choices []struct {
		Message nvidiaMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewNVIDIAClientFromEnv() *NVIDIAClient {
	url := strings.TrimSpace(config.GetEnv("NVIDIA_URL", ""))
	key := strings.TrimSpace(config.GetEnv("NVIDIA_KEY", ""))
	model := strings.TrimSpace(config.GetEnv("NVIDIA_MODEL", config.GetEnv("NVIDIA_DEFAULT_MODEL", "meta/llama-3.1-405b-instruct")))
	if url == "" || key == "" || model == "" {
		return nil
	}
	return &NVIDIAClient{
		url:   url,
		key:   key,
		model: model,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *NVIDIAClient) Enabled() bool {
	return c != nil && c.url != "" && c.key != "" && c.model != ""
}

func (c *NVIDIAClient) GenerateProductContent(ctx context.Context, title, description, platform string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("nvidia client not configured")
	}
	prompt := fmt.Sprintf("你是电商运营文案助手。请为以下商品生成一段简洁但可直接用于发布页的中文商品文案，控制在120字内。输出纯文本，不要加标题。商品标题：%s。商品描述：%s。目标平台：%s。", title, description, platform)
	body := nvidiaRequest{
		Model: c.model,
		Messages: []nvidiaMessage{
			{Role: "system", Content: "你擅长生成电商商品发布文案，输出准确、简洁、可直接使用。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.4,
		MaxTokens:   220,
	}
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
		return "", fmt.Errorf("nvidia api status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var out nvidiaResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", err
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", fmt.Errorf("nvidia api error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("nvidia api returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
