package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"juyu-ai-platform/internal/llm"
)

type providerConfigRequest struct {
	ContentGen providerSection `json:"content_gen"`
	PageGen    providerSection `json:"page_gen"`
}

type providerSection struct {
	Provider       string   `json:"provider"`
	Model          string   `json:"model"`
	Temperature    *float64 `json:"temperature,omitempty"`
	MaxTokens      *int     `json:"max_tokens,omitempty"`
	SystemPrompt   *string  `json:"system_prompt,omitempty"`
	PromptTemplate *string  `json:"prompt_template,omitempty"`
}

func providerRuntimeInfo() map[string]any {
	cfg := llm.LoadContentGenConfig()
	pageCfg := llm.LoadPageGenConfig()
	providers := []string{"stub", "nvidia", "openai_compat"}
	return map[string]any{
		"content_gen": map[string]any{
			"provider":        cfg.Provider,
			"enabled":         cfg.Enabled,
			"model":           cfg.Model,
			"temperature":     cfg.Temperature,
			"max_tokens":      cfg.MaxTokens,
			"system_prompt":   cfg.SystemPrompt,
			"prompt_template": cfg.PromptTemplate,
		},
		"page_gen": map[string]any{
			"provider":        pageCfg.Provider,
			"enabled":         pageCfg.Enabled,
			"model":           pageCfg.Model,
			"temperature":     pageCfg.Temperature,
			"max_tokens":      pageCfg.MaxTokens,
			"system_prompt":   pageCfg.SystemPrompt,
			"prompt_template": pageCfg.PromptTemplate,
		},
		"providers": providers,
		"notes": map[string]any{
			"nvidia_requires":        []string{"NVIDIA_URL", "NVIDIA_KEY"},
			"openai_compat_requires": []string{"OPENAI_COMPAT_URL", "OPENAI_COMPAT_KEY", "OPENAI_COMPAT_MODEL or CONTENT_GEN_MODEL"},
		},
	}
}

func runtimeInfo() map[string]any { return providerRuntimeInfo() }

func (s *Server) handleProviderAdmin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeAPI(w, http.StatusOK, true, "", "ok", "", providerRuntimeInfo())
		return
	case http.MethodPut:
		if !s.requireWriteAuth(w, r) {
			return
		}
		var req providerConfigRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
			return
		}
		if err := applyProviderSection("CONTENT_GEN", req.ContentGen); err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
			return
		}
		if err := applyProviderSection("PAGE_GEN", req.PageGen); err != nil {
			writeAPI(w, http.StatusBadRequest, false, statusCodeToErr(http.StatusBadRequest), "", err.Error(), nil)
			return
		}
		writeAPI(w, http.StatusOK, true, "", "updated", "", providerRuntimeInfo())
		return
	default:
		writeAPI(w, http.StatusMethodNotAllowed, false, statusCodeToErr(http.StatusMethodNotAllowed), "", "method not allowed", nil)
	}
}

func applyProviderSection(prefix string, section providerSection) error {
	provider := strings.TrimSpace(section.Provider)
	if provider == "" {
		provider = "stub"
	}
	switch provider {
	case "stub", "mock", "nvidia", "openai_compat", "openai-compatible":
	default:
		return &httpError{status: http.StatusBadRequest, message: "unsupported provider: " + provider}
	}
	setEnv(prefix+"_PROVIDER", provider)
	if strings.TrimSpace(section.Model) != "" {
		setEnv(prefix+"_MODEL", section.Model)
	}
	if section.Temperature != nil {
		setEnv(prefix+"_TEMPERATURE", strconv.FormatFloat(*section.Temperature, 'f', -1, 64))
	}
	if section.MaxTokens != nil {
		setEnv(prefix+"_MAX_TOKENS", strconv.Itoa(*section.MaxTokens))
	}
	if section.SystemPrompt != nil {
		setEnv(prefix+"_SYSTEM_PROMPT", *section.SystemPrompt)
	}
	if section.PromptTemplate != nil {
		setEnv(prefix+"_PROMPT_TEMPLATE", *section.PromptTemplate)
	}
	return nil
}

func setEnv(key, value string) { _ = os.Setenv(key, strings.TrimSpace(value)) }

type httpError struct {
	status  int
	message string
}

func (e *httpError) Error() string { return e.message }
